#!/usr/bin/env bash
# Build Terrasentry, then scan every scenario under test-projects/.
# Usage: ./test-projects/run.sh
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

echo "Building terrasentry..."
go build -o terrasentry ./cmd/terrasentry
echo "Build OK."

for dir in test-projects/*/; do
  [ -f "${dir}plan.json" ] || continue
  echo
  echo "############################################################"
  echo "# ${dir}"
  echo "############################################################"
  ./terrasentry scan --plan "${dir}plan.json" || true
done
