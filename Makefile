# --- Configuration ---
DB_NAME=ticket_db
DB_USER=postgres
DB_PASSWORD=secret
DB_HOST=localhost
DB_PORT=5433
SSL_MODE=disable

# Connection String (Format URL Postgres)
DB_URL="postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(SSL_MODE)"

# Folder tempat file .sql disimpan
MIGRATE_PATH=db/migrations

# --- Quick Start (Docker Compose) ---
quick-start:
	docker compose up --build -d

quick-stop:
	docker compose down

quick-clean:
	docker compose down -v

# --- Docker Commands (Optional tapi Recommended) ---
# Menjalankan container Postgres
docker-up:
	docker run --name ticket-postgres -e POSTGRES_USER=$(DB_USER) -e POSTGRES_PASSWORD=$(DB_PASSWORD) -e POSTGRES_DB=$(DB_NAME) -p $(DB_PORT):5432 -d postgres:alpine

docker-upc:
	docker run --name ticket-redis -p 6379:6379 -d redis:alpine
# Mematikan container
docker-down:
	docker stop ticket-postgres && docker rm ticket-postgres

# --- Migration Commands ---
# Cara pakai: make migrate-create name=create_users_table
migrate-create:
	migrate create -ext sql -dir $(MIGRATE_PATH) -seq $(name)

# Menerapkan migrasi (UP)
migrate-up:
	migrate -path $(MIGRATE_PATH) -database $(DB_URL) -verbose up

# Membatalkan migrasi terakhir (DOWN)
migrate-down:
	migrate -path $(MIGRATE_PATH) -database $(DB_URL) -verbose down 1

# Membatalkan SEMUA migrasi (Reset DB)
migrate-reset:
	migrate -path $(MIGRATE_PATH) -database $(DB_URL) -verbose down

# Memaksa version (Jika migrasi error/dirty state)
migrate-force:
	migrate -path $(MIGRATE_PATH) -database $(DB_URL) force $(version)

# --- Application Commands ---
# Menjalankan aplikasi
run:
	go run cmd/api/main.go

# Seed database dengan admin account dan sample events
seed:
	go run cmd/seed/main.go

build:
	go build ./...

# Membersihkan file binary/cache
clean:
	go clean
	rm -f bin/api

swagger:
	swag init -g cmd/api/main.go -o docs