#!/bin/bash

# Script to sync seats from external service to local database for event ID=1
# Uses the same environment variables as docker-compose for database connection

echo "Setting up environment variables..."

# Database configuration (matching docker-compose)
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=bulbul
export DB_PASSWORD=bulbul123
export DB_NAME=bulbul
export DB_SSLMODE=disable

# External service configuration (you may need to adjust these)
export TICKETING_BASE_URL="https://hub.hackload.kz/event-provider/common"
export TICKETING_TIMEOUT=${TICKETING_TIMEOUT:-"30s"}

# Logging configuration
export LOG_LEVEL=${LOG_LEVEL:-"info"}

echo "Environment variables set:"
echo "  DB_HOST=$DB_HOST"
echo "  DB_PORT=$DB_PORT" 
echo "  DB_USER=$DB_USER"
echo "  DB_NAME=$DB_NAME"
echo "  TICKETING_BASE_URL=$TICKETING_BASE_URL"

# Check if PostgreSQL is running in docker-compose
echo "Checking if PostgreSQL is accessible..."
if ! pg_isready -h $DB_HOST -p $DB_PORT -U $DB_USER > /dev/null 2>&1; then
    echo "WARNING: PostgreSQL is not accessible at $DB_HOST:$DB_PORT"
    echo "Make sure docker-compose is running with: docker-compose up postgres"
    echo "Continuing anyway - the sync command will fail if DB is not available."
fi

echo "Building sync-seats command..."
go run ./cmd/sync-seats -event-id=1

# echo "Running seat synchronization for event ID=1..."
# ./bin/sync-seats -event-id=1

echo "Seat synchronization completed!"