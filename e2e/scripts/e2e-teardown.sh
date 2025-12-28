#!/bin/bash
set -e

echo "Stopping E2E environment..."
cd e2e
docker compose -f docker-compose.e2e.yml down -v

echo "E2E environment stopped!"
