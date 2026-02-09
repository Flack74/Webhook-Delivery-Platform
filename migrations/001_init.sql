-- =========================
-- Initial schema (Phase 3)
-- =========================

CREATE TABLE events (
    id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_event_type
ON events (event_type);
