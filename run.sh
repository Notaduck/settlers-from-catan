#!/usr/bin/env bash
set -euo pipefail

echo "==> Step 1: Planning (3 iterations)"
./loop.sh plan 3

echo
echo "==> Step 2: Reviewing plan"
cat IMPLEMENTATION_PLAN.md

echo
read -p "Press ENTER to continue to build, or Ctrl+C to abort..."

echo
echo "==> Step 3: Building (20 iterations)"
./loop.sh build 20

echo
echo "==> Done"
