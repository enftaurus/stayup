package engine

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type RetentionCleaner struct {
	db *pgxpool.Pool
}

func NewRetentionCleaner(db *pgxpool.Pool) *RetentionCleaner {
	return &RetentionCleaner{db: db}
}

func (r *RetentionCleaner) Start(ctx context.Context) {
	log.Println("[RetentionCleaner] Started hourly retention job (7-day window)")
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	// Initial sweep
	r.cleanOldHistory(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("[RetentionCleaner] Stopped")
			return
		case <-ticker.C:
			r.cleanOldHistory(ctx)
		}
	}
}

func (r *RetentionCleaner) cleanOldHistory(ctx context.Context) {
	cutoff := time.Now().UTC().Add(-7 * 24 * time.Hour)
	totalDeleted := int64(0)

	for {
		// Delete in batches of 10,000 using subquery with ctid to avoid holding a long lock
		cmd, err := r.db.Exec(ctx, `
			DELETE FROM check_history
			WHERE id IN (
				SELECT id FROM check_history
				WHERE checked_at < $1
				LIMIT 10000
			)
		`, cutoff)

		if err != nil {
			log.Printf("[RetentionCleaner] Error during deletion batch: %v", err)
			break
		}

		rowsAffected := cmd.RowsAffected()
		totalDeleted += rowsAffected

		if rowsAffected < 10000 {
			break
		}
		time.Sleep(100 * time.Millisecond) // Yield between batches
	}

	if totalDeleted > 0 {
		log.Printf("[RetentionCleaner] Cleaned %d check_history rows older than 7 days", totalDeleted)
	}
}
