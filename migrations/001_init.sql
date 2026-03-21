CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- APPLICATIONS
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    is_active BOOLEAN DEFAULT TRUE
);

-- ENDPOINTS
CREATE TABLE endpoints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    url TEXT NOT NULL,
    secret TEXT NOT NULL,
    description TEXT,
    is_active BOOLEAN DEFAULT TRUE,
    rate_limit INTEGER DEFAULT 1000,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_endpoints_application
ON endpoints(application_id);

-- EVENTS
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),

    UNIQUE(application_id, idempotency_key)
);

CREATE INDEX idx_events_application
ON events(application_id);

CREATE INDEX idx_events_event_type
ON events(event_type);

CREATE INDEX idx_events_created_at
ON events(created_at);

-- DELIVERIES
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    endpoint_id UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,

    status TEXT NOT NULL DEFAULT 'pending'
        CHECK (
            status IN (
                'pending',
                'in_progress',
                'succeeded',
                'failed',
                'dead_letter'
            )
        ),

    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    last_error TEXT,
    max_attempts INTEGER NOT NULL DEFAULT 5,

    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),

    UNIQUE (event_id, endpoint_id)
);

CREATE INDEX idx_deliveries_status
ON deliveries(status);

CREATE INDEX idx_deliveries_retry
ON deliveries(status, next_retry_at);

CREATE INDEX idx_deliveries_event
ON deliveries(event_id);

CREATE INDEX idx_deliveries_worker_queue
ON deliveries(status, next_retry_at, id);

-- DELIVERY ATTEMPTS
CREATE TABLE delivery_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    delivery_id UUID NOT NULL REFERENCES deliveries(id) ON DELETE CASCADE,

    attempt_number INTEGER NOT NULL,

    request_headers JSONB,
    request_body JSONB,

    response_status INTEGER,
    response_headers JSONB,
    response_body TEXT,

    error_message TEXT,
    duration_ms INTEGER,

    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_attempts_delivery
ON delivery_attempts(delivery_id);

CREATE INDEX idx_attempts_created
ON delivery_attempts(created_at);

-- API KEYS
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    key_hash TEXT UNIQUE NOT NULL,
    name TEXT,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE INDEX idx_api_keys_application
ON api_keys(application_id);
