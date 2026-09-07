CREATE TABLE edge_alerts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sender VARCHAR(255) NOT NULL,
    ingested_at TIMESTAMPTZ NOT NULL,
    payload BYTEA NOT NULL
);

CREATE INDEX idx_alerts_time ON edge_alerts (ingested_at DESC);

CREATE TABLE metrics_fallback (
    id UUID DEFAULT gen_random_uuid(),
    sender VARCHAR(255) NOT NULL,
    ingested_at TIMESTAMPTZ NOT NULL,
    payload BYTEA NOT NULL,
    processed BOOLEAN DEFAULT FALSE,
    PRIMARY KEY (id, ingested_at)
) PARTITION BY RANGE (ingested_at);
