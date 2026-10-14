# SPEC.md — AssettoServer-Kommander Containerization & Hardcoded Paths Refactor

## Objective
Remove all hardcoded developer paths (`/home/rs/...`) across the backend and docker configurations, make paths fully configurable via environment variables, ensure seamless deployment via Docker Compose, and provide clean defaults so `compujuckel/assettoserver:latest` starts without plugin missing errors.

## Context & Tech Stack
- **Backend:** Go 1.22+ (net/http, sqlite, docker CLI)
- **Frontend:** React + Vite + TailwindCSS (served via Nginx in multi-stage Docker build)
- **Engine Container:** `compujuckel/assettoserver:latest`
- **Host Target:** Linux x86_64 server (`/home/miguelcc06/AssettoServer-Kommander`)

## Specific Requirements

### 1. Remove Hardcoded `/home/rs` in Go Backend
- `backend/cmd/server/main.go`:
  - `captureFrontendLogs()`: remove hardcoded `/home/rs/web-gui-acserver/frontend/frontend.log`. If frontend logging tail is required, make the path configurable via `FRONTEND_LOG_PATH` env var, and if the file does not exist, log a gentle debug notice and do not fail or continuously error.
- `backend/internal/services/smartupload.go`:
  - Replace any fallback `/home/rs/docker/acsm/server/assetto/content` with `/data/content` (or value of `AC_CONTENT_PATH`).
- `backend/internal/api/content.go`:
  - Replace any fallback `/home/rs/docker/acsm/server/assetto/content` with `/data/content` (or value of `AC_CONTENT_PATH`).
- `backend/internal/api/upgrade.go`:
  - In `ServerUpgradeHandler`, `PerformAutoUpdate`, and `ServerUpgradeZipHandler`:
  - Replace hardcoded `--project-directory /home/rs/web-gui-acserver` with an environment variable `HOST_PROJECT_DIR` (defaulting to `/app/host` or configurable via env `HOST_PROJECT_DIR`).
  - Use modern `docker compose` (with space) with fallback to `docker-compose`.

### 2. Docker Compose & Environment Configuration
- In `docker-compose.yaml`:
  - Update services volume mounts:
    - In `assettoserver`:
      - Mount `./backend/runtime/data:/data`
      - Mount `./backend/runtime/data/content:/data/content` (replace `/home/rs/docker/acsm/server/assetto/content`)
    - In `kommander-backend`:
      - Mount `./backend/runtime/data:/app/data`
      - Mount `./backend/runtime/data/content:/data/content`
      - Mount `/var/run/docker.sock:/var/run/docker.sock`
      - Mount `./:/app/host`
      - Environment:
        - `AC_SERVER_INI_PATH=/app/data/cfg/server_cfg.ini`
        - `AC_SERVER_YAML_PATH=/app/data/cfg/extra_cfg.yml`
        - `AC_CONTENT_PATH=/data/content`
        - `AC_PRESETS_PATH=/app/data/presets.json`
        - `AC_LOGS_PATH=/app/data/logs`
        - `HOST_PROJECT_DIR=${PWD}` or `${HOST_PROJECT_DIR:-.}`
        - `PORT=8080`
    - In `kommander-frontend`:
      - Map port `${KOMMANDER_PORT:-8086}:80` (default port 8086 on host to avoid port 3000 collision).
- Update `.env.example` and `backend/.env.example` to document all clean environment variables (`HOST_PROJECT_DIR`, `KOMMANDER_PORT`, `ASSETTOSERVER_IMAGE`).

### 3. Default Configuration & Preset Fixes
- In `backend/runtime/data/cfg/extra_cfg.yml`:
  - By default, standard `compujuckel/assettoserver` only contains built-in plugins (`RandomWeatherPlugin`, `AutoModerationPlugin`, `DiscordAuditPlugin`, `LiveWeatherPlugin`, etc.). It does NOT include `FastTravelPlugin` out-of-the-box.
  - Set default `EnablePlugins:` to `[RandomWeatherPlugin]` so the container boots up successfully out of the box without throwing `ConfigurationException: No plugin found with name FastTravelPlugin`.

### 4. Deployment Script
- Add an executable `deploy.sh` at the repository root that:
  1. Creates any missing directories (`backend/runtime/data/content/cars`, `backend/runtime/data/content/tracks`, `backend/runtime/data/logs`, `backend/runtime/data/cfg`).
  2. Runs `docker compose up -d --build`.
  3. Checks health at `http://127.0.0.1:${KOMMANDER_PORT:-8086}/api/health`.

### 5. Deployment Webhook Notification
- When done, execute curl POST to `https://deploy.kelpiecas.xyz/deploy/AssettoServer-Kommander`.
