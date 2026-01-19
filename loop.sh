#!/bin/bash
# Ralph loop - autonomous coding agent
# DO NOT use set -e, we want the loop to continue even if commands fail

MODE="${1:-build}"
MAX_ITERATIONS="${2:-0}"
ITERATION=0

if [ "$MODE" = "plan" ]; then
    PROMPT_FILE="PROMPT_plan.md"
else
    PROMPT_FILE="PROMPT_build.md"
fi

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🔄 Ralph Loop Starting"
echo "   Mode: $MODE"
echo "   Prompt: $PROMPT_FILE"
echo "   Max iterations: ${MAX_ITERATIONS:-unlimited}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

while true; do
    if [ "$MAX_ITERATIONS" -gt 0 ] && [ "$ITERATION" -ge "$MAX_ITERATIONS" ]; then
        echo "✅ Reached max iterations: $MAX_ITERATIONS"
        break
    fi

    ITERATION=$((ITERATION + 1))
    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🚀 ITERATION $ITERATION"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Run Codex iteration
    # --dangerously-bypass-approvals-and-sandbox: full access, no prompts
    # || true: continue loop even if codex exits with error
    codex --dangerously-bypass-approvals-and-sandbox "$(cat "$PROMPT_FILE")" || {
        echo "⚠️  Codex exited with error, continuing to next iteration..."
        sleep 2
    }

    # Push after each iteration (ignore failures)
    git push origin "$(git branch --show-current)" 2>/dev/null || echo "⚠️  Git push failed (continuing anyway)"

    echo ""
    echo "✓ Iteration $ITERATION complete"
    
    # Small delay to allow filesystem to settle
    sleep 1
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "🏁 Ralph Loop Finished"
echo "   Total iterations: $ITERATION"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"