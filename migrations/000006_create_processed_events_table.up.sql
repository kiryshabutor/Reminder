CREATE TABLE IF NOT EXISTS analytics.processed_events (
    event_id UUID PRIMARY KEY,
    processed_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_processed_events_processed_at ON analytics.processed_events(processed_at);
