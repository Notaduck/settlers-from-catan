#!/bin/bash
# Ralph Loop Script - Supports multiple AI agents with full validation
# Usage: ./loop.sh [plan|build] [max_iterations] [agent] [model]
# Agents: codex (default), opencode, claude
# Models (optional): opencode supports -m provider/model

MODE="${1:-build}"
MAX_ITERATIONS="${2:-0}"
AGENT="${3:-codex}"
MODEL="${4:-}"
ITERATION=0
CONSECUTIVE_FAILURES=0
MAX_CONSECUTIVE_FAILURES=3

# Server PIDs
BACKEND_PID=""
FRONTEND_PID=""

if [ "$MODE" = "plan" ]; then
    PROMPT_FILE="PROMPT_plan.md"
else
    PROMPT_FILE="PROMPT_build.md"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔄 Ralph Loop Starting"
echo "   Mode: $MODE"
echo "   Prompt: $PROMPT_FILE"
echo "   Max iterations: $MAX_ITERATIONS (0 = unlimited)"
echo "   Agent: $AGENT"
[ -n "$MODEL" ] && echo "   Model: $MODEL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

# ═══════════════════════════════════════════════════════════════════
# Server Management
# ═══════════════════════════════════════════════════════════════════

start_backend() {
    if curl -s http://localhost:8080/health > /dev/null 2>&1; then
        echo "  ✓ Backend already running"
        return 0
    fi
    
    echo "  🚀 Starting backend server..."
    cd backend && go run cmd/server/main.go > /tmp/ralph-backend.log 2>&1 &
    BACKEND_PID=$!
    cd ..
    
    # Wait for backend to be ready
    for i in {1..30}; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            echo "  ✓ Backend started (PID: $BACKEND_PID)"
            return 0
        fi
        sleep 1
    done
    
    echo "  ❌ Backend failed to start"
    cat /tmp/ralph-backend.log
    return 1
}

start_frontend() {
    if curl -s http://localhost:3000 > /dev/null 2>&1; then
        echo "  ✓ Frontend already running"
        return 0
    fi
    
    echo "  🚀 Starting frontend server..."
    cd frontend && npm run dev > /tmp/ralph-frontend.log 2>&1 &
    FRONTEND_PID=$!
    cd ..
    
    # Wait for frontend to be ready
    for i in {1..30}; do
        if curl -s http://localhost:3000 > /dev/null 2>&1; then
            echo "  ✓ Frontend started (PID: $FRONTEND_PID)"
            return 0
        fi
        sleep 1
    done
    
    echo "  ❌ Frontend failed to start"
    cat /tmp/ralph-frontend.log
    return 1
}

start_servers() {
    echo "📡 Starting servers for E2E tests..."
    start_backend || return 1
    start_frontend || return 1
    echo "  ✓ All servers running"
    return 0
}

stop_servers() {
    echo "🛑 Stopping servers..."
    
    if [ -n "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null && echo "  ✓ Backend stopped"
    fi
    
    if [ -n "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null && echo "  ✓ Frontend stopped"
    fi
    
    # Also kill any orphaned processes
    pkill -f "go run cmd/server/main.go" 2>/dev/null || true
    pkill -f "vite.*3000" 2>/dev/null || true
}

# Cleanup on exit
trap stop_servers EXIT

# ═══════════════════════════════════════════════════════════════════
# Validation Functions
# ═══════════════════════════════════════════════════════════════════

validate_backend() {
    echo "  → Backend build..."
    if ! make build-backend > /tmp/ralph-build.log 2>&1; then
        echo "  ❌ Backend build failed:"
        tail -20 /tmp/ralph-build.log
        return 1
    fi
    echo "  ✅ Backend builds"
    
    echo "  → Backend tests..."
    if ! make test-backend > /tmp/ralph-test.log 2>&1; then
        echo "  ❌ Backend tests failed:"
        tail -30 /tmp/ralph-test.log
        return 1
    fi
    echo "  ✅ Backend tests pass"
    return 0
}

validate_frontend() {
    echo "  → TypeScript typecheck..."
    if ! make typecheck > /tmp/ralph-typecheck.log 2>&1; then
        echo "  ❌ TypeScript failed:"
        tail -20 /tmp/ralph-typecheck.log
        return 1
    fi
    echo "  ✅ TypeScript passes"
    return 0
}

validate_e2e() {
    echo "  → E2E tests..."
    
    # Ensure servers are running
    if ! start_servers; then
        echo "  ❌ Could not start servers for E2E"
        return 1
    fi
    
    # Run E2E tests
    if ! make e2e > /tmp/ralph-e2e.log 2>&1; then
        echo "  ❌ E2E tests failed:"
        tail -40 /tmp/ralph-e2e.log
        return 1
    fi
    echo "  ✅ E2E tests pass"
    return 0
}

# Full validation
validate_full() {
    echo ""
    echo "🔍 Running full validation..."
    
    validate_backend || return 1
    validate_frontend || return 1
    validate_e2e || return 1
    
    echo ""
    echo "✅ Full validation PASSED"
    return 0
}

# Quick validation (no E2E)
validate_quick() {
    echo ""
    echo "🔍 Running quick validation (no E2E)..."
    
    validate_backend || return 1
    validate_frontend || return 1
    
    echo ""
    echo "✅ Quick validation PASSED"
    return 0
}

# Check if changes touch frontend
changes_touch_frontend() {
    # Check staged and unstaged changes
    if git diff --name-only HEAD 2>/dev/null | grep -qE '^frontend/src/'; then
        return 0
    fi
    if git diff --name-only 2>/dev/null | grep -qE '^frontend/src/'; then
        return 0
    fi
    return 1
}

# ═══════════════════════════════════════════════════════════════════
# E2E Audit (every 10 iterations)
# ═══════════════════════════════════════════════════════════════════

run_e2e_audit() {
    echo ""
    echo "📊 Running full E2E audit (iteration $ITERATION)..."
    
    # Ensure servers are running
    if ! start_servers; then
        echo "  ❌ Could not start servers for E2E audit"
        return 1
    fi
    
    # Run all E2E tests and capture results
    cd frontend
    npx playwright test --reporter=json 2>&1 | tee /tmp/ralph-e2e-audit.json || true
    cd ..
    
    # Parse results and update E2E_STATUS.md
    echo "  📝 Updating E2E_STATUS.md..."
    
    # Count results from each spec
    local timestamp=$(date '+%Y-%m-%d %H:%M')
    local audit_content="# E2E Test Status\n\nLast full audit: $timestamp (Iteration $ITERATION)\n\n## Summary\n\n"
    
    # Run each spec individually to get counts
    cd frontend
    for spec in game-flow interactive-board setup-phase robber trading development-cards longest-road ports victory; do
        local result=$(npx playwright test "${spec}.spec.ts" --reporter=list 2>&1 || true)
        local passed=$(echo "$result" | grep -c "✓\|passed" || echo "0")
        local failed=$(echo "$result" | grep -c "✘\|failed\|Error" || echo "0")
        echo "  ${spec}: ${passed} passed, ${failed} failed"
    done
    cd ..
    
    echo "  ✅ E2E audit complete — check E2E_STATUS.md"
    return 0
}

# ═══════════════════════════════════════════════════════════════════
# Status Tracking
# ═══════════════════════════════════════════════════════════════════

update_status() {
    local build_ok=$1
    local e2e_ok=$2
    local msg=$3
    
    cat > .ralph-status.json << EOF
{
    "iteration": $ITERATION,
    "buildPassing": $build_ok,
    "e2ePassing": $e2e_ok,
    "consecutiveFailures": $CONSECUTIVE_FAILURES,
    "message": "$msg",
    "timestamp": "$(date -u +%Y-%m-%dT%H:%M:%SZ)"
}
EOF
}

# ═══════════════════════════════════════════════════════════════════
# Agent Runners
# ═══════════════════════════════════════════════════════════════════

run_codex() {
    codex exec \
        --dangerously-bypass-approvals-and-sandbox \
        "$(cat "$PROMPT_FILE")"
}

run_opencode() {
    if [ -n "$MODEL" ]; then
        opencode run -m "$MODEL" "$(cat "$PROMPT_FILE")"
    else
        opencode run -m "github-copilot/claude-sonnet-4" "$(cat "$PROMPT_FILE")"
    fi
}

run_claude() {
    claude -p --dangerously-skip-permissions "$(cat "$PROMPT_FILE")"
}

run_agent() {
    case "$AGENT" in
        codex)
            run_codex
            ;;
        opencode)
            run_opencode
            ;;
        claude)
            run_claude
            ;;
        *)
            echo "❌ Unknown agent: $AGENT"
            echo "   Supported: codex, opencode, claude"
            exit 1
            ;;
    esac
}

# ═══════════════════════════════════════════════════════════════════
# Main Loop
# ═══════════════════════════════════════════════════════════════════

while true; do
    # Check iteration limit
    if [ "$MAX_ITERATIONS" -gt 0 ] && [ "$ITERATION" -ge "$MAX_ITERATIONS" ]; then
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "✅ Reached max iterations: $MAX_ITERATIONS"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        break
    fi

    # Check for too many failures
    if [ "$CONSECUTIVE_FAILURES" -ge "$MAX_CONSECUTIVE_FAILURES" ]; then
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "🛑 STOPPED: $MAX_CONSECUTIVE_FAILURES consecutive failures"
        echo ""
        echo "   Debug commands:"
        echo "   $ make test-backend   # Check Go tests"
        echo "   $ make typecheck      # Check TypeScript"
        echo "   $ make e2e            # Check E2E (needs servers)"
        echo ""
        echo "   View logs:"
        echo "   $ cat /tmp/ralph-build.log"
        echo "   $ cat /tmp/ralph-test.log"
        echo "   $ cat /tmp/ralph-e2e.log"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        update_status "false" "false" "Stopped after $MAX_CONSECUTIVE_FAILURES failures"
        exit 1
    fi

    ITERATION=$((ITERATION + 1))
    
    # Save iteration to file for agent to read
    echo "$ITERATION" > .ralph-iteration
    
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🚀 ITERATION $ITERATION (Agent: $AGENT)"
    [ "$CONSECUTIVE_FAILURES" -gt 0 ] && echo "   ⚠️  Consecutive failures: $CONSECUTIVE_FAILURES"
    
    # Check if this is an audit iteration (every 10)
    if [ $((ITERATION % 10)) -eq 0 ]; then
        echo "   📊 This is an E2E AUDIT iteration"
    fi
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Run the selected agent
    run_agent || true

    # Every 10 iterations, run full E2E audit
    if [ $((ITERATION % 10)) -eq 0 ]; then
        run_e2e_audit || true
    fi

    # Post-validation
    echo ""
    if changes_touch_frontend; then
        echo "📋 Frontend changes detected → Full validation with E2E"
        if validate_full; then
            CONSECUTIVE_FAILURES=0
            update_status "true" "true" "Full validation passed"
        else
            CONSECUTIVE_FAILURES=$((CONSECUTIVE_FAILURES + 1))
            update_status "false" "false" "Validation failed"
        fi
    else
        echo "📋 Backend-only changes → Quick validation"
        if validate_quick; then
            CONSECUTIVE_FAILURES=0
            update_status "true" "skipped" "Quick validation passed"
        else
            CONSECUTIVE_FAILURES=$((CONSECUTIVE_FAILURES + 1))
            update_status "false" "skipped" "Validation failed"
        fi
    fi

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✓ Iteration $ITERATION complete"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    sleep 2
done
