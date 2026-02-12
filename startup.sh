#!/bin/sh

echo "=== TicRes Startup ==="
echo "DB_HOST: ${DB_HOST}"
echo "DB_PORT: ${DB_PORT}"
echo "DB_USER: ${DB_USER}"
echo "DB_NAME: ${DB_NAME}"
echo "SSL_MODE: ${SSL_MODE}"
echo "CACHE_HOST: ${CACHE_HOST}"
echo "CACHE_PORT: ${CACHE_PORT}"
echo "PORT: ${PORT}"

echo "Running database migrations..."
migrate -path /app/db/migrations \
  -database "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${SSL_MODE}" \
  -verbose up 2>&1 || echo "Migration failed or already up-to-date (exit code: $?)"

if [ "$RUN_SEED" = "true" ]; then
  echo "Running database seed..."
  ./seed 2>&1 || echo "Seed failed (exit code: $?)"
fi

echo "Starting API server..."
exec ./api
