# 🚀 Webhook Delivery Platform

> A production-grade, scalable webhook delivery system built with Go, PostgreSQL, and Redis

[![Go Version](https://img.shields.io/badge/Go-1.25.6-00ADD8?style=flat&logo=go)](https://golang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

---

## 📋 Table of Contents

- [Overview](#-overview)
- [Features](#-features)
- [Architecture](#-architecture)
- [Tech Stack](#-tech-stack)
- [Getting Started](#-getting-started)
- [Database Schema](#-database-schema)
- [Project Structure](#-project-structure)
- [Development Phases](#-development-phases)
- [Configuration](#-configuration)
- [Testing](#-testing)
- [Roadmap](#-roadmap)
- [Contributing](#-contributing)

---

## 🎯 Overview

The **Webhook Delivery Platform** is a robust, system designed to reliably deliver webhook events to customer endpoints. Inspired by production systems like Svix, Stripe, and Convoy, this platform handles event ingestion, persistent storage, asynchronous delivery, retry logic with exponential backoff, and comprehensive observability.

### Why This Project?

- 🔒 **Reliability First**: Events are never lost, even during system failures
- ⚡ **High Performance**: Asynchronous processing with worker pools
- 🔄 **Smart Retries**: Exponential backoff with jitter for transient failures
- 📊 **Full Observability**: Complete audit trail of every delivery attempt
- 🛡️ **Security**: HMAC-SHA256 signing for webhook authenticity
- 🎯 **Production Ready**: Built with real-world patterns and best practices

---

## ✨ Features

### ✅ Phase 4 Complete: Durable Queue & Async Delivery

- [x] **Event Ingestion API** - Accept webhook events via `POST /events`
- [x] **Payload Validation** - JSON schema validation with Gin bindings
- [x] **PostgreSQL Persistence** - Events stored as JSONB for flexibility
- [x] **UUID Generation** - Unique identifiers for all events
- [x] **Delivery Tracking** - Separate table for delivery attempts and status
- [x] **Database Migrations** - Version-controlled schema management
- [x] **Docker Compose Setup** - One-command local development environment
- [x] **Health Checks** - PostgreSQL readiness probes
- [x] **Error Handling** - Comprehensive error responses
- [x] **Redis Queue Integration** - Durable message queue with AOF persistence
- [x] **Worker Pool** - 3 concurrent workers for parallel delivery
- [x] **Asynchronous Processing** - Decoupled ingestion from delivery
- [x] **Crash Recovery** - Automatic recovery of stalled jobs on restart
- [x] **Reliable Queue Pattern** - BRPopLPush for atomic job processing
- [x] **Job Acknowledgment** - Nack/Ack pattern for retry handling

### 🚧 Coming Soon

- [ ] **Retry Logic** - Exponential backoff (5s → 5m → 30m → 2h → 12h)
- [ ] **HMAC Signing** - Cryptographic signatures for webhook security
- [ ] **Dead Letter Queue** - Handle permanently failed deliveries
- [ ] **Metrics & Monitoring** - Prometheus metrics and Grafana dashboards
- [ ] **Rate Limiting** - Protect customer endpoints from overload
- [ ] **Admin Dashboard** - Web UI for event inspection and replay

---

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Application                       │
│                  (Sends webhook events)                     │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                   API Service (Gin)                         │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  POST /events                                        │   │
│  │  • Validate JSON payload                             │   │
│  │  • Generate UUID                                     │   │
│  │  • Store in PostgreSQL                               │   │
│  │  • Create delivery record                            │   │
│  │  • Enqueue to Redis                                  │   │
│  │  • Return 202 Accepted                               │   │
│  └──────────────────────────────────────────────────────┘   │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│                    PostgreSQL Database                      │
│  ┌──────────────────┐      ┌──────────────────────────┐     │
│  │  events          │      │  deliveries              │     │
│  │  • id (UUID)     │◄─────│  • id (UUID)             │     │
│  │  • event_type    │      │  • event_id (FK)         │     │
│  │  • payload       │      │  • endpoint_url          │     │
│  │  • created_at    │      │  • status                │     │
│  └──────────────────┘      │  • attempts              │     │
│                            │  • next_retry_at         │     │
│                            │  • last_error            │     │
│                            └──────────────────────────┘     │
└─────────────────────────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│         Redis Queue (AOF Persistence Enabled)               │
│  • BRPopLPush for reliable delivery                         │
│  • Crash recovery on startup                                │
│  • 3 concurrent workers                                     │
└─────────────────────────────────────────────────────────────┘
                         │
                         ↓
┌─────────────────────────────────────────────────────────────┐
│              Worker Pool (3 Goroutines)                     │
│  • Concurrent HTTP POST to customer endpoints               │
│  • Ack on success / Nack on failure                         │
│  • Update delivery status in PostgreSQL                     │
└─────────────────────────────────────────────────────────────┘
```

---

## 🛠️ Tech Stack

| Component | Technology | Purpose |
|-----------|-----------|---------|
| **Language** | Go 1.25.6 | High-performance, concurrent backend |
| **Web Framework** | Gin | Fast HTTP router and middleware |
| **Database** | PostgreSQL 15 | Persistent event and delivery storage |
| **Queue**  | Redis 7 (AOF) | Durable message queue with persistence |
| **Driver** | pgx/v5 | Native PostgreSQL driver with connection pooling |
| **Containerization** | Docker & Docker Compose | Local development environment |
| **JSON Handling** | encoding/json | Event payload serialization |
| **UUID** | uuid | Unique event identifiers |
| **Queue Pattern** | BRPopLPush | Reliable queue with crash recovery |

---

## 🚀 Getting Started

### Prerequisites

- **Go** 1.25.6 or higher
- **Docker** & **Docker Compose**
- **PostgreSQL** 15 (via Docker)
- **Make** (optional, for convenience commands)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Flack74/Webhook-Delivery-Platform.git
   cd Webhook-Delivery-Platform
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   ```
   
   Edit `.env`:
   ```env
   DB_NAME=webhook_delivery_platform
   DB_USER=webhookuser
   DB_PASSWORD=webhook123
   DB_PORT=5432
   DB_HOST=localhost
   ```

3. **Start PostgreSQL with Docker**
   ```bash
   docker compose --env-file .env up -d
   ```

4. **Wait for database to be ready**
   ```bash
   sleep 5
   ```

5. **Run database migrations**
   ```bash
   PGPASSWORD=webhook123 psql -h localhost -U webhookuser -d webhook_delivery_platform -f migrations/001_init.sql
   ```

6. **Install Go dependencies**
   ```bash
   go mod download
   ```

7. **Run the application**
   ```bash
   DB_NAME=webhook_delivery_platform \
   DB_USER=webhookuser \
   DB_PASSWORD=webhook123 \
   DB_PORT=5432 \
   DB_HOST=localhost \
   go run cmd/api/main.go
   ```

8. **Verify the server is running**
   ```bash
   curl http://localhost:8000/health
   ```

---

## 📡 API Documentation

### Create Event

**Endpoint:** `POST /events`

**Description:** Accept a webhook event, validate the payload, and store it in the database.

**Request Headers:**
```
Content-Type: application/json
```

**Request Body:**
```json
{
  "event_type": "user.created",
  "data": {
    "id": "usr_123",
    "email": "user@example.com"
  }
}
```

**Response:** `202 Accepted`
```json
{
  "status": "accepted",
  "event_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Error Response:** `400 Bad Request`
```json
{
  "error": "Key: 'Event.EventType' Error:Field validation for 'EventType' failed on the 'required' tag"
}
```

**Example with cURL:**
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

---

## 🗄️ Database Schema

### Events Table

Stores all incoming webhook events immutably.

```sql
CREATE TABLE events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_events_event_type ON events (event_type);
```

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique event identifier |
| `event_type` | TEXT | Event category (e.g., "user.created") |
| `payload` | JSONB | Flexible JSON payload |
| `created_at` | TIMESTAMPTZ | Event creation timestamp |

### Deliveries Table

Tracks delivery attempts for each event to customer endpoints.

```sql
CREATE TABLE deliveries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    endpoint_url TEXT NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('pending', 'in_progress', 'succeeded', 'failed')),
    attempts INTEGER NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMPTZ,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (event_id, endpoint_url)
);

CREATE INDEX idx_deliveries_pending ON deliveries (status, next_retry_at);
```

| Column | Type | Description |
|--------|------|-------------|
| `id` | UUID | Unique delivery identifier |
| `event_id` | UUID | Foreign key to events table |
| `endpoint_url` | TEXT | Customer webhook endpoint |
| `status` | TEXT | Delivery status (pending/in_progress/succeeded/failed) |
| `attempts` | INTEGER | Number of delivery attempts |
| `next_retry_at` | TIMESTAMPTZ | Scheduled retry time |
| `last_error` | TEXT | Last error message |
| `created_at` | TIMESTAMPTZ | Delivery record creation time |
| `updated_at` | TIMESTAMPTZ | Last update time |

---

## 📁 Project Structure

```
Webhook-Delivery-Platform/
├── cmd/
│   ├── api/
│   │   └── main.go              # API server entry point
│   └── receiver/
│       └── main.go              # Test webhook receiver (Phase 1)
├── internal/
│   ├── handler/
│   │   ├── events.go            # Event creation handler
│   │   └── routes.go            # Route definitions
│   ├── repository/
│   │   ├── events.go            # Event database operations
│   │   ├── deliveries.go        # Delivery database operations
│   │   └── postgres.go          # PostgreSQL connection setup
│   ├── queue/
│   │   └── queue.go             # Redis queue (Phase 4)
│   └── worker/
│       └── worker.go            # Delivery worker pool (Phase 4)
├── migrations/
│   ├── 001_init.sql             # Initial schema
│   └── 002_drop.sql             # Rollback script
├── .env                         # Environment variables
├── .gitignore
├── docker-compose.yml           # PostgreSQL container
├── go.mod                       # Go dependencies
├── go.sum
├── Architecture.md              # Detailed architecture diagram
└── README.md                    # This file
```

---

## 🎯 Development Phases

This project follows an incremental build approach, implementing one concept at a time.

### ✅ Phase 0: Event Ingestion (Complete)
- HTTP server with Gin
- `POST /events` endpoint
- Payload validation
- `202 Accepted` response

### ✅ Phase 1: Synchronous Delivery (Complete)
- Test webhook receiver
- Direct HTTP POST to endpoint
- Success/failure logging

### ✅ Phase 2: Asynchronous Boundary (Complete)
- In-memory queue with Go channels
- Worker goroutine
- Decoupled ingestion from delivery

### ✅ Phase 3: Persistent Storage (Complete)
- PostgreSQL integration
- Event and delivery tables
- Database migrations
- Connection pooling with pgx

### ✅ Phase 4: Durable Queue (Complete)
- Redis integration with AOF persistence
- Durable message queue (BRPopLPush pattern)
- Worker pool (3 concurrent workers)
- Job acknowledgment (Ack/Nack)
- Crash recovery on startup
- Enhanced error logging

### 📅 Phase 5: Retry & Backoff (Planned)
- Exponential backoff algorithm
- Retry scheduling
- Max attempt limits
- Transient vs permanent failure detection

### 📅 Phase 6: Observability (Planned)
- Delivery attempt logging
- Dead letter queue
- Metrics (Prometheus)
- Grafana dashboards

### 📅 Phase 7: Security (Planned)
- HMAC-SHA256 signing
- Timestamp validation
- Replay attack prevention

### 📅 Phase 8: Production Features (Planned)
- Rate limiting
- Manual replay API
- Admin dashboard
- Docker deployment

...

---

## ⚙️ Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_NAME` | PostgreSQL database name | `webhook_delivery_platform` |
| `DB_USER` | PostgreSQL username | `webhook8me` |
| `DB_PASSWORD` | PostgreSQL password | `webhooks74621` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |

### Database Connection String

The application constructs the connection string as:
```
postgres://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}?sslmode=disable
```

---

## 🧪 Testing

### Manual Testing

1. **Start the server**
   ```bash
   go run cmd/api/main.go
   ```

2. **Send a test event**
   ```bash
   curl -X POST http://localhost:8000/events \
     -H "Content-Type: application/json" \
     -d '{
       "event_type": "order.completed",
       "data": {
         "id": "order_789",
         "email": "customer@example.com"
       }
     }'
   ```

3. **Verify in database**
   ```bash
   PGPASSWORD=webhooks74621 psql -h localhost -U webhook8me -d webhook_delivery_platform
   ```
   ```sql
   SELECT * FROM events ORDER BY created_at DESC LIMIT 5;
   SELECT * FROM deliveries ORDER BY created_at DESC LIMIT 5;
   ```

### Unit Tests (Coming Soon)

```bash
go test ./...
```

---

## 🗺️ Roadmap

- [x] **Phase 0**: Event ingestion API
- [x] **Phase 1**: Synchronous webhook delivery
- [x] **Phase 2**: Asynchronous processing
- [x] **Phase 3**: PostgreSQL persistence
- [x] **Phase 4**: Redis queue integration
- [ ] **Phase 5**: Retry logic with exponential backoff
- [ ] **Phase 6**: Delivery attempt logging & DLQ
- [ ] **Phase 7**: HMAC-SHA256 signing
- [ ] **Phase 8**: Admin dashboard & metrics
- [ ] **Phase 9**: Multi-endpoint support
- [ ] **Phase 10**: Rate limiting & circuit breakers
- [ ] **Phase 11**: Kubernetes deployment
- [ ] **Phase 12**: Horizontal scaling & load balancing

---

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- Inspired by [Svix](https://www.svix.com/), [Convoy](https://getconvoy.io/), and [Hookdeck](https://hookdeck.com/)
- Built with guidance from [webhooks.fyi](https://webhooks.fyi/)
- PostgreSQL best practices from [pgx documentation](https://github.com/jackc/pgx)
- Gin framework by [gin-gonic](https://github.com/gin-gonic/gin)

---

## 📞 Contact

**Flack74** - [@Flack74](https://github.com/Flack74)

Project Link: [https://github.com/Flack74/Webhook-Delivery-Platform](https://github.com/Flack74/Webhook-Delivery-Platform)

---

<div align="center">

**⭐ Star this repo if you find it helpful!**

Made with ❤️ By Flack74

</div>
