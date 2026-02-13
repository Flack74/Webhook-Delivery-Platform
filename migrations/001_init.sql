CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- =========================================================
-- EVENTS
-- =========================================================
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_event_type
ON events (event_type);


-- =========================================================
-- DELIVERIES
-- =========================================================
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    event_id UUID NOT NULL
        REFERENCES events(id)
        ON DELETE CASCADE,

    endpoint_url TEXT NOT NULL,

    status TEXT NOT NULL
        CHECK (
            status IN (
                'pending',
                'in_progress',
                'succeeded',
                'failed'
            )
        ),

    attempts INTEGER NOT NULL DEFAULT 0,

    next_retry_at TIMESTAMPTZ,

    last_error TEXT,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    UNIQUE (event_id, endpoint_url)
);

CREATE INDEX idx_deliveries_pending
ON deliveries (status, next_retry_at);
