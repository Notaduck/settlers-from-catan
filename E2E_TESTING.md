# End-to-End Testing Guide

This guide explains how to run the Playwright E2E tests for the Settlers from Catan game.

## Quick Start

### 1. Start Backend with DEV_MODE

E2E tests that grant resources require the backend to be running with `DEV_MODE=true`:

```bash
# Terminal 1: Start backend with DEV_MODE enabled
cd backend
DEV_MODE=true go run ./cmd/server
```

### 2. Start Frontend

```bash
# Terminal 2: Start frontend dev server
cd frontend  
npm run dev
```

### 3. Run E2E Tests

```bash
# Terminal 3: Run E2E tests
make e2e

# Or run specific tests
cd frontend
npm test -- --grep "Development Cards"
```

## DEV_MODE Test Endpoints

When `DEV_MODE=true`, the backend exposes additional test endpoints:

- `POST /test/grant-resources` - Grant resources to a player
- `POST /test/force-dice-roll` - Force next dice roll to specific value  
- `POST /test/set-game-state` - Advance game to specific phase

These endpoints allow E2E tests to set up specific game scenarios for testing.

## Test Organization

### Core Test Files

- `tests/game-flow.spec.ts` - Basic game flow (lobby → setup → playing)
- `tests/development-cards.spec.ts` - Development card functionality
- `tests/trading.spec.ts` - Trading system
- `tests/robber-flow.spec.ts` - Robber mechanics

### Test Helpers

- `tests/helpers.ts` - Reusable test utilities
  - `startTwoPlayerGame()` - Sets up 2-player game  
  - `completeSetupPhase()` - Completes setup phase
  - `grantResources()` - Grants resources (DEV_MODE only)
  - `rollDice()`, `endTurn()`, etc. - Game actions

## Test Strategy

### 1. UI-Focused Tests

Tests that don't require resource manipulation run without DEV_MODE:

```typescript
test("should display development cards panel", async ({ page }) => {
  // These tests work with any backend
});
```

### 2. Resource-Dependent Tests  

Tests requiring specific resources check for DEV_MODE availability:

```typescript  
test("should buy dev card with resources", async ({ page, request }) => {
  const devModeEnabled = await isDevModeAvailable(request);
  if (!devModeEnabled) {
    test.skip("DEV_MODE test endpoints not available");
  }
  // Test with granted resources...
});
```

## Troubleshooting

### "Test endpoints not available" Error

This means the backend is not running with `DEV_MODE=true`. Start it with:

```bash
DEV_MODE=true go run ./cmd/server
```

### Timeout Issues

E2E tests have been configured with longer timeouts (60s total, 10s for actions). If tests still timeout:

1. Check that both backend (port 8080) and frontend (port 3000) are running
2. Ensure WebSocket connections are stable
3. Increase timeouts in `playwright.config.ts` if needed

### Resource State Debugging

If resource-related tests fail, check the browser console or test output for resource state information.

## Running Tests in CI/CD

For automated testing environments, ensure:

1. Backend starts with `DEV_MODE=true`
2. Both servers are fully started before running tests
3. Use `make e2e` which includes proper startup checks

Example CI script:

```bash
# Start backend with DEV_MODE
DEV_MODE=true go run ./cmd/server &
BACKEND_PID=$!

# Start frontend  
cd frontend && npm run dev &
FRONTEND_PID=$!

# Wait for services to start
sleep 5

# Run tests
make e2e

# Cleanup
kill $BACKEND_PID $FRONTEND_PID
```