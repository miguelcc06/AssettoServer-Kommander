#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$ROOT"

mkdir -p \
  backend/runtime/data/content/cars \
  backend/runtime/data/content/tracks \
  backend/runtime/data/logs \
  backend/runtime/data/cfg

docker compose up -d --build

PORT="${KOMMANDER_PORT:-8086}"
HEALTH_URL="http://127.0.0.1:${PORT}/api/health"

echo "Checking ${HEALTH_URL}"
ok=0
for _ in $(seq 1 30); do
  if curl -fsS "$HEALTH_URL"; then
    echo
    ok=1
    break
  fi
  sleep 2
done

if [ "$ok" -ne 1 ]; then
  echo "Health check failed: ${HEALTH_URL}" >&2
  exit 1
fi
