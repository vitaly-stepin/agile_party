# Agile Party - A Scrum Poker tool to plan faster and have more fun doing it 🚀

A real-time Scrum Poker estimation tool built with Go (backend) and React (frontend).

## Project Structure

```
agile_party/
├── backend/          # Go backend with Hexagonal Architecture
│   ├── cmd/api/      # Application entry point
│   ├── internal/     # Private application code
│   │   ├── domain/   # Core business logic
│   │   ├── application/  # Use cases
│   │   ├── adapters/     # External system adapters
│   │   └── interfaces/   # HTTP/WebSocket handlers
│   └── pkg/          # Public utilities
├── frontend/         # React + TypeScript + Tailwind CSS
└── docker-compose.yml
```

## Technology Stack

### Backend
- **Language**: Go 1.25+
- **Framework**: Fiber (HTTP + WebSocket)
- **Database**: PostgreSQL 16
- **Architecture**: Hexagonal (Ports & Adapters)

### Frontend
- **Framework**: React 18 + TypeScript
- **Build Tool**: Vite
- **Styling**: Tailwind CSS
- **Routing**: React Router

## Quick Start

### Prerequisites
- Docker & Docker Compose

### Running Locally

1. **Clone the repository**
   ```bash
   git clone github.com/vitaly-stepin/agile_party
   cd agile_party
   ```

2. **Copy environment variables**
   ```bash
   cp .env.example .env
   ```

3. **Start all services**
   ```bash
   docker-compose up
   ```

4. **Access the application**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - Health Check: http://localhost:8080/api/health

### Development

The setup includes hot reload for both backend and frontend:
- **Backend**: Air watches Go files and rebuilds automatically
- **Frontend**: Vite dev server with Hot Module Replacement (HMR)


## Testing

### Backend Tests

Run all backend tests with coverage:

```bash
make test
```

Or manually:

```bash
cd backend
go test -v -race -coverprofile=coverage.out ./...
```

### E2E Tests

End-to-end tests run in an isolated environment with dedicated services on different ports (5174, 8081, 5433).

**Quick Start:**

```bash
# Start E2E environment
./scripts/e2e-setup.sh

# Run tests
cd e2e && npm test

# Cleanup
./scripts/e2e-teardown.sh
```

**Using Makefile:**

```bash
make e2e-setup    # Start E2E services
make e2e-test     # Run E2E tests
make e2e-teardown # Stop E2E services
```

For detailed documentation on writing tests, debugging, and CI/CD integration, see [e2e/README.md](e2e/README.md).

## License

MIT
