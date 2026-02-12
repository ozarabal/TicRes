# TicRes — Event Ticket Reservation System

A full-stack event ticket reservation platform where users can browse events, select seats, and book tickets with real-time availability. Admins manage events with automatic refund processing on cancellation. Built with a **Go** backend and **React** frontend, deployed on **Microsoft Azure**.

> Clean Architecture backend with concurrency-safe booking, background workers, and Redis caching — paired with a modern React + TypeScript frontend.

---

## Table of Contents

- [Key Engineering Highlights](#key-engineering-highlights)
- [Architecture](#architecture)
- [Tech Stack](#tech-stack)
- [Database Design](#database-design)
- [API Endpoints](#api-endpoints)
- [Getting Started](#getting-started)
- [Azure Deployment](#azure-deployment)

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
  seed/main.go             → Database seeder (admin account + 20 sample events)

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

client/                    → React frontend (Vite + TypeScript + Tailwind CSS)
db/migrations/             → 9 versioned SQL migration files
```

---

## Tech Stack

### Backend

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

### Frontend

| Layer | Technology |
|---|---|
| Framework | React 19 |
| Language | TypeScript |
| Build Tool | Vite |
| Styling | Tailwind CSS |
| HTTP Client | Axios |
| Routing | React Router v7 |

---

## Database Design

![Database Diagram](docs/screenshots/ticresERD.png)

**7 tables** with foreign keys, ENUM types, and proper constraints:

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
git clone https://github.com/ozarabal/TicRes.git
cd ticres
make quick-start
```

This starts **PostgreSQL**, **Redis**, runs **migrations**, and launches the **API** — all in one command.

The API will be available at `http://localhost:8080`.

### Seed Sample Data

To populate the database with an admin account and 20 sample events:

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

### Frontend Development

```bash
cd client
npm install
npm run dev     # Starts on http://localhost:3000
```

### Stop & Cleanup

```bash
make quick-stop     # Stop containers
make quick-clean    # Stop + remove volumes
```

---

## Azure Deployment

This project is deployed on **Microsoft Azure** (Student account, Southeast Asia region).

### Azure Architecture

| Component | Azure Service | Tier |
|-----------|--------------|------|
| Frontend | Azure Static Web Apps | Free |
| Backend API | Azure App Service (Linux) | B1 |
| Database | Azure Database for PostgreSQL Flexible Server | Burstable B1ms |
| Cache | Azure Cache for Redis | Basic C0 |
| Container Registry | Azure Container Registry | Basic |

### Prerequisites

- [Azure CLI](https://docs.microsoft.com/en-us/cli/azure/install-azure-cli) installed
- Azure for Students subscription activated
- Docker installed locally

### 1. Create Azure Resources

```bash
# Login to Azure
az login

# Create Resource Group (Southeast Asia)
az group create --name rg-ticres --location southeastasia

# PostgreSQL Flexible Server
az postgres flexible-server create \
  --resource-group rg-ticres --name ticres-db \
  --location southeastasia --admin-user ticresadmin \
  --admin-password <STRONG_PASSWORD> \
  --sku-name Standard_B1ms --tier Burstable \
  --storage-size 32 --version 16 --yes

az postgres flexible-server firewall-rule create \
  --resource-group rg-ticres --name ticres-db \
  --rule-name AllowAzureServices \
  --start-ip-address 0.0.0.0 --end-ip-address 0.0.0.0

az postgres flexible-server db create \
  --resource-group rg-ticres --server-name ticres-db \
  --database-name ticket_db

# Azure Cache for Redis
az redis create --resource-group rg-ticres --name ticres-cache \
  --location southeastasia --sku Basic --vm-size c0

# Container Registry
az acr create --resource-group rg-ticres --name ticresregistry \
  --sku Basic --location southeastasia --admin-enabled true

# App Service Plan + Web App
az appservice plan create --resource-group rg-ticres \
  --name ticres-plan --location southeastasia --is-linux --sku B1

az webapp create --resource-group rg-ticres --plan ticres-plan \
  --name ticres-api \
  --deployment-container-image-name ticresregistry.azurecr.io/ticres-api:latest
```

### 2. Configure Environment Variables

```bash
az webapp config appsettings set --resource-group rg-ticres --name ticres-api \
  --settings \
    PORT=8080 \
    DB_HOST=ticres-db.postgres.database.azure.com \
    DB_PORT=5432 \
    DB_USER=ticresadmin \
    DB_PASSWORD=<DB_PASSWORD> \
    DB_NAME=ticket_db \
    SSL_MODE=require \
    JWT_SECRET=<STRONG_SECRET> \
    JWT_EXP_TIME=24 \
    CACHE_HOST=ticres-cache.redis.cache.windows.net \
    CACHE_PORT=6380 \
    CACHE_PASSWORD=<REDIS_ACCESS_KEY> \
    CACHE_TLS=true \
    APP_MODE=production \
    RUN_SEED=true \
    WEBSITES_PORT=8080
```

> Set `RUN_SEED=false` after the first successful deployment.

### 3. Deploy Backend

```bash
# Login to ACR
az acr login --name ticresregistry

# Build and push Docker image
docker build -t ticresregistry.azurecr.io/ticres-api:latest .
docker push ticresregistry.azurecr.io/ticres-api:latest

# Restart to pull new image
az webapp restart --resource-group rg-ticres --name ticres-api
```

### 4. Deploy Frontend

```bash
# Create Static Web App (via Azure Portal or CLI)
az staticwebapp create --resource-group rg-ticres \
  --name ticres-frontend --location southeastasia

# Build with production API URL
cd client
npm ci
VITE_API_URL=https://ticres-api.azurewebsites.net/api/v1 npm run build

# Deploy using SWA CLI
npx @azure/static-web-apps-cli deploy ./dist \
  --deployment-token <SWA_DEPLOYMENT_TOKEN>
```

### Environment Variables Reference

| Variable | Description | Local | Azure |
|----------|-------------|-------|-------|
| `PORT` | API server port | 8080 | 8080 |
| `DB_HOST` | PostgreSQL host | localhost | ticres-db.postgres.database.azure.com |
| `DB_PORT` | PostgreSQL port | 5433 | 5432 |
| `DB_USER` | Database user | postgres | ticresadmin |
| `DB_PASSWORD` | Database password | secret | (strong password) |
| `DB_NAME` | Database name | ticket_db | ticket_db |
| `SSL_MODE` | PostgreSQL SSL mode | disable | require |
| `JWT_SECRET` | JWT signing secret | rahasia_negara | (strong secret) |
| `JWT_EXP_TIME` | JWT expiry (hours) | 24 | 24 |
| `CACHE_HOST` | Redis host | localhost | ticres-cache.redis.cache.windows.net |
| `CACHE_PORT` | Redis port | 6379 | 6380 |
| `CACHE_PASSWORD` | Redis password | (empty) | (access key) |
| `CACHE_TLS` | Redis TLS enabled | false | true |
| `RUN_SEED` | Run seed on startup | - | true (first deploy only) |
| `WEBSITES_PORT` | Azure container port | - | 8080 |
