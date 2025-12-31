.PHONY: help up down restart logs clean test shell-db migrate e2e-setup e2e-test e2e-teardown

help:
	@echo "Available commands:"
	@echo "  make up              - Start all services"
	@echo "  make down            - Stop all services"
	@echo "  make restart         - Restart all services"
	@echo "  make logs            - Show logs from all services"
	@echo "  make clean           - Stop services and remove volumes"
	@echo "  make test            - Run all Go tests with coverage and summary"
	@echo "  make shell-db        - Open PostgreSQL shell"
	@echo "  make migrate         - Run database migrations"
	@echo "  make e2e-setup       - Start E2E test environment"
	@echo "  make e2e-test        - Run E2E tests"
	@echo "  make e2e-teardown    - Stop E2E test environment"

up:
	docker compose up -d

down:
	docker compose down

restart:
	docker compose restart

logs:
	docker compose logs -f

clean:
	docker compose down -v

test:
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Running Go tests with coverage..."
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@docker compose exec backend sh -c 'go test -coverprofile=/tmp/coverage.out -covermode=atomic ./... 2>&1 | tee /tmp/test-output.txt'
	@echo ""
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@echo "Test Summary:"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
	@docker compose exec backend sh -c "grep -E '^(ok|PASS|FAIL)' /tmp/test-output.txt"
	@echo ""
	@echo "Total Coverage:"
	@docker compose exec backend sh -c "go tool cover -func=/tmp/coverage.out | grep total" || echo "No coverage data available"
	@echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

shell-db:
	docker compose exec postgres psql -U postgres -d agile_party

migrate:
	@echo "Migrations run automatically on postgres container startup"
	@echo "Check: ./backend/internal/adapters/postgres/migrations"

e2e-setup:
	./scripts/e2e-setup.sh

e2e-test:
	cd e2e && npm test

e2e-teardown:
	./scripts/e2e-teardown.sh
