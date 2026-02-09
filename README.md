# TicRes — Event Ticket Reservation System

A production-ready backend service for event ticket reservations built with **Go**, featuring real-time seat locking, payment processing, background job workers, and role-based access control.

> Built with Clean Architecture, designed for concurrency safety and horizontal scalability.

---

## Table of Contents

- [Key Engineering Highlights](#key-engineering-highlights)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Database Design](#database-design)
- [API Endpoints](#api-endpoints)
- [Getting Started](#getting-started)
- [API Documentation Screenshots](#api-documentation-screenshots)

---

## Key Engineering Highlights

### Concurrency-Safe Seat Booking
Prevents double-booking through **pessimistic locking** at the database level. Seat reservation uses atomic `UPDATE ... WHERE is_booked = FALSE` queries inside transactions — if two users try to book the same seat simultaneously, only one succeeds.

### Background Worker with Graceful Shutdown
A channel-based **async job worker** handles mass refund processing and email notifications without blocking HTTP responses. On event cancellation, the admin gets an instant response while refunds are processed in the background. The worker drains its job queue before shutdown using `sync.WaitGroup`.

### Payment State Machine
Bookings follow a strict state lifecycle: `PENDING → PAID / EXPIRED / REFUNDED / CANCELLED`. Each transition is validated — expired bookings automatically release seats, and duplicate payments are rejected. Payment methods (credit card, bank transfer, e-wallet) generate unique external IDs for gateway integration.

### Redis Caching with Invalidation
Event listings are cached in Redis with **10-minute TTL** and **explicit invalidation** on create/update/delete. Cache failures degrade gracefully — the app falls back to PostgreSQL without errors.

### Clean Architecture with Strict Layer Separation
```
Handler (HTTP) → Usecase (Business Logic) → Repository (Data Access) → Database
```
Each layer communicates through **interfaces**, making every component independently testable with mock implementations. No layer skips — handlers never touch the database directly.

### Production-Grade Patterns
- **Connection pooling** (pgx) with tuned pool size, lifetime, and idle timeout
- **Context timeouts** on all usecase operations to prevent hanging requests
- **Structured logging** (Zap) with environment-specific output (dev: pretty, prod: JSON)
- **Graceful HTTP shutdown** with signal handling (`SIGINT`, `SIGTERM`)
- **Database migrations** with versioned SQL files (golang-migrate)
- **bcrypt password hashing** with time-safe comparison
- **Request validation** using declarative struct tags

---

## Architecture

```
cmd/
  api/main.go              → Entry point, DI wiring, graceful shutdown
  seed/main.go             → Database seeder (admin account + sample events)

internal/
  config/                  → Environment config (Viper, 12-factor app)
  entity/                  → Domain models + domain-specific errors
  repository/              → PostgreSQL queries + Redis caching layer
  usecase/                 → Business logic orchestration
  usecase/mocks/           → Testify mock implementations
  delivery/http/           → Gin HTTP handlers
  delivery/http/middleware/ → JWT auth + admin RBAC middleware
  worker/                  → Background notification & refund worker

pkg/
  database/                → PostgreSQL pool + Redis client setup
  logger/                  → Structured logging (Zap)
  response/                → HTTP response helpers

db/migrations/             → 9 versioned SQL migration files
```

---

## Tech Stack

| Layer | Technology |
|---|---|
| Language | Go 1.24 |
| HTTP Framework | Gin |
| Database | PostgreSQL |
| Caching | Redis |
| Auth | JWT (HMAC-SHA256) + bcrypt |
| Logging | Uber Zap |
| Config | Viper (.env / env vars) |
| Migrations | golang-migrate |
| Testing | testify (mock + assert), testcontainers |
| API Docs | Swagger / OpenAPI (swaggo) |
| Containerization | Docker, Docker Compose |

---

## Database Design

![Database Diagram](docs/screenshots/ticresERD.png)

**6 tables** with foreign keys, ENUM types, and proper constraints:

| Table | Purpose | Key Details |
|---|---|---|
| `users` | User accounts | Unique email, bcrypt password, role ENUM (`admin`, `user`) |
| `events` | Event listings | Status ENUM (`available`, `cancelled`, `completed`), capacity tracking |
| `seats` | Individual seats per event | `is_booked` flag for pessimistic locking, `price` as DECIMAL |
| `booking` | Reservation records | Status lifecycle, `expires_at` for 15-min payment window, FK to user + event |
| `booking_items` | Booking ↔ Seat junction | Many-to-many relationship |
| `transactions` | Payment records | 1:1 with booking, external ID for gateway, payment method tracking |
| `refund` | Refund tracking | Amount, reason, status, linked to booking |

**Key constraints:** Foreign keys with referential integrity, unique email, unique booking-transaction relationship, DECIMAL(10,2) for monetary values.

---

## API Endpoints

### Public
| Method | Endpoint | Description |
|---|---|---|
| POST | `/api/v1/register` | Register new user |
| POST | `/api/v1/login` | Login, returns JWT token |
| GET | `/api/v1/events` | List events (search + pagination) |
| GET | `/api/v1/events/:id` | Event detail with available seats |

### Protected (JWT Required)
| Method | Endpoint | Description |
|---|---|---|
| GET | `/api/v1/me` | Current user profile |
| GET | `/api/v1/me/bookings` | User's booking history |
| POST | `/api/v1/events` | Create new event |
| POST | `/api/v1/bookings` | Book seats (with seat locking) |
| POST | `/api/v1/payments` | Process payment for booking |
| GET | `/api/v1/payments/:booking_id` | Check payment status |

### Admin (JWT + Admin Role)
| Method | Endpoint | Description |
|---|---|---|
| PUT | `/api/v1/admin/events/:id` | Update event |
| DELETE | `/api/v1/admin/events/:id` | Cancel event (triggers background refunds) |
| GET | `/api/v1/admin/bookings` | View all bookings |
| GET | `/api/v1/admin/events/:id/bookings` | View bookings for specific event |

---

## Getting Started

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and Docker Compose

### Quick Start

```bash
git clone https://github.com/<your-username>/ticres.git
cd ticres
make quick-start
```

This starts **PostgreSQL**, **Redis**, runs **migrations**, and launches the **API** — all in one command.

The API will be available at `http://localhost:8080`.

### Seed Sample Data

To populate the database with an admin account and sample events:

```bash
# With local Go installed:
make seed

# Admin credentials after seeding:
# Email: admin@ticres.com
# Password: admin123
```

### Manual Setup (without Docker Compose)

```bash
# Start services
make docker-up      # PostgreSQL on port 5433
make docker-upc     # Redis on port 6379

# Run migrations
make migrate-up

# Start the API
make run
```

### Stop & Cleanup

```bash
make quick-stop     # Stop containers
make quick-clean    # Stop + remove volumes
```
