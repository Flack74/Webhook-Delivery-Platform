## Project Status

Phase 1 complete: Basic webhook ingestion and delivery.

The service accepts events via `POST /events`, validates the payload, and acknowledges receipt with `202 Accepted`.  
Each event is synchronously delivered to a test webhook receiver, simulating an external consumer.

Persistence, asynchronous processing, retries, and security will be added in later phases.
