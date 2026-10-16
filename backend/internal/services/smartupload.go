package services

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

// InstalledItem is one car or track that was installed and packaged.
type InstalledItem struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	URL     string `json:"url"`
	Version string `json:"version,omitempty"`
}

type discovered struct {
	Kind    string
	ID      string
	Path    string
	Version string
}

var cmMu sync.Mutex

// ProcessUpload extracts a .zip, .rar or .7z archive, installs any cars and
// tracks it contains, and packages each one for Content Manager.
func ProcessUpload(file multipart.File, header *multipart.FileHeader) ([]InstalledItem, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext != ".zip" && ext != ".rar" && ext != ".7z" {
		return nil, fmt.Errorf("unsupported format: %s", ext)
	}

	tmpFile, err := os.CreateTemp("", "upload-*"+ext)
	if err != nil {
		return nil, err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		return nil, err
	}
	if err := tmpFile.Close(); err != nil {
		return nil, err
	}

	extractPath, err := os.MkdirTemp("", "extract-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(extractPath)

	if err := extractArchive(tmpFile.Name(), extractPath, ext); err != nil {
		return nil, err
	}

	name := filepath.Base(strings.ReplaceAll(header.Filename, "\\", "/"))
	fallback := strings.TrimSuffix(name, filepath.Ext(name))
	return installExtracted(extractPath, fallback)
}

// ProcessRawFiles installs a dropped folder (or loose files) using the same
// car/track detection as archive uploads.
func ProcessRawFiles(headers []*multipart.FileHeader) ([]InstalledItem, error) {
	extractPath, err := os.MkdirTemp("", "raw-*")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(extractPath)

	var installed []InstalledItem
	saved := 0
	for _, header := range headers {
		if header == nil || shouldIgnore(header.Filename) {
			continue
		}
		ext := strings.ToLower(filepath.Ext(header.Filename))
		if ext == ".zip" || ext == ".rar" || ext == ".7z" {
			file, err := header.Open()
			if err != nil {
				return nil, err
			}
			items, err := ProcessUpload(file, header)
			file.Close()
			if err != nil {
				return nil, err
			}
			installed = append(installed, items...)
			continue
		}

		file, err := header.Open()
		if err != nil {
			return nil, err
		}
		err = saveInto(extractPath, header.Filename, file)
		file.Close()
		if err != nil {
			return nil, err
		}
		saved++
	}
	if saved == 0 && len(installed) == 0 {
		return nil, fmt.Errorf("no file uploaded")
	}
	if saved > 0 {
		items, err := installExtracted(extractPath, "")
		if err != nil {
			return nil, err
		}
		installed = append(installed, items...)
	}
	return installed, nil
}

// DownloadsDir is the folder that stores Content Manager distribution zips.
func DownloadsDir() string {
	if p := os.Getenv("AC_DOWNLOADS_PATH"); p != "" {
		return p
	}
	return "/data/downloads"
}

func contentDir() string {
	if p := os.Getenv("AC_CONTENT_PATH"); p != "" {
		return p
	}
	return "/data/content"
}

func installExtracted(extractPath, fallbackID string) ([]InstalledItem, error) {
	copyWeather(extractPath, contentDir())

	found, err := discoverContent(extractPath, fallbackID)
	if err != nil {
		return nil, err
	}
	if len(found) == 0 {
		return nil, fmt.Errorf("no car or track found; expected ui/ui_car.json, data.acd, data/car.ini or ui/ui_track.json")
	}

	var installed []InstalledItem
	var uploadedTracks []string
	for _, item := range found {
		destRoot := contentDir()
		var dest string
		if item.Kind == "car" {
			dest = filepath.Join(destRoot, "cars", item.ID)
		} else {
			dest = filepath.Join(destRoot, "tracks", item.ID)
		}
		if err := copyTree(item.Path, dest); err != nil {
			return nil, fmt.Errorf("install %s %s: %w", item.Kind, item.ID, err)
		}
		chownTree(dest)

		downloadURL, err := packageDistribution(item.Kind, item.ID, item.Path)
		if err != nil {
			return nil, fmt.Errorf("package %s %s: %w", item.Kind, item.ID, err)
		}
		installed = append(installed, InstalledItem{
			Type:    item.Kind,
			ID:      item.ID,
			URL:     downloadURL,
			Version: item.Version,
		})
		if item.Kind == "track" {
			uploadedTracks = append(uploadedTracks, item.ID)
		}
	}

	sort.Slice(installed, func(i, j int) bool {
		if installed[i].Type != installed[j].Type {
			return installed[i].Type < installed[j].Type
		}
		return installed[i].ID < installed[j].ID
	})

	if err := updateCMContent(installed, uploadedTracks); err != nil {
		return nil, err
	}
	return installed, nil
}

func discoverContent(extractPath, fallbackID string) ([]discovered, error) {
	byPath := map[string]*discovered{}

	err := filepath.Walk(extractPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if shouldSkipPath(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if info.IsDir() || !info.Mode().IsRegular() {
			return nil
		}

		name := strings.ToLower(info.Name())
		parent := filepath.Dir(path)
		parentName := strings.ToLower(filepath.Base(parent))

		switch name {
		case "ui_car.json":
			if parentName != "ui" {
				return nil
			}
			version := readContentVersion(path)
			addDiscovered(byPath, filepath.Dir(parent), "car", version)
		case "ui_track.json":
			var root string
			direct := parentName == "ui"
			if direct {
				root = filepath.Dir(parent)
			} else if strings.ToLower(filepath.Base(filepath.Dir(parent))) == "ui" {
				root = filepath.Dir(filepath.Dir(parent))
			} else {
				return nil
			}
			version := readContentVersion(path)
			addDiscovered(byPath, root, "track", version)
			// The base ui/ui_track.json version wins over a layout variant.
			if direct && version != "" {
				if item := byPath[filepath.Clean(root)]; item != nil {
					item.Version = version
				}
			}
		case "data.acd":
			addDiscovered(byPath, parent, "car", "")
		case "car.ini":
			if parentName == "data" {
				addDiscovered(byPath, filepath.Dir(parent), "car", "")
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	extractClean := filepath.Clean(extractPath)
	byKey := map[string]*discovered{}
	for _, item := range byPath {
		if filepath.Clean(item.Path) == extractClean {
			if strings.TrimSpace(fallbackID) == "" {
				return nil, fmt.Errorf("could not determine an id for %s at the upload root; name the archive or folder after the content id", item.Kind)
			}
			id, err := sanitizeID(fallbackID)
			if err != nil {
				return nil, err
			}
			item.ID = id
		} else {
			id, err := sanitizeID(filepath.Base(item.Path))
			if err != nil {
				return nil, err
			}
			item.ID = id
		}
		key := item.Kind + ":" + item.ID
		if prev, ok := byKey[key]; ok {
			if len(item.Path) < len(prev.Path) {
				if item.Version == "" {
					item.Version = prev.Version
				}
				byKey[key] = item
			} else if prev.Version == "" && item.Version != "" {
				prev.Version = item.Version
			}
			continue
		}
		byKey[key] = item
	}

	out := make([]discovered, 0, len(byKey))
	for _, item := range byKey {
		out = append(out, *item)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].ID < out[j].ID
	})
	return out, nil
}

func addDiscovered(byPath map[string]*discovered, root, kind, version string) {
	key := filepath.Clean(root)
	if existing, ok := byPath[key]; ok {
		if existing.Kind == "" {
			existing.Kind = kind
		}
		// A folder with ui_car.json wins over a bare data.acd guess.
		if existing.Kind != kind && kind == "car" && existing.Version == "" {
			existing.Kind = kind
		}
		if existing.Version == "" && version != "" {
			existing.Version = version
		}
		return
	}
	byPath[key] = &discovered{Kind: kind, Path: key, Version: version}
}

func readContentVersion(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	data = trimBOM(data)
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return ""
	}
	switch v := payload["version"].(type) {
	case string:
		return strings.TrimSpace(v)
	default:
		return ""
	}
}

func trimBOM(b []byte) []byte {
	if len(b) >= 3 && b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		return b[3:]
	}
	return b
}

func sanitizeID(id string) (string, error) {
	id = strings.TrimSpace(id)
	if id == "" || id == "." || id == ".." {
		return "", fmt.Errorf("invalid content id %q", id)
	}
	if strings.ContainsAny(id, `/\`) {
		return "", fmt.Errorf("invalid content id %q", id)
	}
	for _, r := range id {
		if r < 32 || r == 127 {
			return "", fmt.Errorf("invalid content id %q", id)
		}
	}
	return id, nil
}

func packageDistribution(kind, id, src string) (string, error) {
	downloads := DownloadsDir()
	if err := os.MkdirAll(downloads, 0o755); err != nil {
		return "", err
	}

	folder := "cars"
	if kind == "track" {
		folder = "tracks"
	}
	prefix := "content/" + folder + "/" + id + "/"

	zipPath := filepath.Join(downloads, id+".zip")
	tmpPath := zipPath + ".tmp"
	out, err := os.Create(tmpPath)
	if err != nil {
		return "", err
	}

	zw := zip.NewWriter(out)
	walkErr := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if shouldSkipPath(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		name := prefix + filepath.ToSlash(rel)
		if info.IsDir() {
			_, err := zw.Create(name + "/")
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil
		}
		hdr, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		hdr.Name = name
		hdr.Method = zip.Deflate
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, in)
		closeErr := in.Close()
		if copyErr != nil {
			return copyErr
		}
		return closeErr
	})

	closeZipErr := zw.Close()
	closeFileErr := out.Close()
	if walkErr != nil {
		os.Remove(tmpPath)
		return "", walkErr
	}
	if closeZipErr != nil {
		os.Remove(tmpPath)
		return "", closeZipErr
	}
	if closeFileErr != nil {
		os.Remove(tmpPath)
		return "", closeFileErr
	}
	if err := os.Rename(tmpPath, zipPath); err != nil {
		os.Remove(tmpPath)
		return "", err
	}
	chownACSM(zipPath)
	return publicDownloadURL(id), nil
}

func publicDownloadURL(id string) string {
	base := strings.TrimRight(strings.TrimSpace(os.Getenv("PUBLIC_DOWNLOAD_BASE_URL")), "/")
	file := url.PathEscape(id) + ".zip"
	if base == "" {
		return "/" + file
	}
	return base + "/" + file
}

func updateCMContent(installed []InstalledItem, uploadedTracks []string) error {
	cmMu.Lock()
	defer cmMu.Unlock()

	var firstErr error
	for _, path := range cmContentPaths() {
		if err := updateOneCMContent(path, installed, uploadedTracks); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func cmContentPaths() []string {
	if p := strings.TrimSpace(os.Getenv("AC_CM_CONTENT_JSON")); p != "" {
		return []string{p}
	}
	ini := os.Getenv("AC_SERVER_INI_PATH")
	if ini == "" {
		ini = "runtime/data/cfg/server_cfg.ini"
	}
	cfgDir := filepath.Dir(ini)
	serverRoot := filepath.Dir(cfgDir)
	paths := []string{
		filepath.Join(serverRoot, "cm_content", "content.json"),
		filepath.Join(cfgDir, "cm_content", "content.json"),
	}
	seen := map[string]bool{}
	var out []string
	for _, p := range paths {
		clean := filepath.Clean(p)
		if seen[clean] {
			continue
		}
		seen[clean] = true
		out = append(out, clean)
	}
	return out
}

func updateOneCMContent(path string, installed []InstalledItem, uploadedTracks []string) error {
	doc := map[string]any{}
	if data, err := os.ReadFile(path); err == nil && len(strings.TrimSpace(string(data))) > 0 {
		if err := json.Unmarshal(data, &doc); err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}

	cars := ensureMap(doc, "cars")
	tracks := mapFrom(doc["tracks"])

	for _, item := range installed {
		target := cars
		if item.Type == "track" {
			target = tracks
		}
		upsertEntry(target, item.ID, item.URL, item.Version)
	}
	if len(tracks) > 0 {
		doc["tracks"] = tracks
	}

	if id := chooseActiveTrack(tracks, uploadedTracks); id != "" {
		if entry, ok := tracks[id].(map[string]any); ok {
			doc["track"] = cloneMap(entry)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}
	chownACSM(path)
	chownACSM(filepath.Dir(path))
	return nil
}

func chooseActiveTrack(tracks map[string]any, uploaded []string) string {
	configured := configuredTrackID()
	if configured != "" {
		if _, ok := tracks[configured]; ok {
			return configured
		}
	}
	if len(uploaded) > 0 {
		return uploaded[len(uploaded)-1]
	}
	return ""
}

func configuredTrackID() string {
	ini := os.Getenv("AC_SERVER_INI_PATH")
	if ini == "" {
		return ""
	}
	data, err := os.ReadFile(ini)
	if err != nil {
		return ""
	}
	inServer := false
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "[") {
			inServer = strings.EqualFold(line, "[SERVER]")
			continue
		}
		if !inServer {
			continue
		}
		if len(line) >= 6 && strings.EqualFold(line[:6], "TRACK=") {
			return strings.TrimSpace(line[6:])
		}
	}
	return ""
}

func ensureMap(doc map[string]any, key string) map[string]any {
	m := mapFrom(doc[key])
	doc[key] = m
	return m
}

func mapFrom(v any) map[string]any {
	if m, ok := v.(map[string]any); ok && m != nil {
		return m
	}
	return map[string]any{}
}

func upsertEntry(dict map[string]any, id, downloadURL, version string) {
	entry, _ := dict[id].(map[string]any)
	if entry == nil {
		if s, ok := dict[id].(string); ok && s != "" {
			entry = map[string]any{"url": s}
		} else {
			entry = map[string]any{}
		}
	}
	entry["url"] = downloadURL
	if version != "" {
		entry["version"] = version
	}
	dict[id] = entry
}

func cloneMap(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func extractArchive(archivePath, dest, ext string) error {
	if ext == ".zip" {
		if err := extractZip(archivePath, dest); err == nil {
			return nil
		} else if _, lookErr := exec.LookPath("7z"); lookErr != nil {
			return err
		}
	}
	if _, err := exec.LookPath("7z"); err != nil {
		return fmt.Errorf("7z is required to extract %s archives", ext)
	}
	cmd := exec.Command("7z", "x", "-y", "-o"+dest, archivePath)
	if output, err := cmd.CombinedOutput(); err != nil {
		msg := strings.TrimSpace(string(output))
		if len(msg) > 500 {
			msg = msg[:500]
		}
		return fmt.Errorf("7z extract failed: %v, output: %s", err, msg)
	}
	return nil
}

func extractZip(archivePath, dest string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	destClean := filepath.Clean(dest)
	for _, file := range reader.File {
		rel := filepath.Clean(file.Name)
		if rel == "." {
			continue
		}
		if strings.HasPrefix(rel, "..") || filepath.IsAbs(file.Name) {
			return fmt.Errorf("invalid path (zip slip): %s", file.Name)
		}
		target := filepath.Join(destClean, rel)
		if target != destClean && !strings.HasPrefix(target, destClean+string(os.PathSeparator)) {
			return fmt.Errorf("invalid path (zip slip): %s", file.Name)
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := file.Open()
		if err != nil {
			return err
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			in.Close()
			return err
		}
		_, copyErr := io.Copy(out, in)
		closeIn := in.Close()
		closeOut := out.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeIn != nil {
			return closeIn
		}
		if closeOut != nil {
			return closeOut
		}
	}
	return nil
}

func copyWeather(extractPath, basePath string) {
	candidates := []string{
		filepath.Join(extractPath, "content", "weather"),
		filepath.Join(extractPath, "weather"),
	}
	for _, src := range candidates {
		info, err := os.Stat(src)
		if err != nil || !info.IsDir() {
			continue
		}
		_ = copyTree(src, filepath.Join(basePath, "weather"))
	}
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if shouldSkipPath(path) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := dst
		if rel != "." {
			target = filepath.Join(dst, rel)
		}
		if info.IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return err
			}
			chownACSM(target)
			return nil
		}
		if info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
			return nil
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	chownACSM(dst)
	return nil
}

func chownTree(root string) {
	_ = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err == nil {
			chownACSM(path)
		}
		return nil
	})
}

func chownACSM(path string) {
	_ = os.Chown(path, 1000, 1000)
}

func shouldIgnore(filename string) bool {
	return shouldSkipPath(filename)
}

func shouldSkipPath(path string) bool {
	for _, part := range strings.Split(filepath.ToSlash(path), "/") {
		switch strings.ToLower(part) {
		case "__macosx", ".ds_store", "thumbs.db", "desktop.ini":
			return true
		}
	}
	return false
}

func saveInto(base, relPath string, r io.Reader) error {
	relPath = strings.ReplaceAll(relPath, "\\", "/")
	relPath = strings.TrimPrefix(relPath, "content/")
	relPath = filepath.Clean(filepath.FromSlash(relPath))
	if relPath == "." || strings.HasPrefix(relPath, "..") {
		return fmt.Errorf("invalid path (zip slip): %s", relPath)
	}

	fullPath := filepath.Join(base, relPath)
	cleanBase := filepath.Clean(base)
	cleanFull := filepath.Clean(fullPath)
	if cleanFull != cleanBase && !strings.HasPrefix(cleanFull, cleanBase+string(os.PathSeparator)) {
		return fmt.Errorf("invalid path (zip slip): %s", relPath)
	}
	return copyReader(fullPath, r)
}

func saveFile(relPath string, r io.Reader) error {
	return saveInto(contentDir(), relPath, r)
}

func copyReader(fullPath string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}
	chownACSM(filepath.Dir(fullPath))
	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, r); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	chownACSM(fullPath)
	return nil
}
