package engine

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

type Scheduler struct {
	db          *pgxpool.Pool
	writer      *kafka.Writer
	tickSecs    int
	lookaheadSec int
}

func NewScheduler(db *pgxpool.Pool, writer *kafka.Writer, tickSecs int, lookaheadSec int) *Scheduler {
	if tickSecs <= 0 {
		tickSecs = 30
	}
	if lookaheadSec <= tickSecs {
		lookaheadSec = 120
	}
	return &Scheduler{
		db:           db,
		writer:       writer,
		tickSecs:     tickSecs,
		lookaheadSec: lookaheadSec,
	}
}

type dueMonitor struct {
	ID              string
	URL             string
	ExpectedStatus  int
	IntervalSeconds int
	NextCheckAt     time.Time
}

func (s *Scheduler) Start(ctx context.Context) {
	log.Printf("[Scheduler] Started with tick=%ds, lookahead=%ds", s.tickSecs, s.lookaheadSec)
	ticker := time.NewTicker(time.Duration(s.tickSecs) * time.Second)
	defer ticker.Stop()

	// Run an immediate check on startup
	s.processDueMonitors(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("[Scheduler] Stopped")
			return
		case <-ticker.C:
			s.processDueMonitors(ctx)
		}
	}
}

func (s *Scheduler) processDueMonitors(ctx context.Context) {
	for {
		count, err := s.processBatch(ctx)
		if err != nil {
			log.Printf("[Scheduler] Batch error: %v", err)
			break
		}
		// If fewer than 1000 rows were processed, we have caught up
		if count < 1000 {
			break
		}
	}
}

func (s *Scheduler) processBatch(ctx context.Context) (int, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		SELECT m.id::text, m.url, m.expected_status, m.interval_seconds, s.next_check_at
		FROM monitor_state s
		JOIN monitors m ON m.id = s.monitor_id
		WHERE m.enabled
		  AND s.next_check_at <= NOW() + make_interval(secs => $1)
		ORDER BY s.next_check_at
		LIMIT 1000
		FOR UPDATE OF s SKIP LOCKED;
	`

	rows, err := tx.Query(ctx, query, s.lookaheadSec)
	if err != nil {
		return 0, fmt.Errorf("querying due monitors failed: %w", err)
	}
	defer rows.Close()

	var batch []dueMonitor
	for rows.Next() {
		var m dueMonitor
		if err := rows.Scan(&m.ID, &m.URL, &m.ExpectedStatus, &m.IntervalSeconds, &m.NextCheckAt); err != nil {
			return 0, fmt.Errorf("scanning due monitor failed: %w", err)
		}
		batch = append(batch, m)
	}
	rows.Close()

	if len(batch) == 0 {
		return 0, nil
	}

	// Prepare Kafka messages
	messages := make([]kafka.Message, len(batch))
	now := time.Now().UTC()

	type advanceInfo struct {
		id         string
		newCheckAt time.Time
	}
	advances := make([]advanceInfo, len(batch))

	for i, m := range batch {
		payload, err := json.Marshal(CheckMessage{
			MonitorID:      m.ID,
			URL:            m.URL,
			ExpectedStatus: m.ExpectedStatus,
			ScheduledFor:   m.NextCheckAt,
		})
		if err != nil {
			return 0, fmt.Errorf("marshaling message failed: %w", err)
		}

		messages[i] = kafka.Message{
			Key:   []byte(m.ID),
			Value: payload,
			Time:  m.NextCheckAt,
		}

		// Calculate next due time: scheduled_for + interval_seconds
		// If in the past, use NOW() + interval_seconds
		newDue := m.NextCheckAt.Add(time.Duration(m.IntervalSeconds) * time.Second)
		if newDue.Before(now) {
			newDue = now.Add(time.Duration(m.IntervalSeconds) * time.Second)
		}
		advances[i] = advanceInfo{
			id:         m.ID,
			newCheckAt: newDue,
		}
	}

	// Publish to Kafka (waits for broker acknowledgment)
	if err := s.writer.WriteMessages(ctx, messages...); err != nil {
		return 0, fmt.Errorf("writing to kafka failed: %w", err)
	}

	// Batch update next_check_at
	batchUpdate := &pgx.Batch{}
	for _, adv := range advances {
		batchUpdate.Queue(`
			UPDATE monitor_state
			SET next_check_at = $1, updated_at = NOW()
			WHERE monitor_id = $2::uuid
		`, adv.newCheckAt, adv.id)
	}

	br := tx.SendBatch(ctx, batchUpdate)
	for range advances {
		_, err := br.Exec()
		if err != nil {
			br.Close()
			return 0, fmt.Errorf("updating monitor_state failed: %w", err)
		}
	}
	if err := br.Close(); err != nil {
		return 0, fmt.Errorf("closing batch failed: %w", err)
	}

	// Commit transaction
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("committing scheduler tx failed: %w", err)
	}

	log.Printf("[Scheduler] Published %d due checks to Kafka and advanced schedules", len(batch))
	return len(batch), nil
}
