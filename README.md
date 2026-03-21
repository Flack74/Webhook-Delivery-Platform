# 🚀 Webhook Delivery Platform

> A durability-first webhook delivery system built with Go, PostgreSQL, and Redis, focused on authenticated event ingestion, asynchronous fan-out, signed delivery, and retryable processing.

[![Go Version](https://img.shields.io/badge/Go-1.25.6-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io/)
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
* [License](#-license)

---

## 🎯 Overview

The **Webhook Delivery Platform** is a local-development webhook infrastructure project that accepts events from authenticated producer applications, stores them durably in PostgreSQL, fans them out to registered endpoints, and delivers them asynchronously through Redis-backed workers.

The current codebase has moved beyond a simple `POST /events` prototype. It now includes **application bootstrap**, **hashed API keys**, **endpoint management**, **idempotent event ingestion**, **delivery attempt logging**, **retry scheduling**, **dead-letter handling**, and **HMAC-signed outbound requests**.

### Key Goals

* 🔒 **Durability First**: Events, deliveries, and attempts are persisted in PostgreSQL before delivery work is considered complete
* 🔑 **Authenticated Producers**: Applications use API keys, stored only as hashes, to access protected routes
* ⚡ **Asynchronous Delivery**: Event ingestion is decoupled from outbound delivery through Redis queues
* 🔁 **Retryable Processing**: Retryable failures are rescheduled with backoff into a delayed queue
* 🧾 **Auditability**: Every outbound delivery attempt is recorded with request/response metadata
* 🔐 **Signed Webhooks**: Outbound deliveries include HMAC headers so receivers can verify authenticity
* 🧱 **Recovery-Oriented Design**: Stalled jobs in Redis processing lanes are recovered on startup

> Note: This project is currently optimized for local development and learning. It is not yet a production-ready control plane.

---

## 🏗️ Architecture

### High-Level Flow

```text
Producer App -> API (Gin) -> PostgreSQL -> Redis Queue -> Worker Pool -> Customer Endpoint
```

### Detailed Architecture

```text
┌──────────────────────────────────────────────────────────────────────┐
│                    Producer Application / Tenant                     │
│     Creates an application, receives API key, sends event data       │
└───────────────────────────────┬──────────────────────────────────────┘
                                │
                                ↓
┌──────────────────────────────────────────────────────────────────────┐
│                         API Service (Gin)                            │
│  POST /v1/applications                                               │
│  POST /v1/applications/:id/api-keys                                  │
│  POST /v1/applications/:id/endpoints                                 │
│  POST /v1/events                                                     │
│  • Validate Bearer API key                                           │
│  • Validate request payload                                          │
│  • Enforce idempotency per application                               │
│  • Persist event and delivery rows                                   │
│  • Enqueue delivery IDs to Redis                                     │
│  • Return 202 Accepted                                               │
└───────────────────────────────┬──────────────────────────────────────┘
                                │
                                ↓
┌──────────────────────────────────────────────────────────────────────┐
│                    PostgreSQL (Source of Truth)                      │
│  • Applications                                                      │
│  • API keys (hashed)                                                 │
│  • Endpoints                                                         │
│  • Events                                                            │
│  • Deliveries                                                        │
│  • Delivery attempts                                                 │
└───────────────────────────────┬──────────────────────────────────────┘
                                │
                                ↓
┌──────────────────────────────────────────────────────────────────────┐
│                         Redis Queue Layer                            │
│  • Main queue                                                        │
│  • Processing queue                                                  │
│  • Delayed retry sorted set                                          │
│  • Startup recovery for stalled jobs                                 │
└───────────────────────────────┬──────────────────────────────────────┘
                                │
                                ↓
┌──────────────────────────────────────────────────────────────────────┐
│                     Worker Pool (Goroutines)                        │
│  • Reserve delivery ID                                              │
│  • Load delivery + event + endpoint from PostgreSQL                 │
│  • Sign request with endpoint secret                                │
│  • POST payload to customer endpoint                                │
│  • Record attempt metadata                                          │
│  • Ack success / schedule retry / mark dead_letter                  │
└──────────────────────────────────────────────────────────────────────┘
```

**Ingestion Order (Current Design):**  
Authenticate application → Persist event → Create delivery rows → Enqueue delivery IDs → Return `202 Accepted`

**Source of Truth Rule:**  
PostgreSQL owns business state. Redis only carries delivery IDs for execution.

---

## 🛠️ Tech Stack

| Component | Technology | Purpose |
| --- | --- | --- |
| Language | Go 1.25.6 | Core backend runtime |
| Framework | Gin | HTTP routing, middleware, request validation |
| Database | PostgreSQL 15 | Durable storage for applications, events, deliveries, and attempts |
| Queue | Redis 7 | Main queue, processing queue, delayed retry queue |
| DB Driver | pgx/v5 | PostgreSQL access and pooling |
| Query Helper | scany | Mapping pgx query results into structs |
| Config | godotenv | Local environment variable loading |
| Containerization | Docker Compose | Local PostgreSQL and Redis services |
| Delivery Security | HMAC-SHA256 | Outbound webhook signature generation |

---

## 📊 Performance Benchmark (k6 Load Testing)

The repository includes a `k6` script at [`webhook_loadtest.js`](/home/flack/Documents/Go_Projects/Webhook-delivery-platform/webhook_loadtest.js) as a starting point for ingestion load testing.

### Current State

* A benchmark harness exists in the repo
* The platform now uses authenticated `/v1` routes and `Idempotency-Key` headers
* No canonical benchmark numbers are currently documented in-repo for the latest authenticated flow

### Suggested Benchmark Scope

* Measure `POST /v1/events` throughput under valid Bearer API key authentication
* Include an `Idempotency-Key` per request
* Run against local PostgreSQL + Redis with workers enabled
* Track throughput, p95 latency, queue depth, retry volume, and delivery success rate

> ⚠️ If you want meaningful numbers, update the current `k6` script to target the authenticated `/v1/events` route before relying on results.

---

## ✨ Features

### ✅ Implemented (Current)

* Application creation with bootstrap API key generation
* Additional API key creation per application
* API key authentication middleware with hashed key lookup
* Endpoint registration per application
* Event ingestion API with `Idempotency-Key` enforcement
* PostgreSQL JSONB event storage
* Fan-out from one event to multiple active endpoints
* Redis-backed asynchronous delivery queue
* Processing queue plus delayed retry queue
* Worker pool for concurrent delivery handling
* HMAC-signed outbound webhook requests
* Delivery status lifecycle: `pending`, `in_progress`, `succeeded`, `failed`, `dead_letter`
* Delivery attempt persistence with request/response metadata
* Retry classification for retryable HTTP and network failures
* Crash recovery for jobs left in the processing queue
* Docker Compose local development environment

### 🚧 Planned / Incomplete

* Canonical load-test metrics for the current authenticated API flow
* Better migration workflow and migration runner ergonomics
* Endpoint update, disable, delete, and listing APIs
* Replay tooling for failed or dead-letter deliveries
* Observability metrics and dashboards
* Stronger security hardening such as stricter endpoint validation and SSRF protections

---

## 🗄️ Database Schema

### Applications

```sql
CREATE TABLE applications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    is_active BOOLEAN DEFAULT TRUE
);
```

### API Keys

```sql
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    key_hash TEXT UNIQUE NOT NULL,
    name TEXT,
    last_used_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT now()
);
```

### Endpoints

```sql
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
```

### Events

```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id UUID NOT NULL REFERENCES applications(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    idempotency_key TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(application_id, idempotency_key)
);
```

### Deliveries

```sql
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    endpoint_id UUID NOT NULL REFERENCES endpoints(id) ON DELETE CASCADE,
    status TEXT NOT NULL DEFAULT 'pending',
    attempt_count INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    last_error TEXT,
    max_attempts INTEGER NOT NULL DEFAULT 5,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE (event_id, endpoint_id)
);
```

### Delivery Attempts

```sql
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
```

---

## 📁 Project Structure

```text
Webhook-Delivery-Platform/
├── cmd/
│   ├── api/
│   │   └── main.go            # API bootstrap, queue recovery, worker startup
│   └── receiver/
│       └── main.go            # Simple local webhook receiver for manual testing
├── internal/
│   ├── handler/               # HTTP handlers and route wiring
│   ├── middleware/            # API key auth middleware
│   ├── models/                # Domain models
│   ├── postgres/              # PostgreSQL connection setup
│   ├── queue/                 # Redis queue abstraction
│   ├── redis/                 # Redis client setup
│   ├── repository/            # Persistence layer for applications/events/deliveries
│   ├── utils/                 # API key, secret, and HMAC helpers
│   └── worker/                # Delivery workers and retry logic
├── migrations/
│   ├── 001_init.sql           # Main schema
│   └── 002_drop.sql           # Drop script
├── docker-compose.yml         # Local PostgreSQL and Redis services
├── webhook_loadtest.js        # k6 starter script
├── Architecture.md            # Supplemental architecture notes
└── README.md
```

---

## 🎯 Development Phases

* ✅ Phase 0: Basic event ingestion prototype
* ✅ Phase 1: PostgreSQL-backed event persistence
* ✅ Phase 2: Redis-based asynchronous queue and worker pool
* ✅ Phase 3: Durable delivery tracking and attempt logging
* ✅ Phase 4: Application model, API keys, and endpoint registration
* ✅ Phase 5: Idempotent ingestion and delivery fan-out
* ✅ Phase 6: Retry scheduling, delayed queue, and dead-letter state
* ✅ Phase 7: HMAC-signed outbound webhook delivery
* 🚧 Phase 8: Operational hardening, observability, and admin workflows

---

## 🚀 Getting Started

### Prerequisites

* Go 1.25.6
* Docker and Docker Compose
* PostgreSQL 15
* Redis 7

### Setup

```bash
git clone https://github.com/Flack74/Webhook-Delivery-Platform.git
cd Webhook-Delivery-Platform
```

Create a local `.env` file:

```env
DB_NAME=webhook_delivery_platform
DB_USER=webhookuser
DB_PASSWORD=changeme
DB_HOST=localhost
DB_PORT=5432
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASS=
GIN_MODE=debug
```

### Run Infrastructure

```bash
docker compose up -d
```

### Apply Schema

Apply [`migrations/001_init.sql`](/home/flack/Documents/Go_Projects/Webhook-delivery-platform/migrations/001_init.sql) to your PostgreSQL database using your preferred migration tool or `psql`.

### Run the API

```bash
go run ./cmd/api
```

### Run a Local Receiver

```bash
go run ./cmd/receiver
```

The receiver listens on `http://localhost:9000/webhook` and is useful for manual delivery testing.

---

## ⚙️ Configuration

### Required Environment Variables

| Variable | Required | Purpose |
| --- | --- | --- |
| `DB_NAME` | Yes | PostgreSQL database name |
| `DB_USER` | Yes | PostgreSQL username |
| `DB_PASSWORD` | Yes | PostgreSQL password |
| `DB_HOST` | Yes | PostgreSQL host |
| `DB_PORT` | Yes | PostgreSQL port |

### Optional Environment Variables

| Variable | Default | Purpose |
| --- | --- | --- |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASS` | empty | Redis password |
| `GIN_MODE` | `debug` | Gin mode; use `production` for release mode |

### Current Queue Names

The application currently initializes Redis with these hard-coded queue names:

* `deliveries_queue`
* `processing_queue`
* `deliveries_delayed`

---

## 🧪 Testing

### 1. Create an Application

```bash
curl -X POST http://localhost:8000/v1/applications \
  -H "Content-Type: application/json" \
  -d '{
    "name": "demo-app"
  }'
```

Save the returned `api_key`, because the raw key is only returned once.

### 2. Register an Endpoint

```bash
curl -X POST http://localhost:8000/v1/applications/<application_id>/endpoints \
  -H "Authorization: Bearer <api_key>" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "http://localhost:9000/webhook",
    "description": "local receiver"
  }'
```

### 3. Send an Event

```bash
curl -X POST http://localhost:8000/v1/events \
  -H "Authorization: Bearer <api_key>" \
  -H "Idempotency-Key: evt_001" \
  -H "Content-Type: application/json" \
  -d '{
    "event_type": "user.created",
    "data": {
      "id": "usr_123",
      "email": "user@example.com"
    }
  }'
```

Expected response:

```json
{
  "status": "accepted",
  "event_id": "uuid",
  "deliveries_queued": 1
}
```

### 4. Load Testing with k6

```bash
k6 run webhook_loadtest.js
```

> Note: The checked-in script is a starter and should be updated to use the authenticated `/v1/events` endpoint plus an `Idempotency-Key`.

---

## 🗺️ Roadmap

* Add richer endpoint management APIs
* Add replay and inspection tooling for deliveries
* Add metrics, dashboards, and queue observability
* Add stricter endpoint validation and security controls
* Add cleaner migration workflow and automated setup
* Add production deployment and scaling guidance

---

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test locally with PostgreSQL and Redis running
5. Open a Pull Request

---

## 📄 License

MIT License

---

<div align="center">
⭐ Star this repo if you find it useful!  
Built with ❤️ by Flack
</div>
