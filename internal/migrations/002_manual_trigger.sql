CREATE TABLE metrics_fallback_today 
    PARTITION OF metrics_fallback 
    FOR VALUES FROM ('2026-09-02 00:00:00-03') TO ('2026-09-03 00:00:00-03');

CREATE INDEX idx_metrics_pending ON metrics_fallback (ingested_at ASC) WHERE processed = FALSE;