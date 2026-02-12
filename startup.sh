#!/bin/sh
set -e

echo "Running database migrations..."
migrate -path /app/db/migrations \
  -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${SSL_MODE}" \
  -verbose up

if [ "$RUN_SEED" = "true" ]; then
  echo "Running database seed..."
  ./seed
fi

echo "Starting API server..."
exec ./api
