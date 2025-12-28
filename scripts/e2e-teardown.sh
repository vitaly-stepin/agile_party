#!/bin/bash
set -e

echo "Stopping E2E environment..."
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR/.."
docker compose -f e2e/docker-compose.e2e.yml down -v

echo "E2E environment stopped and cleaned up!"
