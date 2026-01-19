#!/bin/bash
# filepath: /Users/Daniel_1/projects/personal/settlers_from_catan/loop.sh

# Ralph Loop Script - Supports multiple AI agents
# Usage: ./loop.sh [plan|build] [max_iterations] [agent] [model]
# Agents: codex (default), opencode, claude
# Models (optional): opencode supports -m provider/model

MODE="${1:-build}"
MAX_ITERATIONS="${2:-0}"
AGENT="${3:-codex}"
MODEL="${4:-}"
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
echo "   Max iterations: $MAX_ITERATIONS (0 = unlimited)"
echo "   Agent: $AGENT"
[ -n "$MODEL" ] && echo "   Model: $MODEL"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

run_codex() {
    codex exec \
        --dangerously-bypass-approvals-and-sandbox \
        "$(cat "$PROMPT_FILE")"
}

run_opencode() {
    # OpenCode 'run' with optional model selection
    # Provider is 'github-copilot' (not 'copilot')
    if [ -n "$MODEL" ]; then
        opencode run -m "$MODEL" "$(cat "$PROMPT_FILE")"
    else
        # Default to GitHub Copilot provider
        opencode run -m "github-copilot/gpt-4.1" "$(cat "$PROMPT_FILE")"
    fi
}

run_claude() {
    # Claude Code CLI
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
            echo "Unknown agent: $AGENT"
            echo "Supported agents: codex, opencode, claude"
            exit 1
            ;;
    esac
}

while true; do
    if [ "$MAX_ITERATIONS" -gt 0 ] && [ "$ITERATION" -ge "$MAX_ITERATIONS" ]; then
        echo ""
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        echo "✅ Reached max iterations: $MAX_ITERATIONS"
        echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        break
    fi

    ITERATION=$((ITERATION + 1))
    
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "🚀 ITERATION $ITERATION (Agent: $AGENT)"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Run the selected agent
    run_agent || true

    echo ""
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "✓ Iteration $ITERATION complete"
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo ""

    # Small delay between iterations
    sleep 2
done