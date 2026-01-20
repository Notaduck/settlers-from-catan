# agent.md — Settlers from Catan (Go backend + React frontend)

You are working in a monorepo:

- `backend/` — Go server + game engine
- `frontend/` — React (Vite) UI + Playwright e2e
- `proto/` — buf-managed protobuf schema
- `Makefile` — canonical dev/test commands

Primary goals:

1. Keep the rules/game engine correct and deterministic.
2. Everything is unit tested (Go + TS where applicable).
3. Critical user flows are end-to-end tested with Playwright.
4. Never break protobuf generation or websocket contract.

---

## Ground rules

### Don’t edit generated files

- Go generated: `backend/gen/proto/**`
- TS generated: `frontend/src/gen/proto/**`
  If proto changes are needed: update `.proto` files under `proto/` and run `make generate`.

### Prefer deterministic behavior

- Any randomness (board gen, dice) must be seedable and testable.
- Unit tests must not depend on timing, sleep, or real network where avoidable.

### Keep code changes small and test-first

- If implementing a rule: add failing tests first (Go).
- If implementing a UI flow: update/add Playwright spec(s). test with make e2e

---

## Commands (canonical)

From repo root:

### Install + generate + build

- `make install`
- `make generate`
- `make build`

### Dev

- `make dev` (backend + frontend)
- Backend: `make dev-backend` (port :8080)
- Frontend: `make dev-frontend` (port :3000)

### Tests

- `make test` (all unit tests)
- `make test-backend`
- `make test-frontend` (currently may be minimal)
- `make e2e` (starts backend/frontend and runs Playwright)

### Lint

- `make lint`
- `make lint-proto`
- `make typecheck`

### Housekeeping

- `make clean`
- `make db-reset`

When you change anything:

- Run `make test` at minimum.
- If UI or websocket behavior changes, also run `make e2e`.

---

## Architecture expectations

### Backend (Go)

Likely relevant areas:

- Game logic: `backend/internal/game/*`
- Websocket hub: `backend/internal/hub/*`
- HTTP handlers: `backend/internal/handlers/*`
- Server entry: `backend/cmd/server/main.go`

Backend responsibilities:

- Own the source-of-truth game state and rules.
- Validate all client commands (never trust the UI).
- Broadcast authoritative updates over websocket.
- Keep state transitions explicit and testable.

Testing guidelines (Go):

### Table-Driven Tests (The Game Changer)

Use **maps** instead of slices for test cases — provides better IDE navigation, clearer names, and randomized iteration order helps catch test dependencies:

```go
func TestPlaceSettlement(t *testing.T) {
    tests := map[string]struct {
        setup     func(*GameState)
        playerID  string
        vertexID  string
        wantErr   bool
        errorMsg  string
    }{
        "valid placement": {
            setup:    setupValidBoard,
            playerID: "p1",
            vertexID: "v1",
            wantErr:  false,
        },
        "distance rule violation": {
            setup:    setupTooClose,
            playerID: "p1",
            vertexID: "v2",
            wantErr:  true,
            errorMsg: "too close to existing settlement",
        },
    }
    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            state := NewGameState(...)
            tt.setup(state)
            err := PlaceSettlement(state, tt.playerID, tt.vertexID)
            
            if tt.wantErr {
                if err == nil {
                    t.Error("expected error but got none")
                    return
                }
                if tt.errorMsg != "" && err.Error() != tt.errorMsg {
                    t.Errorf("expected error %q, got %q", tt.errorMsg, err.Error())
                }
            } else if err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

### Test Structure: Setup → Execute → Assert

```go
func TestUserService_CreateUser(t *testing.T) {
    // Setup
    mockRepo := &MockUserRepository{users: make(map[string]*User)}
    service := &UserService{repo: mockRepo}
    newUser := &User{Name: "John"}
    
    // Execute
    err := service.CreateUser(newUser)
    
    // Assert
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
    if newUser.ID == "" {
        t.Error("expected user ID to be set")
    }
}
```

### Dependency Injection & Mocking

Define interfaces for external dependencies to enable testing:

```go
// Interface for dependency injection
type DiceRoller interface {
    Roll() (int, int)
}

// Production implementation
type RandomDiceRoller struct{}
func (r *RandomDiceRoller) Roll() (int, int) {
    return rand.Intn(6)+1, rand.Intn(6)+1
}

// Test mock - deterministic
type FixedDiceRoller struct {
    Values []int
    Index  int
}
func (r *FixedDiceRoller) Roll() (int, int) {
    d1, d2 := r.Values[r.Index], r.Values[r.Index+1]
    r.Index += 2
    return d1, d2
}
```

### Use `t.Helper()` for Better Error Reporting

```go
func assertNoError(t *testing.T, err error) {
    t.Helper()  // Points error to caller, not this line
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}

func assertError(t *testing.T, err error, wantMsg string) {
    t.Helper()
    if err == nil {
        t.Fatal("expected error but got none")
    }
    if wantMsg != "" && err.Error() != wantMsg {
        t.Errorf("expected error %q, got %q", wantMsg, err.Error())
    }
}
```

### Testing HTTP Handlers

Use `httptest` for handler testing:

```go
func TestGameHandler_CreateGame(t *testing.T) {
    tests := map[string]struct {
        method       string
        body         string
        expectedCode int
        expectedBody string
    }{
        "valid request": {
            method:       "POST",
            body:         `{"playerName":"Alice"}`,
            expectedCode: 200,
        },
        "invalid json": {
            method:       "POST",
            body:         `{invalid}`,
            expectedCode: 400,
        },
    }
    
    for name, tt := range tests {
        t.Run(name, func(t *testing.T) {
            req := httptest.NewRequest(tt.method, "/games", strings.NewReader(tt.body))
            rec := httptest.NewRecorder()
            
            handler.ServeHTTP(rec, req)
            
            if rec.Code != tt.expectedCode {
                t.Errorf("expected status %d, got %d", tt.expectedCode, rec.Code)
            }
        })
    }
}
```

### Testing Concurrent Code

```go
func TestConcurrentResourceAccess(t *testing.T) {
    state := NewGameState(...)
    numGoroutines := 100
    
    var wg sync.WaitGroup
    wg.Add(numGoroutines)
    
    for i := 0; i < numGoroutines; i++ {
        go func() {
            defer wg.Done()
            // Concurrent operation
            state.AddResource("p1", pb.TileResource_TILE_RESOURCE_WOOD, 1)
        }()
    }
    
    wg.Wait()
    
    // Verify final state
    if state.Players[0].Resources.Wood != int32(numGoroutines) {
        t.Errorf("race condition detected")
    }
}
```

### Common Mistakes to Avoid

1. **Testing implementation, not behavior**:
   ```go
   // BAD: Testing internal state
   if cache.internalMap["key"] == "value" { ... }
   
   // GOOD: Testing public behavior
   value, exists := cache.Get("key")
   ```

2. **Not testing error cases** — most production bugs happen in error paths

3. **Using `time.Sleep()` in tests** — use channels/signals instead:
   ```go
   // BAD
   time.Sleep(100 * time.Millisecond)
   
   // GOOD
   select {
   case result := <-ch:
       // handle result
   case <-time.After(100 * time.Millisecond):
       t.Error("timeout")
   }
   ```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with race detector (catches concurrency bugs)
go test -race ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out  # View in browser

# Run benchmarks
go test -bench=. -benchmem ./...

# Skip long-running tests
go test -short ./...
```

### What to Cover

- Board generation invariants (correct tile distribution, valid coords)
- Turn/state machine transitions (lobby → setup → playing)
- Command validation (reject invalid moves, enforce rules)
- Error scenarios (network failures, invalid inputs, resource exhaustion)
- Edge cases (empty inputs, boundary values, max players)
- Deterministic playthrough tests where possible

### Avoid Flaky Tests

- No real sockets unless explicitly testing handlers/hub
- Prefer in-memory constructs
- Use seedable randomness
- No `time.Sleep()` — use channels/signals
- Run with `-race` flag to catch race conditions

### Frontend (React + TS)

Likely relevant areas:

- `frontend/src/context/GameContext.tsx` (client state + websocket)
- `frontend/src/hooks/useWebSocket.ts`
- UI components:
  - `frontend/src/components/Lobby/*`
  - `frontend/src/components/Game/*`
  - `frontend/src/components/Board/*`

Frontend responsibilities:

- Render server-provided state.
- Send user intent as commands.
- Never compute authoritative game logic locally (only UI convenience).

Testing guidelines (Frontend):

- Prefer Playwright for “real user flows”.
- If adding component-level tests later, keep them pure and mock websocket.

---

## Playwright E2E strategy

Location:

- Specs: `frontend/tests/*.spec.ts`
- Config: `frontend/playwright.config.ts`

Run:

- `make e2e` (requires backend/frontend running)
- `cd frontend && npm test`
- `cd frontend && npm run test:headed` (see browser)

E2E principles:

- Test the real flow:
  1. Create game
  2. Join game
  3. Toggle ready
  4. Start game
  5. Board renders correctly and state updates arrive
- Avoid brittle selectors:
  - Prefer `data-cy="..."` attributes in UI.
- Ensure tests are isolated:
  - Each spec should create its own game/lobby.
  - Backend should support easy reset (in-memory store) or unique game IDs.

If tests are flaky:

- Add explicit waits for websocket state changes.
- Use Playwright's built-in auto-waiting and retryability.

---

## Protobuf workflow

- Proto schema lives under: `proto/catan/v1/*`
- Use buf:
  - `make generate` runs `buf generate`
  - `make lint-proto` lints proto files

Rules:

- Any contract change between backend and frontend must go through proto.
- After changing `.proto`:
  - Run `make generate`
  - Update server/client usage accordingly
  - Update unit tests + Playwright tests

---

## Definition of done (DoD)

A change is “done” when:

- `make test` passes locally.
- `make e2e` passes if UI/websocket behavior changed.
- No generated files were hand-edited.
- New behavior has tests:
  - Go rule/state changes => Go unit tests.
  - User-visible flows => Playwright coverage (or updated existing spec).
- The change is deterministic and does not introduce flakiness.

---

## When in doubt

Prefer:

- Correctness > feature speed.
- Explicit state machines and validation.
- Deterministic seeds for randomness.
- Small PR-sized changes with tests.

If something is unclear:

- Inspect existing tests first:
  - `backend/internal/game/*_test.go`
  - `frontend/tests/game-flow.spec.ts`
    and extend patterns already used in the repo.
