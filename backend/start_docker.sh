#!/bin/bash
# Local helper. The supported entrypoint is the repository-root deploy.sh / docker compose.
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "${HOST_PROJECT_DIR:-$ROOT}"
exec docker compose up
