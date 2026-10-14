package services

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// ProcessUpload handles an archive file (.zip, .rar, .7z)
func ProcessUpload(file multipart.File, header *multipart.FileHeader) error {
	ext := strings.ToLower(filepath.Ext(header.Filename))

	if ext == ".zip" || ext == ".rar" || ext == ".7z" {
		return processArchiveUniversal(file, ext)
	}
	return fmt.Errorf("unsupported format: %s", ext)
}

// ProcessRawFiles handles dropped folder structures
func ProcessRawFiles(headers []*multipart.FileHeader) error {
	for _, header := range headers {
		if shouldIgnore(header.Filename) {
			continue
		}

		file, err := header.Open()
		if err != nil {
			return err
		}

		err = saveFile(header.Filename, file)
		file.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func processArchiveUniversal(file multipart.File, ext string) error {
	// 1. Write the uploaded stream to a temporary archive file
	tmpFile, err := os.CreateTemp("", "upload-*"+ext)
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := io.Copy(tmpFile, file); err != nil {
		tmpFile.Close()
		return err
	}
	tmpFile.Close()

	// 2. Create a temporary extraction directory
	extractPath, err := os.MkdirTemp("", "extract-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(extractPath)

	// 3. Extract the archive using 7z (which supports zip, rar, 7z)
	cmd := exec.Command("7z", "x", "-y", "-o"+extractPath, tmpFile.Name())
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("7z extract failed: %v, output: %s", err, string(output))
	}

	// 4. Run a robust bash script to automatically find cars, tracks, and weather
	// We search for known signature files (ui_track.json, ui_car.json) to locate the root of the mod.
	basePath := os.Getenv("AC_CONTENT_PATH")
	if basePath == "" {
		basePath = "/data/content"
	}

	script := fmt.Sprintf(`
		export EXTRACT_PATH="%[1]s"
		export BASE_PATH="%[2]s"

		# 1. Discover Tracks (look for ui_track.json)
		# The structure is usually tracks/<track_name>/ui/ui_track.json or tracks/<track_name>/ui/<layout>/ui_track.json
		find "$EXTRACT_PATH" -type f -iname "ui_track.json" | while read -r uifile; do
			# Get the directory containing ui_track.json
			ui_dir=$(dirname "$uifile")
			
			# If the directory is named 'ui', the track root is one level up.
			# If the directory is inside 'ui', the track root is two levels up.
			ui_dirname=$(basename "$ui_dir")
			
			if [ "$ui_dirname" = "ui" ]; then
				track_dir=$(dirname "$ui_dir")
			else
				parent_dir=$(dirname "$ui_dir")
				if [ "$(basename "$parent_dir")" = "ui" ]; then
					track_dir=$(dirname "$parent_dir")
				else
					continue
				fi
			fi
			
			track_name=$(basename "$track_dir")
			echo "Found track: $track_name at $track_dir"
			
			mkdir -p "$BASE_PATH/tracks/$track_name"
			rsync -a "$track_dir/" "$BASE_PATH/tracks/$track_name/"
		done

		# 2. Discover Cars (look for ui_car.json)
		# The structure is cars/<car_name>/ui/ui_car.json
		find "$EXTRACT_PATH" -type f -iname "ui_car.json" | while read -r uifile; do
			ui_dir=$(dirname "$uifile")
			if [ "$(basename "$ui_dir")" = "ui" ]; then
				car_dir=$(dirname "$ui_dir")
				car_name=$(basename "$car_dir")
				echo "Found car: $car_name at $car_dir"
				
				mkdir -p "$BASE_PATH/cars/$car_name"
				rsync -a "$car_dir/" "$BASE_PATH/cars/$car_name/"
			fi
		done

		# 3. Discover Weather (fallback for typical weather mod structure)
		if [ -d "$EXTRACT_PATH/content/weather" ]; then
			rsync -a "$EXTRACT_PATH/content/weather/" "$BASE_PATH/weather/"
		fi
		if [ -d "$EXTRACT_PATH/weather" ]; then
			rsync -a "$EXTRACT_PATH/weather/" "$BASE_PATH/weather/"
		fi

		# Fix permissions for ACSM
		chown -R 1000:1000 "$BASE_PATH"
	`, extractPath, basePath)

	cmd = exec.Command("sh", "-c", script)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("mod discovery and move failed: %v, output: %s", err, string(output))
	}

	return nil
}

func shouldIgnore(filename string) bool {
	return false
}

func saveFile(relPath string, r io.Reader) error {
	relPath = strings.TrimPrefix(relPath, "content/")
	relPath = strings.TrimPrefix(relPath, "content\\")

	basePath := os.Getenv("AC_CONTENT_PATH")
	if basePath == "" {
		basePath = "/data/content"
	}

	fullPath := filepath.Join(basePath, relPath)

	cleanBasePath := filepath.Clean(basePath)
	cleanFullPath := filepath.Clean(fullPath)
	if !strings.HasPrefix(cleanFullPath, cleanBasePath+string(os.PathSeparator)) {
		return fmt.Errorf("invalid path (zip slip): %s", relPath)
	}

	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return err
	}
	os.Chown(filepath.Dir(fullPath), 1000, 1000)

	out, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, r)
	if err == nil {
		os.Chown(fullPath, 1000, 1000)
	}
	return err
}
