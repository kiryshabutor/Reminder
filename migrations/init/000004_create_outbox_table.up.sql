CREATE TABLE IF NOT EXISTS reminders_outbox (
    id BIGSERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL,
    aggregate_id BIGINT,
    user_id BIGINT NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) DEFAULT 'PENDING',
    retry_count INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    processed_at TIMESTAMPTZ,
    error_message TEXT
);

CREATE INDEX idx_outbox_pending ON reminders_outbox(status, created_at) 
WHERE status = 'PENDING';

CREATE INDEX idx_outbox_cleanup ON reminders_outbox(status, processed_at)
WHERE status = 'SENT';
