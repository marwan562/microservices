-- Create flows table
CREATE TABLE flows (
    id TEXT PRIMARY KEY,
    org_id TEXT NOT NULL,
    zone_id TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    enabled BOOLEAN DEFAULT TRUE,
    nodes JSONB NOT NULL,
    edges JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create flow_executions table
CREATE TABLE flow_executions (
    id TEXT PRIMARY KEY,
    flow_id TEXT NOT NULL REFERENCES flows(id),
    trigger_id TEXT,
    status TEXT NOT NULL,
    input JSONB,
    output JSONB,
    steps JSONB NOT NULL DEFAULT '[]',
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    ended_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_flows_zone_id ON flows(zone_id);
CREATE INDEX idx_flow_executions_flow_id ON flow_executions(flow_id);
