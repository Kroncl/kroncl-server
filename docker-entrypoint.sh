#!/bin/sh
set -e

echo "PostgreSQL is ready, running migrations..."
./migrate -cmd up -type system

echo "Starting API..."
exec ./main