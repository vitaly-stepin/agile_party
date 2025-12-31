#!/bin/bash
set -e

echo "Starting E2E environment..."
cd e2e
docker compose -f docker-compose.e2e.yml up -d

echo "Waiting for services to be ready..."
sleep 10

echo "Checking frontend health..."
curl --retry 10 --retry-delay 2 --retry-connrefused http://localhost:5174 > /dev/null 2>&1

echo "Checking backend health..."
curl --retry 10 --retry-delay 2 --retry-connrefused http://localhost:8081/api/health > /dev/null 2>&1

echo "E2E environment is ready!"
