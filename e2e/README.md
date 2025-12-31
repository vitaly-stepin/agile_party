# E2E Testing with Playwright

This directory contains end-to-end tests for Agile Party using Playwright. Tests run against an isolated environment with dedicated services on different ports to avoid interference with local development.

## Prerequisites

- Docker and Docker Compose
- Node.js 20+
- npm

## Architecture

### Isolated E2E Environment

E2E tests run in a completely separate environment from development:

| Service | Dev Port | E2E Port |
|---------|----------|----------|
| Frontend | 5173 | 5174 |
| Backend | 8080 | 8081 |
| PostgreSQL | 5432 | 5433 |

This allows running tests without stopping your development environment.

### Key Design Decisions

- **Sequential Execution**: Tests run with `workers: 1` due to shared database
- **Database Cleanup**: Automatic cleanup before/after each test via fixture
- **Page Object Model**: Reusable abstractions for UI interactions
- **WebSocket Testing**: DOM-based assertions with generous timeouts for real-time sync

## Quick Start

### 1. Start E2E Environment

From project root:

```bash
./scripts/e2e-setup.sh
```

Or manually:

```bash
cd e2e
docker compose -f docker-compose.e2e.yml up -d
```

Wait for services to be healthy (~15 seconds).

### 2. Run Tests

```bash
cd e2e
npm test
```

### 3. Stop E2E Environment

From project root:

```bash
./scripts/e2e-teardown.sh
```

Or manually:

```bash
cd e2e
docker compose -f docker-compose.e2e.yml down -v
```

## Available Commands

```bash
npm test              # Run all tests (headless)
npm run test:headed   # Run with browser visible
npm run test:debug    # Run in debug mode
npm run test:ui       # Open Playwright UI mode
npm run report        # Show last test report
npm run codegen       # Generate test code via recording
```

## Writing Tests

### Using Page Objects

Tests use Page Object Model for maintainability:

```typescript
import { test, expect } from '../fixtures/database.fixture';
import { HomePage } from '../page-objects/HomePage';
import { RoomPage } from '../page-objects/RoomPage';

test('should create and join room', async ({ page, cleanDatabase }) => {
  const homePage = new HomePage(page);
  const roomPage = new RoomPage(page);

  await homePage.goto();
  await homePage.createRoom('Test Room', 'Alice');

  const roomId = await roomPage.getRoomId();
  expect(roomId).toBeTruthy();

  await roomPage.waitForWebSocketConnection();
  expect(await roomPage.getUserInList('Alice')).toBeVisible();
});
```

### Database Cleanup Fixture

All tests automatically clean the database before and after execution:

```typescript
import { test, expect } from '../fixtures/database.fixture';

test('my test', async ({ page, cleanDatabase }) => {
  // Database is clean at start
  // Test code here
  // Database will be cleaned after test
});
```

The fixture deletes data from `tasks` and `rooms` tables in the correct order to maintain referential integrity.

### WebSocket Testing Best Practices

WebSocket connections require special handling:

1. **Wait for Connection**: Always verify connection status before assertions

```typescript
await roomPage.waitForWebSocketConnection();
```

2. **Generous Timeouts**: Use longer timeouts for real-time updates

```typescript
await expect(roomPage.getUserInList('Bob')).toBeVisible({ timeout: 10000 });
```

3. **Multi-User Testing**: Use separate browser contexts

```typescript
test('multi-user scenario', async ({ page, browser }) => {
  // First user
  const homePage1 = new HomePage(page);
  await homePage1.createRoom('Room', 'Alice');

  // Second user in new context
  const context2 = await browser.newContext();
  const page2 = await context2.newPage();
  const homePage2 = new HomePage(page2);
  await homePage2.joinRoom(roomId, 'Bob');

  // Verify real-time sync
  const roomPage1 = new RoomPage(page);
  await expect(roomPage1.getUserInList('Bob')).toBeVisible({ timeout: 10000 });

  await context2.close();
});
```

4. **DOM-Based Verification**: Assert on UI changes, not WebSocket messages

```typescript
// Good: Wait for UI to reflect the change
await expect(page.getByText('Bob')).toBeVisible();

// Avoid: Intercepting WebSocket frames (brittle)
```

### Adding New Page Objects

Create new page objects in `page-objects/` directory:

```typescript
import { Page, Locator } from '@playwright/test';

export class MyPage {
  readonly page: Page;
  readonly myButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.myButton = page.getByRole('button', { name: 'Click me' });
  }

  async performAction() {
    await this.myButton.click();
  }
}
```

Use accessible selectors (getByRole, getByLabel, getByText) instead of CSS selectors for resilience.

## Debugging

### Headed Mode

See the browser while tests run:

```bash
npm run test:headed
```

### Debug Mode

Step through tests with Playwright Inspector:

```bash
npm run test:debug
```

### UI Mode

Interactive mode with time-travel debugging:

```bash
npm run test:ui
```

### Trace Viewer

Traces are automatically captured on test failures. View them:

```bash
npx playwright show-trace playwright-report/data/<trace-file>.zip
```

### Console Logs

Add debugging output:

```typescript
console.log('Current URL:', page.url());
console.log('Room ID:', await roomPage.getRoomId());
```

### Pause Execution

Add breakpoints in tests:

```typescript
await page.pause(); // Opens Playwright Inspector
```

## CI/CD Integration

Tests run automatically on:
- Pull requests to main
- Pushes to main

GitHub Actions workflow: `.github/workflows/e2e-tests.yml`

### Viewing CI Results

1. Navigate to Actions tab in GitHub
2. Select the workflow run
3. Download "playwright-report" artifact on failure
4. Extract and open `index.html` locally

### CI-Specific Behavior

- Tests retry up to 2 times on failure
- Screenshots, videos, and traces captured on failure only
- Test results exported in JUnit format
- 20-minute timeout for entire workflow

## Project Structure

```
e2e/
├── docker-compose.e2e.yml    # Isolated E2E services
├── playwright.config.ts       # Playwright configuration
├── package.json               # E2E dependencies
├── tsconfig.json              # TypeScript config
├── fixtures/                  # Test fixtures
│   └── database.fixture.ts    # Database cleanup
├── page-objects/              # Page Object Models
│   ├── HomePage.ts
│   └── RoomPage.ts
├── tests/                     # Test files
│   └── critical-flow.spec.ts
└── utils/                     # Helper utilities (future)
```

## Configuration

### playwright.config.ts

Key settings:
- Base URL: `http://localhost:5174`
- Test timeout: 60s (WebSocket setup is slow)
- Action timeout: 15s
- Navigation timeout: 30s
- Workers: 1 (sequential execution)
- Retries: 2 in CI, 0 locally
- Browsers: Chromium only (fast feedback)

### docker-compose.e2e.yml

Dedicated services:
- `postgres_e2e`: Isolated database on port 5433
- `backend_e2e`: API on port 8081
- `frontend_e2e`: UI on port 5174
- Separate network: `agile_party_e2e`
- Health checks ensure services ready before tests

## Troubleshooting

### Tests Fail Locally But Pass in CI

- Check Docker service health: `docker compose -f e2e/docker-compose.e2e.yml ps`
- Verify no port conflicts: `lsof -i :5174 -i :8081 -i :5433`
- Ensure services fully started: Wait 15-20 seconds after `up -d`

### Database Connection Errors

```
Error: connect ECONNREFUSED 127.0.0.1:5433
```

**Solution**: Database not ready. Run health check:

```bash
docker compose -f e2e/docker-compose.e2e.yml exec postgres_e2e pg_isready
```

### WebSocket Connection Timeouts

```
TimeoutError: Waiting for connection status failed
```

**Solutions**:
- Increase timeout in test: `{ timeout: 15000 }`
- Check backend logs: `docker compose -f e2e/docker-compose.e2e.yml logs backend_e2e`
- Verify CORS configuration includes E2E port (5174)

### Flaky Tests

- Add explicit waits for WebSocket events: `await page.waitForTimeout(1000)`
- Use stricter locators: Prefer `getByRole` over `getByTestId`
- Increase action timeouts for slow operations
- Check for race conditions in multi-user scenarios

### Port Already in Use

```
Error: bind: address already in use
```

**Solution**: Stop conflicting services:

```bash
./scripts/e2e-teardown.sh
# Or check what's using the port
lsof -i :5174
```

## Extending Tests

### Adding Voting Tests

1. Create `VotingPage.ts` page object:

```typescript
export class VotingPage {
  async submitVote(value: string) {
    await this.page.getByRole('button', { name: value }).click();
  }

  async revealVotes() {
    await this.page.getByRole('button', { name: 'Reveal' }).click();
  }
}
```

2. Write test in `tests/voting-flow.spec.ts`:

```typescript
test('should submit and reveal votes', async ({ page, cleanDatabase }) => {
  // Setup room with multiple users
  // Submit votes
  // Reveal and verify average
});
```

### Adding Task Management Tests

Similar pattern - create `TaskPage.ts` page object and test CRUD operations.

### Adding Error Scenarios

Test edge cases:
- Invalid room ID
- Disconnection handling
- Concurrent modifications
- Input validation

## Performance

### Test Execution Time

Current critical flow test: ~15-20 seconds
- Service startup: ~10s (one-time, reused across tests)
- Test execution: ~5-10s per test

### Optimization Tips

- Run only changed tests: `npx playwright test --grep "pattern"`
- Use `test.skip()` to temporarily disable slow tests
- Consider parallel execution once database isolation solved
- Add more browsers only when needed (Firefox, WebKit)

## Best Practices

1. **Test Real User Flows**: Focus on end-to-end scenarios, not unit-level interactions
2. **Avoid Implementation Details**: Test behavior, not internals
3. **Deterministic Tests**: Use database cleanup, avoid random data
4. **Accessible Selectors**: Use semantic queries (role, label, text)
5. **One Assertion Per Test**: Keep tests focused and debuggable
6. **Descriptive Names**: `test('should allow user to vote and see results')` not `test('voting')`
7. **Clean Up Contexts**: Close extra browser contexts to prevent leaks
8. **Avoid Hardcoded Timeouts**: Use `waitFor*` methods with explicit conditions

## Resources

- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Playwright Best Practices](https://playwright.dev/docs/best-practices)
- [Trace Viewer Guide](https://playwright.dev/docs/trace-viewer)
- [Debugging Guide](https://playwright.dev/docs/debug)
- [Page Object Model](https://playwright.dev/docs/pom)

## Support

For issues or questions:
1. Check this README and troubleshooting section
2. Review Playwright documentation
3. Examine test traces for failures
4. Open an issue with reproduction steps and trace file
