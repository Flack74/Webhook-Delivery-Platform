```
WEBHOOK DELIVERY PLATFORM - COMPLETE ARCHITECTURE
==================================================

┌─────────────────────────────────────────────────────────────────────────────┐
│                         YOUR APPLICATION (e.g., Stripe)                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                    Sends webhook events to your system                      │
│                        POST /events (API Key auth)                          │
└─────────────────────────────┬───────────────────────────────────────────────┘
                              │
                              ↓
┌─────────────────────────────────────────────────────────────────────────────┐
│                         API SERVICE (Go + Gin)                              │
├─────────────────────────────────────────────────────────────────────────────┤
│  ✓ Validate API Key (compare SHA256 hash)                                   │
│  ✓ Parse JSON payload                                                       │
│  ✓ Store event in PostgreSQL (immutable)                                    │
│  ✓ Fan-out: For each active endpoint, create delivery record                │
│  ✓ Enqueue delivery jobs to Redis Streams                                   │
│  ✓ Return 202 Accepted immediately (don't wait)                             │
└─────────────────────────────┬───────────────────────────────────────────────┘
                              │
                ┌─────────────┼─────────────┐
                │             │             │
                ↓             ↓             ↓
        ┌─────────┐   ┌─────────┐   ┌─────────┐
        │PostgreSQ│   │PostgreSQ│   │PostgreSQ│
        │ Event   │   │ Event   │   │ Event   │
        │ 1       │   │ 2       │   │ 3       │
        └────┬────┘   └────┬────┘   └────┬────┘
             │             │             │
             └─────────────┼─────────────┘
                           │
                           ↓
        ┌──────────────────────────────────────┐
        │    Redis Streams: webhooks:pending   │
        │  (persistent, durable message queue) │
        │                                      │
        │  Message 1: {event_id, endpoint_id}  │
        │  Message 2: {event_id, endpoint_id}  │
        │  Message 3: {event_id, endpoint_id}  │
        └──────────────┬───────────────────────┘
                       │
        ┌──────────────┼──────────────┐
        │              │              │
        ↓              ↓              ↓
  ┌──────────┐  ┌──────────┐  ┌──────────┐
  │ Worker 1 │  │ Worker 2 │  │ Worker N │
  │(Goroutin)│  │(Goroutin)│  │(Goroutin)│
  └──────┬───┘  └──────┬───┘  └──────┬───┘
         │             │             │
         └─────────────┼─────────────┘
                       │
         ┌─────────────┴─────────────┐
         │                           │
         ↓                           ↓
    ┌─────────────┐          ┌─────────────────┐
    │ Load Data:  │          │ Build Request:  │
    │ -delivery   │          │ -Marshal payload│
    │ -event      │          │ -Compute HMAC   │
    │ -endpoint   │          │ -Add headers    │
    └─────┬───────┘          └────────┬────────┘
          │                           │
          └───────────────┬───────────┘
                          │
                          ↓
              ┌──────────────────────┐
              │   HTTP POST Request  │
              │  (with signature)    │
              │  Timeout: 30 seconds │
              └──────────┬───────────┘
                         │
          ┌──────────────┼──────────────┐
          │              │              │
          ↓              ↓              ↓
      ┌──────┐      ┌──────┐      ┌──────────────┐
      │ 2xx  │      │ 5xx  │      │ Timeout/     │
      │Success      │Retry │      │Network Err   │
      └───┬──┘      └───┬──┘      └──────┬───────┘
          │             │               │
          ↓             ↓               ↓
    ┌──────────┐  ┌──────────────┐  ┌──────────┐
    │ Mark as  │  │ Schedule     │  │ Retry    │
    │ SUCCESS  │  │ NEXT RETRY:  │  │ (same)   │
    │          │  │              │  │          │
    │Update:   │  │ Attempt 1→2: │  │          │
    │ status=  │  │   5s delay   │  │          │
    │succeeded │  │ Attempt 2→3: │  │          │
    │          │  │   5m delay   │  │          │
    │Record    │  │ Attempt 3→4: │  │          │
    │attempt   │  │   30m delay  │  │          │
    │          │  │ ...etc...    │  │          │
    │Remove    │  │              │  │          │
    │from      │  │ Update:      │  │          │
    │queue     │  │ status=      │  │          │
    │          │  │ pending      │  │          │
    │          │  │ attempt_count+1 │          │
    │          │  │ next_retry_at   │          │
    │          │  │              │  │          │
    │          │  │ Re-enqueue to│  │          │
    │          │  │ Redis (delayed) │          │
    └──────────┘  └──────────────┘  └──────────┘
          │             │                │
          │             │                │
          └─────────────┼────────────────┘
                        │
                        ↓
           ┌────────────────────────────┐
           │  PostgreSQL: delivery_     │
           │  attempts (log entry)      │
           │                            │
           │ {                          │
           │   delivery_id,             │
           │   attempt_number,          │
           │   response_status,         │
           │   response_body,           │
           │   error_message,           │
           │   duration_ms,             │
           │   created_at               │
           │ }                          │
           └────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                      SPECIAL CASES & EDGE CASES                              │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. MAX RETRIES REACHED (14 attempts over 24-48 hours)                       │
│     └→ Move to DEAD LETTER QUEUE                                             │
│     └→ Update: status = 'dead_letter'                                        │
│     └→ Dashboard shows failed webhook for manual replay                      │
│                                                                              │
│  2. PERMANENT FAILURE (4xx non-retryable, customer rejected)                 │
│     └→ Stop immediately                                                      │
│     └→ Mark: status = 'failed'                                               │
│     └→ Customer can diagnose and replay                                      │
│                                                                              │
│  3. THUNDERING HERD PREVENTION (Jitter)                                      │
│     └→ If 1000 webhooks scheduled for 10:00:00                               │
│     └→ Add random 0-50% jitter to delay                                      │
│     └→ Spread across 10:00:00 - 10:05:00                                     │
│     └→ Prevents overwhelming customer endpoint                               │
│                                                                              │
│  4. RATE LIMITING                                                            │
│     └→ If customer endpoint returns 429 (Too Many Requests)                  │
│     └→ Reduce delivery rate (e.g., 10 req/sec → 1 req/sec)                   │
│     └→ Respect their capacity                                                │ 
│                                                                              │
│  5. CUSTOMER RECONNECTION                                                    │
│     └→ Customer endpoint was down for 3 hours                                │
│     └→ Retry schedule: 5s, 5m, 30m... reaches 2h later                       │
│     └→ Endpoint recovers → Successfully delivered ✓                          │
│                                                                              │
│  6. MANUAL REPLAY                                                            │
│     └→ Customer says: "My endpoint was failing, now fixed"                   │
│     └→ POST /events/{id}/deliveries/{ep_id}/replay                           │
│     └→ Reset: status='pending', attempt_count=0                              │
│     └→ Immediately enqueue for delivery                                      │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         MONITORING & OBSERVABILITY                           │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Prometheus Metrics:                                                         │
│  - webhooks_events_created_total (counter)                                   │
│  - webhooks_deliveries_processed_total (counter, labeled by status)          │
│  - webhooks_delivery_duration_seconds (histogram)                            │
│  - webhooks_queue_depth (gauge)                                              │
│  - webhooks_dead_letter_count (gauge)                                        │
│                                                                              │
│  Grafana Dashboards:                                                         │
│  - Events Created (rate)                                                     │
│  - Delivery Success Rate (%)                                                 │
│  - Average Delivery Latency (p50, p95, p99)                                  │
│  - Queue Depth (pending webhooks)                                            │
│  - Dead Letter Queue Size                                                    │
│                                                                              │
│  Logging:                                                                    │
│  - Every delivery attempt logged with full context                           │
│  - Error messages and stack traces                                           │
│  - Request/response headers and body (truncated)                             │
│                                                                              │
│  Alerts:                                                                     │
│  - Dead letter rate > 1%                                                     │
│  - Queue depth > 1 million                                                   │
│  - Delivery latency p95 > 5 seconds                                          │
│  - Worker pool unhealthy                                                     │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         CUSTOMER DASHBOARD (React/HTMX)                      │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  Main Views:                                                                 │
│  1. Events List                                                              │
│     - Filter by event type, date range                                       │
│     - Click to see payload                                                   │
│                                                                              │
│  2. Delivery Details                                                         │
│     - Status: pending/in_progress/succeeded/failed/dead_letter               │
│     - Attempt count, next retry time                                         │
│     - Full HTTP request/response logs                                        │
│                                                                              │
│  3. Endpoints Management                                                     │
│     - Add/edit/delete webhook endpoints                                      │
│     - View secret (masked)                                                   │
│     - Test endpoint                                                          │
│     - View health status                                                     │
│                                                                              │
│  4. Statistics                                                               │
│     - Events sent (24h, 7d, 30d)                                             │
│     - Delivery rate (%)                                                      │
│     - Average delivery time                                                  │
│     - Failed endpoints                                                       │
│                                                                              │
│  5. Actions                                                                  │
│     - Replay failed webhook                                                  │
│     - View full attempt history                                              │
│     - Search by message ID, timestamp                                        │ 
│     - Export logs                                                            │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────────────────┐
│                         SECURITY FEATURES                                    │
├──────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. HMAC-SHA256 Signing                                                      │
│     - Every webhook signed with customer's secret                            │
│     - Customer verifies: hash(payload, secret) == provided signature         │
│     - Prevents tampering and proves authenticity                             │
│                                                                              │
│  2. API Key Authentication                                                   │
│     - Bearer token in Authorization header                                   │
│     - Keys never stored raw (SHA256 hash only)                               │
│     - Can rotate/revoke keys                                                 │
│                                                                              │
│  3. SSRF Prevention                                                          │
│     - Reject private IPs (127.0.0.1, 192.168.*, 10.*)                        │
│     - Reject metadata services (169.254.*, 127.0.0.1)                        │
│     - Validate URL format and hostname                                       │
│                                                                              │
│  4. Rate Limiting                                                            │
│     - Per-endpoint rate limit (default 1000 req/sec)                         │
│     - Token bucket algorithm                                                 │
│     - Prevent DOS of customers' endpoints                                    │
│                                                                              │
│  5. HTTPS Only                                                               │
│     - Reject non-HTTPS webhook URLs                                          │
│     - TLS encryption for all webhooks                                        │
│     - Certificate verification                                               │
│                                                                              │
│  6. Request Timeout                                                          │
│     - 30 second max timeout per delivery                                     │
│     - Prevent hanging connections                                            │
│     - Free up worker goroutines                                              │
│                                                                              │
└──────────────────────────────────────────────────────────────────────────────┘
```

***
