-- 0002_scheduling.sql
ALTER TABLE monitor_state ADD COLUMN IF NOT EXISTS next_check_at TIMESTAMPTZ NOT NULL DEFAULT NOW();
CREATE INDEX IF NOT EXISTS idx_monitor_state_next_check ON monitor_state (next_check_at);
CREATE INDEX IF NOT EXISTS idx_check_history_checked_at ON check_history (checked_at);
CREATE UNIQUE INDEX IF NOT EXISTS uq_incident_open_per_monitor ON incidents (monitor_id) WHERE resolved_at IS NULL;
ALTER TABLE monitors ALTER COLUMN method SET DEFAULT 'HEAD';
DO $$ BEGIN
    ALTER TABLE monitors ADD CONSTRAINT chk_interval CHECK (interval_seconds BETWEEN 60 AND 3600);
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
DO $$ BEGIN
    ALTER TABLE monitors ADD CONSTRAINT chk_expected_status CHECK (expected_status BETWEEN 100 AND 599);
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;
