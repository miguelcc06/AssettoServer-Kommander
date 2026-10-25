package services

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSyncActiveTrackSwitchesSelectedEvent(t *testing.T) {
	root := t.TempDir()
	data := filepath.Join(root, "data")
	cfg := filepath.Join(data, "cfg")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	ini := filepath.Join(cfg, "server_cfg.ini")
	if err := os.WriteFile(ini, []byte("[SERVER]\nTRACK=imola\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AC_SERVER_INI_PATH", ini)
	t.Setenv("AC_CM_CONTENT_JSON", "")

	catalog := `{
  "cars": {"ks_alfa": {"url": "http://example.test/downloads/ks_alfa.zip", "version": "1"}},
  "tracks": {
    "imola": {"url": "http://example.test/downloads/imola.zip", "version": "9"},
    "spa": "http://example.test/downloads/spa.zip"
  },
  "track": {"url": "http://example.test/downloads/imola.zip", "version": "9"},
  "extra": "keep"
}`
	primary := filepath.Join(data, "cm_content", "content.json")
	secondary := filepath.Join(cfg, "cm_content", "content.json")
	for _, path := range []string{primary, secondary} {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		body := catalog
		if path == secondary {
			body = `{"cars":{"only_here":{"url":"http://example.test/downloads/only_here.zip"}},"tracks":{}}`
		}
		if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := SyncActiveTrack("spa"); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{primary, secondary} {
		doc := readJSON(t, path)
		track := doc["track"].(map[string]any)
		if track["url"] != "http://example.test/downloads/spa.zip" {
			t.Fatalf("%s track = %#v", path, track)
		}
		tracks := doc["tracks"].(map[string]any)
		if tracks["imola"] == nil && path == primary {
			t.Fatalf("%s dropped imola: %#v", path, tracks)
		}
		spa := tracks["spa"].(map[string]any)
		if spa["url"] != "http://example.test/downloads/spa.zip" {
			t.Fatalf("%s tracks.spa = %#v", path, spa)
		}
	}

	primaryDoc := readJSON(t, primary)
	if primaryDoc["extra"] != "keep" {
		t.Fatalf("extra field = %#v", primaryDoc["extra"])
	}
	cars := primaryDoc["cars"].(map[string]any)
	if cars["ks_alfa"].(map[string]any)["url"] != "http://example.test/downloads/ks_alfa.zip" {
		t.Fatalf("cars = %#v", cars)
	}
	secondaryDoc := readJSON(t, secondary)
	if secondaryDoc["cars"].(map[string]any)["only_here"] == nil {
		t.Fatalf("secondary cars = %#v", secondaryDoc["cars"])
	}
}

func TestSyncActiveTrackUsesPackagedZip(t *testing.T) {
	root := t.TempDir()
	ini := filepath.Join(root, "cfg", "server_cfg.ini")
	if err := os.MkdirAll(filepath.Dir(ini), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ini, []byte("[SERVER]\nTRACK=old\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("AC_SERVER_INI_PATH", ini)
	t.Setenv("AC_CM_CONTENT_JSON", "")
	t.Setenv("AC_DOWNLOADS_PATH", filepath.Join(root, "downloads"))
	t.Setenv("AC_CONTENT_PATH", filepath.Join(root, "content"))
	t.Setenv("PUBLIC_DOWNLOAD_BASE_URL", "http://example.test/downloads")

	if err := os.MkdirAll(filepath.Join(root, "downloads"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "downloads", "spa.zip"), []byte("zip"), 0o644); err != nil {
		t.Fatal(err)
	}
	ui := filepath.Join(root, "content", "tracks", "spa", "ui", "gp")
	if err := os.MkdirAll(ui, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(ui, "ui_track.json"), []byte(`{"version":"4"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SyncActiveTrack("spa"); err != nil {
		t.Fatal(err)
	}

	for _, path := range []string{
		filepath.Join(root, "cm_content", "content.json"),
		filepath.Join(root, "cfg", "cm_content", "content.json"),
	} {
		doc := readJSON(t, path)
		track := doc["track"].(map[string]any)
		if track["url"] != "http://example.test/downloads/spa.zip" || track["version"] != "4" {
			t.Fatalf("%s track = %#v", path, track)
		}
	}
}

func TestSyncActiveTrackLeavesUnknownTrackUntouched(t *testing.T) {
	root := t.TempDir()
	cmPath := filepath.Join(root, "cm_content", "content.json")
	t.Setenv("AC_CM_CONTENT_JSON", cmPath)
	t.Setenv("AC_DOWNLOADS_PATH", filepath.Join(root, "downloads"))
	t.Setenv("AC_CONTENT_PATH", filepath.Join(root, "content"))
	original := []byte(`{"track":{"url":"http://example.test/downloads/imola.zip"},"tracks":{"imola":{"url":"http://example.test/downloads/imola.zip"}}}`)
	if err := os.MkdirAll(filepath.Dir(cmPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cmPath, original, 0o644); err != nil {
		t.Fatal(err)
	}

	if err := SyncActiveTrack("missing_track"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(cmPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(original) {
		t.Fatalf("content.json changed: %s", got)
	}
	if err := SyncActiveTrack(""); err != nil {
		t.Fatal(err)
	}
}

func TestSyncActiveTrackRejectsInvalidJSON(t *testing.T) {
	root := t.TempDir()
	cmPath := filepath.Join(root, "cm_content", "content.json")
	t.Setenv("AC_CM_CONTENT_JSON", cmPath)
	if err := os.MkdirAll(filepath.Dir(cmPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cmPath, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := SyncActiveTrack("spa")
	if err == nil || !strings.Contains(err.Error(), cmPath) {
		t.Fatalf("expected read error, got %v", err)
	}
	got, readErr := os.ReadFile(cmPath)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if string(got) != "{" {
		t.Fatalf("corrupt file was rewritten: %s", got)
	}
}

func TestSyncActiveTrackMatchesTrackIDCase(t *testing.T) {
	root := t.TempDir()
	cmPath := filepath.Join(root, "cm_content", "content.json")
	t.Setenv("AC_CM_CONTENT_JSON", cmPath)
	if err := os.MkdirAll(filepath.Dir(cmPath), 0o755); err != nil {
		t.Fatal(err)
	}
	body := `{"tracks":{"SPA":{"url":"http://example.test/downloads/spa.zip","version":"2"}},"track":{"url":"http://example.test/downloads/old.zip"}}`
	if err := os.WriteFile(cmPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := SyncActiveTrack("spa"); err != nil {
		t.Fatal(err)
	}
	doc := readJSON(t, cmPath)
	track := doc["track"].(map[string]any)
	if track["url"] != "http://example.test/downloads/spa.zip" || track["version"] != "2" {
		t.Fatalf("track = %#v", track)
	}
	tracks := doc["tracks"].(map[string]any)
	if _, ok := tracks["SPA"]; !ok {
		t.Fatalf("tracks = %#v", tracks)
	}
	if _, ok := tracks["spa"]; ok {
		t.Fatalf("duplicated track key: %#v", tracks)
	}
}
