package engine

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/enftaurus/stayup/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/segmentio/kafka-go"
)

const (
	MaxStaleness = 5 * time.Minute
	UserAgent    = "StayUp-Monitor/1.0"
	CheckTimeout = 10 * time.Second
)

type WorkerPool struct {
	db           *pgxpool.Pool
	brokers      []string
	topic        string
	consumerGrp  string
	workerCount  int
	allowPrivate bool
	httpClient   *http.Client
}

func NewWorkerPool(db *pgxpool.Pool, brokers []string, topic string, consumerGrp string, count int, allowPrivate bool) *WorkerPool {
	if count <= 0 {
		count = 5
	}
	if consumerGrp == "" {
		consumerGrp = "stayup-workers"
	}

	dialer := utils.NewSafeDialer(allowPrivate, CheckTimeout)
	transport := &http.Transport{
		DialContext:         dialer.DialContext,
		DisableKeepAlives:   true,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: false},
		MaxIdleConns:        0,
		IdleConnTimeout:     1 * time.Second,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   CheckTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("stopped after 5 redirects")
			}
			req.Header.Set("User-Agent", UserAgent)
			return nil
		},
	}

	return &WorkerPool{
		db:           db,
		brokers:      brokers,
		topic:        topic,
		consumerGrp:  consumerGrp,
		workerCount:  count,
		allowPrivate: allowPrivate,
		httpClient:   client,
	}
}

func (p *WorkerPool) Start(ctx context.Context, wg *sync.WaitGroup) {
	log.Printf("[WorkerPool] Starting %d workers on group=%s, topic=%s", p.workerCount, p.consumerGrp, p.topic)

	for i := 1; i <= p.workerCount; i++ {
		wg.Add(1)
		go p.runWorker(ctx, i, wg)
	}
}

func (p *WorkerPool) runWorker(ctx context.Context, id int, wg *sync.WaitGroup) {
	defer wg.Done()

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        p.brokers,
		GroupID:        p.consumerGrp,
		Topic:          p.topic,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 0, // synchronous offset commit
		StartOffset:    kafka.FirstOffset,
	})
	defer reader.Close()

	log.Printf("[Worker %d] Listening for check messages", id)

	for {
		msg, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				log.Printf("[Worker %d] Shutting down", id)
				return
			}
			log.Printf("[Worker %d] Fetch error: %v", id, err)
			time.Sleep(1 * time.Second)
			continue
		}

		if err := p.processMessage(ctx, id, msg); err != nil {
			log.Printf("[Worker %d] Message processing error: %v", id, err)
		}

		// Commit Kafka offset after processing and DB transaction commit
		if err := reader.CommitMessages(ctx, msg); err != nil {
			if !errors.Is(err, context.Canceled) {
				log.Printf("[Worker %d] Failed to commit offset: %v", id, err)
			}
		}
	}
}

func (p *WorkerPool) processMessage(ctx context.Context, workerID int, msg kafka.Message) error {
	var check CheckMessage
	if err := json.Unmarshal(msg.Value, &check); err != nil {
		return fmt.Errorf("malformed check payload: %w", err)
	}

	now := time.Now().UTC()

	// 1. If scheduled_for is more than 5 minutes in the past, skip it
	if now.Sub(check.ScheduledFor) > MaxStaleness {
		log.Printf("[Worker %d] Dropping stale check for monitor %s (scheduled %s)", workerID, check.MonitorID, check.ScheduledFor.Format(time.RFC3339))
		return nil
	}

	// 2. If scheduled_for is in the future, sleep until then
	if check.ScheduledFor.After(now) {
		waitDuration := check.ScheduledFor.Sub(now)
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(waitDuration):
		}
	}

	// 3. Perform the HEAD check
	checkedAt := time.Now().UTC()
	status, statusCode, responseTimeMs, errMsg := p.executeCheck(ctx, check.URL, check.ExpectedStatus)

	// 4. Save result in a single DB transaction
	return p.saveResult(ctx, check.MonitorID, status, statusCode, responseTimeMs, errMsg, checkedAt)
}

func (p *WorkerPool) executeCheck(ctx context.Context, targetURL string, expectedStatus int) (status string, statusCode *int, responseTimeMs int, errMsg *string) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, targetURL, nil)
	if err != nil {
		s := "DOWN"
		e := fmt.Sprintf("invalid request: %v", err)
		return s, nil, 0, &e
	}
	req.Header.Set("User-Agent", UserAgent)

	start := time.Now()
	resp, err := p.httpClient.Do(req)
	duration := int(time.Since(start).Milliseconds())

	if err != nil {
		s := "DOWN"
		cleanErr := cleanHTTPError(err)
		return s, nil, duration, &cleanErr
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code == expectedStatus {
		return "UP", &code, duration, nil
	}

	s := "DOWN"
	e := fmt.Sprintf("expected %d, got %d", expectedStatus, code)
	return s, &code, duration, &e
}

func cleanHTTPError(err error) string {
	msg := err.Error()
	if strings.Contains(msg, "context deadline exceeded") || strings.Contains(msg, "Client.Timeout") {
		return "timeout after 10s"
	}
	if strings.Contains(msg, "connection refused") {
		return "connection refused"
	}
	if strings.Contains(msg, "no such host") {
		return "DNS failure (no such host)"
	}
	if strings.Contains(msg, "certificate") || strings.Contains(msg, "tls:") {
		return "TLS handshake/certificate error"
	}
	if strings.Contains(msg, "blocked by SSRF guard") {
		return "blocked: internal IP address"
	}
	return msg
}

func (p *WorkerPool) saveResult(
	ctx context.Context,
	monitorID string,
	status string,
	statusCode *int,
	responseTimeMs int,
	errMsg *string,
	checkedAt time.Time,
) error {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock monitor_state row FOR UPDATE
	var prevStatus string
	var lastCheckedAt *time.Time
	var enabled bool
	var consecutiveFailures int
	var consecutiveSuccesses int

	err = tx.QueryRow(ctx, `
		SELECT s.status::text, s.last_checked_at, m.enabled, s.consecutive_failures, s.consecutive_successes
		FROM monitor_state s
		JOIN monitors m ON m.id = s.monitor_id
		WHERE s.monitor_id = $1::uuid
		FOR UPDATE OF s
	`, monitorID).Scan(&prevStatus, &lastCheckedAt, &enabled, &consecutiveFailures, &consecutiveSuccesses)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Monitor was deleted while check was in flight
			return nil
		}
		return fmt.Errorf("querying monitor state failed: %w", err)
	}

	// If PAUSED, discard result
	if !enabled {
		return nil
	}

	// Insert into check_history
	_, err = tx.Exec(ctx, `
		INSERT INTO check_history (monitor_id, status, response_time_ms, status_code, error_message, checked_at)
		VALUES ($1::uuid, $2::monitor_status, $3, $4, $5, $6)
	`, monitorID, status, responseTimeMs, statusCode, errMsg, checkedAt)
	if err != nil {
		return fmt.Errorf("inserting check_history failed: %w", err)
	}

	// Only update monitor_state if checkedAt >= last_checked_at
	if lastCheckedAt == nil || !checkedAt.Before(*lastCheckedAt) {
		newFailures := 0
		newSuccesses := 0
		if status == "UP" {
			newSuccesses = consecutiveSuccesses + 1
			newFailures = 0
		} else {
			newFailures = consecutiveFailures + 1
			newSuccesses = 0
		}

		_, err = tx.Exec(ctx, `
			UPDATE monitor_state
			SET status = $2::monitor_status,
			    last_checked_at = $3,
			    last_response_time_ms = $4,
			    last_status_code = $5,
			    last_error = $6,
			    consecutive_failures = $7,
			    consecutive_successes = $8,
			    updated_at = NOW()
			WHERE monitor_id = $1::uuid
		`, monitorID, status, checkedAt, responseTimeMs, statusCode, errMsg, newFailures, newSuccesses)
		if err != nil {
			return fmt.Errorf("updating monitor_state failed: %w", err)
		}

		// Incidents management
		// A change from UNKNOWN or UP to DOWN inserts an incident with started_at = checkedAt
		if (prevStatus == "UNKNOWN" || prevStatus == "UP") && status == "DOWN" {
			_, err = tx.Exec(ctx, `
				INSERT INTO incidents (monitor_id, started_at, status)
				VALUES ($1::uuid, $2, 'DOWN'::monitor_status)
				ON CONFLICT DO NOTHING
			`, monitorID, checkedAt)
			if err != nil {
				return fmt.Errorf("creating incident failed: %w", err)
			}
		} else if prevStatus == "DOWN" && status == "UP" {
			// A change from DOWN to UP sets resolved_at and duration_seconds on the open incident
			_, err = tx.Exec(ctx, `
				UPDATE incidents
				SET resolved_at = $2,
				    duration_seconds = GREATEST(1, EXTRACT(EPOCH FROM ($2 - started_at))::int),
				    status = 'UP'::monitor_status
				WHERE monitor_id = $1::uuid AND resolved_at IS NULL
			`, monitorID, checkedAt)
			if err != nil {
				return fmt.Errorf("resolving incident failed: %w", err)
			}
		}
	}

	return tx.Commit(ctx)
}

// Unused dummy to satisfy pgtype imports if needed
var _ = pgtype.UUID{}
