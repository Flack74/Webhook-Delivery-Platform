## Project Status

**Phase 3 complete: Persistent webhook event storage**

The service accepts events via `POST /events`, validates the payload, and responds with `202 Accepted`.  
Each event is assigned a UUID and persisted in PostgreSQL with the payload stored as JSONB for flexibility and future processing.

The database schema is designed to support reliable webhook delivery, retries, and scalability in later phases.

Next phases will introduce database-driven workers, durable delivery queues, retry/backoff logic, and webhook security (HMAC signing).
