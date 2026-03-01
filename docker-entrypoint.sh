#!/bin/sh
set -e

echo "Waiting for PostgreSQL..."
./wait-for-it.sh postgres:5432 -t 30

echo "Running migrations..."
./migrate -cmd up -type system

echo "Starting API..."
exec ./main