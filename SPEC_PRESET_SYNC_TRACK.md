# SPEC: Synchronize Active Track in cm_content/content.json on Preset Launch

## Context
In AssettoServer, Content Manager queries the server endpoint `/api/details` (handled by AssettoServer) to check available downloads for missing cars and tracks.
AssettoServer reads `cm_content/content.json` (at `cfg/cm_content/content.json` and `<server_root>/cm_content/content.json`).
AssettoServer expects a top-level `"track"` object containing the download info for the currently loaded track:
```json
"track": {
  "url": "http://mc.kelpiecas.xyz/assetto/sc_arteixo.zip",
  "version": "BlckBox 0.01"
}
```
Currently, `smartupload.go` populates `content.json` with a catalog of `"tracks": { "<track_id>": { "url": "...", "version": "..." } }`.
However, when a preset is launched via `POST /api/presets/:id/launch` in `backend/internal/api/presets.go`, the backend updates `server_cfg.ini`, `entry_list.ini`, and `extra_cfg.yml`, but does NOT update the top-level `"track"` object in `content.json`.
As a result, if the server changes tracks (e.g. from `la_canyons` to `sc_arteixo`), `content.json` still points to the old track in `"track"`, causing Content Manager to download the wrong track or fail to download automatically.

## Requirements

### 1. Synchronize `track` in `cm_content/content.json` on Preset Launch
In `backend/internal/api/presets.go` (inside `launchPresetHandler`):
- Before restarting the server via `process.Manager.Restart()`, update the active track in `cm_content/content.json`.
- The update must be applied to all configured `cm_content` paths (reusing `services.SyncActiveTrack` or a helper that covers `cmContentPaths()`).
- Look up `targetPreset.Track` in the existing `tracks` dictionary in `content.json`.
  - If found in `tracks[targetPreset.Track]`, copy that entry to `doc["track"]`.
  - If not found in `tracks`, but `PUBLIC_DOWNLOAD_BASE_URL` is set, construct the URL `${PUBLIC_DOWNLOAD_BASE_URL}/${targetPreset.Track}.zip` and set `doc["track"] = map[string]any{"url": url}`.
- Write the updated JSON with indentation (preserving permissions/formatting).

### 2. Preserve Server Assists Flags
In `backend/internal/api/presets.go`, ensure `server_cfg.ini` template preserves the driving assists:
```ini
ABS_ALLOWED=1
TC_ALLOWED=1
STABILITY_ALLOWED=1
AUTOCLUTCH_ALLOWED=1
TYRE_BLANKETS_ALLOWED=1
```

### 3. Testing & Verification
- Add or update unit tests in `backend/internal/api/` or `backend/internal/services/` ensuring that launching a preset with track `sc_arteixo` updates `"track"` in `content.json` to match `sc_arteixo`.
- Ensure `go test ./...` passes in `backend/`.

### 4. Deployment
- Target repository: `AssettoServer-Kommander`
- Once implemented and tests pass, trigger deployment webhook:
  `curl -X POST https://deploy.kelpiecas.xyz/deploy/AssettoServer-Kommander`
