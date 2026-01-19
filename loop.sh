#!/bin/bash
set -euo pipefail

# Usage: ./loop.sh [plan|build] [max_iterations]

MODE="${1:-build}"
MAX_ITERATIONS="${2:-0}"
ITERATION=0

if [ "$MODE" = "plan" ]; then
    PROMPT_FILE="PROMPT_plan.md"
else
    PROMPT_FILE="PROMPT_build.md"
fi

echo "Mode: $MODE | Prompt: $PROMPT_FILE"

while true; do
    if [ "$MAX_ITERATIONS" -gt 0 ] && [ "$ITERATION" -ge "$MAX_ITERATIONS" ]; then
        echo "Reached max iterations: $MAX_ITERATIONS"
        break
    fi

    # Run Codex iteration
    # --dangerously-bypass-approvals-and-sandbox: full access, no prompts
    # Required for git operations (.git write access)
    # WARNING: Only run in isolated/trusted environments
    codex --dangerously-bypass-approvals-and-sandbox "$(cat "$PROMPT_FILE")"

    # Push after each iteration (codex should have committed)
    git push origin "$(git branch --show-current)" 2>/dev/null || true

    ITERATION=$((ITERATION + 1))
    echo -e "\n=== LOOP $ITERATION COMPLETE ===\n"
done