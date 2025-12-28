#!/bin/bash
set -e

echo "Starting E2E environment..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."
docker compose -f e2e/docker-compose.e2e.yml up -d

echo "Waiting for services to start..."
sleep 10

echo "Checking frontend health..."
curl --retry 10 --retry-delay 2 --retry-connrefused http://localhost:5174 || true

echo "Checking backend health..."
curl --retry 10 --retry-delay 2 --retry-connrefused http://localhost:8081/api/health || true

echo "E2E environment is ready!"
