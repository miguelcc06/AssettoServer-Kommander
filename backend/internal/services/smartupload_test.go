package services

import (
	"archive/zip"
	"encoding/json"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallExtractedPacksCarsAndTracks(t *testing.T) {
	root := t.TempDir()
	content := filepath.Join(root, "content")
	downloads := filepath.Join(root, "downloads")
	data := filepath.Join(root, "data")
	cfg := filepath.Join(data, "cfg")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	ini := filepath.Join(cfg, "server_cfg.ini")
	if err := os.WriteFile(ini, []byte("[SERVER]\nTRACK=imola\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AC_CONTENT_PATH", content)
	t.Setenv("AC_DOWNLOADS_PATH", downloads)
	t.Setenv("PUBLIC_DOWNLOAD_BASE_URL", "http://example.test/downloads")
	t.Setenv("AC_SERVER_INI_PATH", ini)
	t.Setenv("AC_CM_CONTENT_JSON", "")

	extract := t.TempDir()
	carUI := filepath.Join(extract, "content", "cars", "ks_test", "ui")
	if err := os.MkdirAll(carUI, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(carUI, "ui_car.json"), []byte(`{"name":"Test","version":"1.2.3"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extract, "content", "cars", "ks_test", "data.acd"), []byte("acd"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(extract, "loose_car"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extract, "loose_car", "data.acd"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	layout := filepath.Join(extract, "tracks", "imola", "ui", "gp")
	if err := os.MkdirAll(layout, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(layout, "ui_track.json"), []byte(`{"name":"Imola GP","version":"9"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(extract, "tracks", "imola", "models.ini"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	items, err := installExtracted(extract, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("installed %#v", items)
	}

	if _, err := os.Stat(filepath.Join(content, "cars", "ks_test", "ui", "ui_car.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(content, "cars", "loose_car", "data.acd")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(content, "tracks", "imola", "ui", "gp", "ui_track.json")); err != nil {
		t.Fatal(err)
	}

	assertZipEntry(t, filepath.Join(downloads, "ks_test.zip"), "content/cars/ks_test/ui/ui_car.json")
	assertZipEntry(t, filepath.Join(downloads, "ks_test.zip"), "content/cars/ks_test/data.acd")
	assertZipEntry(t, filepath.Join(downloads, "imola.zip"), "content/tracks/imola/ui/gp/ui_track.json")
	assertZipEntry(t, filepath.Join(downloads, "loose_car.zip"), "content/cars/loose_car/data.acd")

	for _, path := range []string{
		filepath.Join(data, "cm_content", "content.json"),
		filepath.Join(cfg, "cm_content", "content.json"),
	} {
		doc := readJSON(t, path)
		cars := doc["cars"].(map[string]any)
		ks := cars["ks_test"].(map[string]any)
		if ks["url"] != "http://example.test/downloads/ks_test.zip" || ks["version"] != "1.2.3" {
			t.Fatalf("%s cars.ks_test = %#v", path, ks)
		}
		if cars["loose_car"].(map[string]any)["url"] != "http://example.test/downloads/loose_car.zip" {
			t.Fatalf("%s missing loose_car", path)
		}
		tracks := doc["tracks"].(map[string]any)
		imola := tracks["imola"].(map[string]any)
		if imola["url"] != "http://example.test/downloads/imola.zip" || imola["version"] != "9" {
			t.Fatalf("%s tracks.imola = %#v", path, imola)
		}
		track := doc["track"].(map[string]any)
		if track["url"] != "http://example.test/downloads/imola.zip" {
			t.Fatalf("%s track = %#v", path, track)
		}
	}
}

func TestInstallExtractedMergesExistingContentJSON(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AC_CONTENT_PATH", filepath.Join(root, "content"))
	t.Setenv("AC_DOWNLOADS_PATH", filepath.Join(root, "downloads"))
	t.Setenv("PUBLIC_DOWNLOAD_BASE_URL", "https://cdn.example/packs")
	cmPath := filepath.Join(root, "cm_content", "content.json")
	t.Setenv("AC_CM_CONTENT_JSON", cmPath)
	t.Setenv("AC_SERVER_INI_PATH", filepath.Join(root, "missing.ini"))

	if err := os.MkdirAll(filepath.Dir(cmPath), 0o755); err != nil {
		t.Fatal(err)
	}
	original := []byte(`{"cars":{"old_car":{"url":"https://cdn.example/packs/old_car.zip","version":"1"}},"track":{"url":"https://cdn.example/packs/old_track.zip"}}`)
	if err := os.WriteFile(cmPath, original, 0o644); err != nil {
		t.Fatal(err)
	}

	extract := t.TempDir()
	car := filepath.Join(extract, "new_car", "data")
	if err := os.MkdirAll(car, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(car, "car.ini"), []byte("[HEADER]\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	trackUI := filepath.Join(extract, "spa", "ui")
	if err := os.MkdirAll(trackUI, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trackUI, "ui_track.json"), []byte(`{"name":"Spa"}`), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := installExtracted(extract, ""); err != nil {
		t.Fatal(err)
	}

	doc := readJSON(t, cmPath)
	cars := doc["cars"].(map[string]any)
	if cars["old_car"].(map[string]any)["url"] != "https://cdn.example/packs/old_car.zip" {
		t.Fatalf("old car was dropped: %#v", cars)
	}
	if cars["new_car"].(map[string]any)["url"] != "https://cdn.example/packs/new_car.zip" {
		t.Fatalf("new car missing: %#v", cars)
	}
	if _, ok := cars["new_car"].(map[string]any)["version"]; ok {
		t.Fatal("empty version should be omitted")
	}
	track := doc["track"].(map[string]any)
	if track["url"] != "https://cdn.example/packs/spa.zip" {
		t.Fatalf("track = %#v", track)
	}
}

func TestProcessUploadZipAtArchiveRoot(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AC_CONTENT_PATH", filepath.Join(root, "content"))
	t.Setenv("AC_DOWNLOADS_PATH", filepath.Join(root, "downloads"))
	t.Setenv("PUBLIC_DOWNLOAD_BASE_URL", "http://127.0.0.1:8086/downloads")
	t.Setenv("AC_CM_CONTENT_JSON", filepath.Join(root, "cm_content", "content.json"))

	zipPath := filepath.Join(root, "ks_porsche.zip")
	writeZip(t, zipPath, map[string]string{
		"ui/ui_car.json": `{"name":"Porsche","version":"2"}`,
		"data.acd":       "secret",
	})

	f, err := os.Open(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	items, err := ProcessUpload(f, &multipart.FileHeader{Filename: "ks_porsche.zip"})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || items[0].ID != "ks_porsche" || items[0].Type != "car" {
		t.Fatalf("items = %#v", items)
	}
	if items[0].URL != "http://127.0.0.1:8086/downloads/ks_porsche.zip" {
		t.Fatalf("url = %s", items[0].URL)
	}
	assertZipEntry(t, filepath.Join(root, "downloads", "ks_porsche.zip"), "content/cars/ks_porsche/ui/ui_car.json")
	if _, err := os.Stat(filepath.Join(root, "content", "cars", "ks_porsche", "data.acd")); err != nil {
		t.Fatal(err)
	}
}

func TestExtractZipRejectsZipSlip(t *testing.T) {
	root := t.TempDir()
	zipPath := filepath.Join(root, "bad.zip")
	writeZip(t, zipPath, map[string]string{
		"../evil.txt": "nope",
	})
	dest := filepath.Join(root, "out")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	err := extractZip(zipPath, dest)
	if err == nil || !strings.Contains(err.Error(), "zip slip") {
		t.Fatalf("expected zip slip error, got %v", err)
	}
	if _, statErr := os.Stat(filepath.Join(root, "evil.txt")); !os.IsNotExist(statErr) {
		t.Fatal("zip slip file was written")
	}
}

func TestSaveIntoRejectsZipSlip(t *testing.T) {
	base := t.TempDir()
	err := saveInto(base, "../escape.txt", strings.NewReader("x"))
	if err == nil || !strings.Contains(err.Error(), "zip slip") {
		t.Fatalf("expected zip slip, got %v", err)
	}
}

func TestPublicDownloadURLRelative(t *testing.T) {
	t.Setenv("PUBLIC_DOWNLOAD_BASE_URL", "")
	if got := publicDownloadURL("ks_test"); got != "/ks_test.zip" {
		t.Fatalf("got %s", got)
	}
	t.Setenv("PUBLIC_DOWNLOAD_BASE_URL", "http://example.test/downloads/")
	if got := publicDownloadURL("my car"); got != "http://example.test/downloads/my%20car.zip" {
		t.Fatalf("got %s", got)
	}
}

func assertZipEntry(t *testing.T, zipPath, name string) {
	t.Helper()
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	for _, file := range reader.File {
		if file.Name == name {
			return
		}
	}
	names := make([]string, 0, len(reader.File))
	for _, file := range reader.File {
		names = append(names, file.Name)
	}
	t.Fatalf("%s missing %s, has %v", zipPath, name, names)
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	return doc
}

func writeZip(t *testing.T, path string, files map[string]string) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	for name, body := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(body)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
}
