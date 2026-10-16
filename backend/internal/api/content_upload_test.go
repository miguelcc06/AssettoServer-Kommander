package api

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestUploadContentHandlerFolderUsesPaths(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AC_CONTENT_PATH", filepath.Join(root, "content"))
	t.Setenv("AC_DOWNLOADS_PATH", filepath.Join(root, "downloads"))
	t.Setenv("PUBLIC_DOWNLOAD_BASE_URL", "http://example.test/downloads")
	t.Setenv("AC_CM_CONTENT_JSON", filepath.Join(root, "cm_content", "content.json"))
	t.Setenv("AC_SERVER_INI_PATH", filepath.Join(root, "cfg", "server_cfg.ini"))
	if err := os.MkdirAll(filepath.Dir(os.Getenv("AC_SERVER_INI_PATH")), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(os.Getenv("AC_SERVER_INI_PATH"), []byte("[SERVER]\nTRACK=imola\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	// Filenames are base names, matching Go's multipart parser. The relative
	// paths travel in paths[].
	if err := addFormFile(writer, "files[]", "ui_track.json", `{"name":"Imola","version":"9"}`); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField("paths[]", "imola/ui/gp/ui_track.json"); err != nil {
		t.Fatal(err)
	}
	if err := addFormFile(writer, "files[]", "models.ini", "MODEL=x"); err != nil {
		t.Fatal(err)
	}
	if err := writer.WriteField("paths[]", "imola/models.ini"); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/content/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "test"})
	rr := httptest.NewRecorder()
	UploadContentHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}

	var resp UploadResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Installed) != 1 || resp.Installed[0].ID != "imola" || resp.Installed[0].Type != "track" {
		t.Fatalf("installed = %#v", resp.Installed)
	}
	if _, err := os.Stat(filepath.Join(root, "content", "tracks", "imola", "ui", "gp", "ui_track.json")); err != nil {
		t.Fatal(err)
	}
}

func addFormFile(writer *multipart.Writer, field, name, contents string) error {
	part, err := writer.CreateFormFile(field, name)
	if err != nil {
		return err
	}
	_, err = part.Write([]byte(contents))
	return err
}

func TestUploadContentHandlerPacksZip(t *testing.T) {
	root := t.TempDir()
	t.Setenv("AC_CONTENT_PATH", filepath.Join(root, "content"))
	t.Setenv("AC_DOWNLOADS_PATH", filepath.Join(root, "downloads"))
	t.Setenv("PUBLIC_DOWNLOAD_BASE_URL", "http://example.test/downloads")
	t.Setenv("AC_CM_CONTENT_JSON", filepath.Join(root, "cm_content", "content.json"))

	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	ui, err := zw.Create("my_track/ui/ui_track.json")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ui.Write([]byte(`{"name":"My Track","version":"3"}`)); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "ignored.zip")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(zipBuf.Bytes()); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/content/upload", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session_token", Value: "test"})
	rr := httptest.NewRecorder()

	UploadContentHandler(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}

	var resp UploadResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Installed) != 1 || resp.Installed[0].Type != "track" || resp.Installed[0].ID != "my_track" {
		t.Fatalf("installed = %#v", resp.Installed)
	}
	if resp.Installed[0].URL != "http://example.test/downloads/my_track.zip" {
		t.Fatalf("url = %s", resp.Installed[0].URL)
	}
	if _, err := os.Stat(filepath.Join(root, "content", "tracks", "my_track", "ui", "ui_track.json")); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "downloads", "my_track.zip")); err != nil {
		t.Fatal(err)
	}
}
