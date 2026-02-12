# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/api cmd/api/main.go
RUN CGO_ENABLED=0 go build -o /app/seed cmd/seed/main.go

# Run stage
FROM alpine:3.21
RUN apk --no-cache add ca-certificates curl

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/seed .
COPY --from=builder /app/db/migrations ./db/migrations
COPY --from=builder /app/docs ./docs

# Install golang-migrate for running migrations on startup
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate

COPY startup.sh .
RUN chmod +x startup.sh

EXPOSE 8080

CMD ["./startup.sh"]
