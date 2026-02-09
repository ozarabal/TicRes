# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app/api cmd/api/main.go

# Run stage
FROM alpine:3.21

WORKDIR /app

COPY --from=builder /app/api .
COPY --from=builder /app/db/migrations ./db/migrations
COPY --from=builder /app/docs ./docs

EXPOSE 8080

CMD ["./api"]
