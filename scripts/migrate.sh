#!/bin/bash
# Run migrations for AIAggregator

set -e

DB_URL="${DATABASE_URL:-postgres://postgres:postgres@localhost:5432/aiaggregator?sslmode=disable}"
MIGRATIONS_DIR="infrastructure/postgres/migrations"

echo "Running migrations..."

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo "Installing golang-migrate..."
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
fi

# Run migrations
migrate -path "$MIGRATIONS_DIR" -database "$DB_URL" up

echo "Migrations completed successfully!"
