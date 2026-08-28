-- =====================================================
-- EXTENSIONS
-- =====================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- =====================================================
-- ENUMS
-- =====================================================

CREATE TYPE http_method AS ENUM (
    'GET',
    'POST',
    'PUT',
    'PATCH',
    'DELETE',
    'HEAD',
    'OPTIONS'
);

CREATE TYPE monitor_status AS ENUM (
    'UNKNOWN',
    'UP',
    'DOWN',
    'PAUSED'
);

CREATE TYPE notification_type AS ENUM (
    'EMAIL',
    'DISCORD',
    'SLACK',
    'WEBHOOK'
);

-- =====================================================
-- USERS
-- =====================================================

C
CREATE INDEX idx_users_github_id ON users(github_id);

-- =====================================================
-- USER SESSIONS
-- =====================================================

CREATE TABLE user_sessions (
                               id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                               user_id UUID NOT NULL
                                   REFERENCES users(id)
                                       ON DELETE CASCADE,

                               refresh_token_hash TEXT NOT NULL,

                               expires_at TIMESTAMPTZ NOT NULL,

                               ip_address INET,
                               user_agent TEXT,

                               created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
                               last_used_at TIMESTAMPTZ
);

CREATE INDEX idx_sessions_user
    ON user_sessions(user_id);

-- =====================================================
-- MONITORS
-- =====================================================

CREATE TABLE monitors (

                          id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                          user_id UUID NOT NULL
                              REFERENCES users(id)
                                  ON DELETE CASCADE,

                          name VARCHAR(255) NOT NULL,

                          url TEXT NOT NULL,

                          method http_method NOT NULL DEFAULT 'GET',

                          expected_status SMALLINT NOT NULL DEFAULT 200,

                          interval_seconds INTEGER NOT NULL DEFAULT 60,

                          timeout_seconds INTEGER NOT NULL DEFAULT 10,

                          retries SMALLINT NOT NULL DEFAULT 3,

                          body TEXT,

                          ssl_verify BOOLEAN NOT NULL DEFAULT TRUE,

                          follow_redirects BOOLEAN NOT NULL DEFAULT TRUE,

                          enabled BOOLEAN NOT NULL DEFAULT TRUE,

                          created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                          updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_monitors_user
    ON monitors(user_id);

-- =====================================================
-- OPTIONAL REQUEST HEADERS
-- =====================================================

CREATE TABLE monitor_headers (

                                 id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                                 monitor_id UUID NOT NULL
                                     REFERENCES monitors(id)
                                         ON DELETE CASCADE,

                                 header_key VARCHAR(255) NOT NULL,

                                 header_value TEXT NOT NULL
);

CREATE INDEX idx_headers_monitor
    ON monitor_headers(monitor_id);

-- =====================================================
-- CURRENT STATUS
-- =====================================================

CREATE TABLE monitor_state (

                               monitor_id UUID PRIMARY KEY
                                   REFERENCES monitors(id)
                                       ON DELETE CASCADE,

                               status monitor_status NOT NULL DEFAULT 'UNKNOWN',

                               last_checked_at TIMESTAMPTZ,

                               last_response_time_ms INTEGER,

                               last_status_code INTEGER,

                               last_error TEXT,

                               consecutive_failures INTEGER NOT NULL DEFAULT 0,

                               consecutive_successes INTEGER NOT NULL DEFAULT 0,

                               updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- =====================================================
-- CHECK HISTORY
-- =====================================================

CREATE TABLE check_history (

                               id BIGSERIAL PRIMARY KEY,

                               monitor_id UUID NOT NULL
                                   REFERENCES monitors(id)
                                       ON DELETE CASCADE,

                               status monitor_status NOT NULL,

                               response_time_ms INTEGER,

                               status_code INTEGER,

                               error_message TEXT,

                               checked_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_check_history_monitor_time
    ON check_history(monitor_id, checked_at DESC);

-- =====================================================
-- INCIDENTS
-- =====================================================

CREATE TABLE incidents (

                           id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                           monitor_id UUID NOT NULL
                               REFERENCES monitors(id)
                                   ON DELETE CASCADE,

                           started_at TIMESTAMPTZ NOT NULL,

                           resolved_at TIMESTAMPTZ,

                           duration_seconds INTEGER,

                           status monitor_status NOT NULL
);

CREATE INDEX idx_incidents_monitor
    ON incidents(monitor_id);

-- =====================================================
-- NOTIFICATION CHANNELS
-- =====================================================

CREATE TABLE notification_channels (

                                       id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

                                       user_id UUID NOT NULL
                                           REFERENCES users(id)
                                               ON DELETE CASCADE,

                                       type notification_type NOT NULL,

                                       name VARCHAR(100) NOT NULL,

                                       destination TEXT NOT NULL,

                                       enabled BOOLEAN NOT NULL DEFAULT TRUE,

                                       created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notification_user
    ON notification_channels(user_id);

-- =====================================================
-- NOTIFICATION LOGS
-- =====================================================

CREATE TABLE notification_logs (

                                   id BIGSERIAL PRIMARY KEY,

                                   incident_id UUID
                                       REFERENCES incidents(id)
                                           ON DELETE CASCADE,

                                   channel_id UUID
                                       REFERENCES notification_channels(id)
                                           ON DELETE CASCADE,

                                   sent_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

                                   success BOOLEAN NOT NULL,

                                   provider_response TEXT
);