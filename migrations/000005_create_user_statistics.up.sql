CREATE SCHEMA IF NOT EXISTS analytics;

CREATE TABLE IF NOT EXISTS analytics.user_statistics (
    user_id UUID PRIMARY KEY,
    total_reminders_created INT DEFAULT 0,
    total_reminders_completed INT DEFAULT 0,
    total_reminders_deleted INT DEFAULT 0,
    active_reminders INT DEFAULT 0,
    completion_rate DECIMAL(5,2) DEFAULT 0.00,
    first_reminder_at TIMESTAMP,
    last_activity_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_last_activity ON analytics.user_statistics(last_activity_at DESC);
CREATE INDEX IF NOT EXISTS idx_completion_rate ON analytics.user_statistics(completion_rate DESC);
