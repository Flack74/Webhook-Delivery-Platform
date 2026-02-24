# 🚀 Webhook Delivery Platform

> A scalable webhook delivery system built with Go, PostgreSQL, and Redis, focused on reliability, durability, and asynchronous processing.

[![Go Version](https://img.shields.io/badge/Go-1.25.6-00ADD8?style=flat\&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat\&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat\&logo=redis)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 📋 Table of Contents

* [Overview](#-overview)
* [Architecture](#-architecture)
* [Tech Stack](#-tech-stack)
* [Performance Benchmark (k6 Load Testing)](#-performance-benchmark-k6-load-testing)
* [Features](#-features)
* [Database Schema](#-database-schema)
* [Project Structure](#-project-structure)
* [Development Phases](#-development-phases)
* [Getting Started](#-getting-started)
* [Configuration](#-configuration)
* [Testing](#-testing)
* [Roadmap](#-roadmap)
* [Contributing](#-contributing)

---

## 🎯 Overview

The **Webhook Delivery Platform** is a robust, scalable system designed to reliably ingest, store, and asynchronously deliver webhook events to customer endpoints. Inspired by real-world systems like Stripe, Svix, and Convoy, this project emphasizes durability, fault tolerance, and clean system design.

### Key Goals

* 🔒 **Durability First**: Events are persisted before queueing to prevent data loss
* ⚡ **Asynchronous Processing**: Decoupled ingestion and delivery via Redis queue
* 🔄 **Crash Recovery**: Automatic recovery of stalled jobs on restart
* 📊 **Observability-Ready Design**: Clear delivery tracking and audit trail
* 🧱 **Production-Inspired Architecture**: Reliable queue patterns and worker pools
* 🚧 **Extensible**: Retry logic, DLQ, and metrics planned in future phases

> Note: The system is currently in active local development and not yet deployed to production.

---

## 🏗️ Architecture

### High-Level Flow

```
Client → API (Gin) → PostgreSQL (Persistence) → Redis Queue → Worker Pool → Customer Endpoint
```

### Detailed Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Application                       │
│                  (Sends webhook events)                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                   API Service (Gin)                         │
│  POST /events                                               │
│  • Validate payload                                         │
│  • Generate UUID                                            │
│  • Persist event (PostgreSQL)                               │
│  • Create delivery record                                   │
│  • Enqueue job to Redis                                     │
│  • Return 202 Accepted                                      │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL (Durable Store)               │
│  • Immutable event storage (JSONB)                          │
│  • Delivery state tracking                                  │
└─────────────────────────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│              Redis Queue (Reliable Pattern)                 │
│  • BRPopLPush (processing queue)                            │
│  • AOF Persistence                                          │
│  • Stalled job recovery                                     │
└─────────────────────────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                Worker Pool (Goroutines)                     │
│  • Concurrent HTTP delivery                                 │
│  • Ack on success / Nack on failure                         │
│  • Update delivery status in PostgreSQL                     │
└─────────────────────────────────────────────────────────────┘
```

**Ingestion Order (Durability-First Design):**
Persist to PostgreSQL → Create delivery record → Enqueue to Redis → Return `202 Accepted`

This ensures zero event loss even if the worker or Redis fails after ingestion.

---

## 🛠️ Tech Stack

| Component        | Technology              | Purpose                               |
| ---------------- | ----------------------- | ------------------------------------- |
| Language         | Go 1.25.6               | High-performance concurrent backend   |
| Framework        | Gin                     | Fast HTTP routing and middleware      |
| Database         | PostgreSQL 15           | Persistent event and delivery storage |
| Queue            | Redis 7 (AOF Enabled)   | Durable message queue                 |
| DB Driver        | pgx/v5                  | Native PostgreSQL driver with pooling |
| Containerization | Docker & Docker Compose | Reproducible local environment        |
| Queue Pattern    | BRPopLPush              | Reliable job processing with recovery |

---

## 📊 Performance Benchmark (k6 Load Testing)

Synthetic load testing was conducted on the `/events` ingestion API using k6 in a local containerized environment (Go + Gin + PostgreSQL + Redis + async queue).

### Test Setup

* Environment: Local machine (Dockerized PostgreSQL & Redis)
* Endpoint: `POST /events`
* Response: `202 Accepted` (async processing)
* Duration: 60 seconds per run
* Workers: 3 concurrent goroutines
* Queue Pattern: Redis BRPopLPush (reliable queue)

### Results

| Concurrent VUs | Throughput (req/s) | Avg Latency | P95 Latency | Error Rate |
| -------------- | ------------------ | ----------- | ----------- | ---------- |
| 50             | ~448 req/s         | 111 ms      | ~120 ms     | 0%         |
| 100            | ~433 req/s         | 230 ms      | ~301 ms     | 0%         |
| 200            | ~336 req/s         | 592 ms      | ~696 ms     | 0%         |
| 500            | ~334 req/s         | 1.48 s      | ~1.57 s     | 0%         |

### Key Observations

* Stable throughput plateau (~330–450 req/s) due to synchronous DB persistence
* Graceful latency degradation under higher concurrency
* 0% error rate across all tested loads
* Ingestion remains stable because delivery is fully asynchronous

> ⚠️ Metrics are based on synthetic local testing using k6 and are indicative benchmarks, not production deployment metrics.

---

## ✨ Features

### ✅ Implemented (Current)

* Event Ingestion API (`POST /events`)
* JSON payload validation with Gin bindings
* PostgreSQL JSONB immutable event storage
* Delivery tracking with status lifecycle
* Redis durable queue (AOF persistence)
* Reliable queue pattern using BRPopLPush
* Asynchronous ingestion → delivery decoupling
* Worker pool (concurrent goroutines)
* Crash recovery for stalled jobs on startup
* Structured error handling and logging
* Docker Compose local development setup
* Connection pooling via pgx

### 🚧 Planned (Next Phases)

* Retry logic with exponential backoff
* Dead Letter Queue (DLQ)
* HMAC-SHA256 webhook signing
* Prometheus metrics & Grafana dashboards
* Rate limiting & circuit breakers
* Admin dashboard for replay & inspection

---

## 🗄️ Database Schema

### Events Table (Immutable Storage)

```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### Deliveries Table (Delivery Lifecycle Tracking)

```sql
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    endpoint_url TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending','in_progress','succeeded','failed')),
    attempts INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, endpoint_url)
);
```

---

## 📁 Project Structure

```
Webhook-Delivery-Platform/
├── cmd/
│   └── api/
│       └── main.go          # API entry point
├── internal/
│   ├── handler/             # HTTP handlers
│   ├── repository/          # DB layer (pgx)
│   ├── queue/               # Redis queue abstraction
│   ├── worker/              # Worker pool & delivery logic
│   └── redis/               # Redis client setup
├── migrations/              # SQL migrations
├── docker-compose.yml       # Local DB & Redis setup
├── .env.example             # Environment template
└── README.md
```

---

## 🚀 Getting Started

### Prerequisites

* Go 1.22+
* Docker & Docker Compose
* PostgreSQL (via Docker)
* Redis (via Docker)

### Setup

```bash
git clone https://github.com/Flack74/Webhook-Delivery-Platform.git
cd Webhook-Delivery-Platform
cp .env.example .env
```

### Example `.env`

```env
DB_NAME=webhook_delivery_platform
DB_USER=webhookuser
DB_PASSWORD=changeme
DB_HOST=localhost
DB_PORT=5432
REDIS_HOST=localhost
REDIS_PORT=6379
```

> ⚠️ Never commit real credentials. Use `.env.example` for safe configuration.

### Run Services

```bash
docker compose up -d
```

### Run the API

```bash
go run cmd/api/main.go
```

---

## 🧪 Testing

### Manual API Test

```bash
curl -X POST http://localhost:8000/events \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "user.created",
    "data": {
      "id": "usr_123",
      "email": "user@example.com"
    }
  }'
```

Expected Response:

```json
{
  "status": "accepted",
  "event_id": "uuid"
}
```

### Load Testing (k6)

```bash
k6 run --vus 50 --duration 60s webhook_loadtest.js
```

---

## 🎯 Development Phases

* ✅ Phase 0: Event ingestion API
* ✅ Phase 1: Synchronous delivery prototype
* ✅ Phase 2: Async boundary (queue + workers)
* ✅ Phase 3: PostgreSQL persistence layer
* ✅ Phase 4: Durable Redis queue + crash recovery
* 🚧 Phase 5: Retry & exponential backoff (planned)
* 🚧 Phase 6: Observability & metrics
* 🚧 Phase 7: Security (HMAC signing)
* 🚧 Phase 8: Production deployment & scaling

---

## 🗺️ Roadmap

* Retry scheduler with backoff strategy
* Dead letter queue (DLQ)
* Multi-endpoint delivery
* Prometheus + Grafana monitoring
* Horizontal worker scaling
* Kubernetes deployment

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push and open a Pull Request

---

## 📄 License

MIT License

---

<div align="center">
⭐ Star this repo if you find it useful!  
Built with ❤️ by Flack
</div>
