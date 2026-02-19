-- Create events table for flow retriggering
CREATE TABLE IF NOT EXISTS events (
    id VARCHAR(64) PRIMARY KEY,
    type VARCHAR(128) NOT NULL,
    zone_id VARCHAR(64) NOT NULL,
    data JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_events_zone_id ON events(zone_id);
CREATE INDEX IF NOT EXISTS idx_events_created_at ON events(created_at);
