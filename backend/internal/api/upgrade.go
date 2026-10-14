package api

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"web-gui-acserver/internal/configparser"
	"web-gui-acserver/internal/db"
)

type UpgradeRequest struct {
	IsAutoMode     bool   `json:"isAutoMode"`
	GithubUsername string `json:"githubUsername"`
	GithubToken    string `json:"githubToken"`
	Image          string `json:"image"`
	PatreonKey     string `json:"patreonKey"`
}

func hostProjectDir() string {
	if dir := os.Getenv("HOST_PROJECT_DIR"); dir != "" {
		return dir
	}
	return "/app/host"
}

func composeBaseArgs() []string {
	return []string{
		"--project-directory", hostProjectDir(),
		"--env-file", "/app/host/.env",
		"-p", "web-gui-acserver",
		"-f", "/app/host/docker-compose.yaml",
	}
}

// runCompose prefers modern `docker compose` and falls back to `docker-compose`.
func runCompose(args ...string) ([]byte, error) {
	composeArgs := append(composeBaseArgs(), args...)
	out, err := exec.Command("docker", append([]string{"compose"}, composeArgs...)...).CombinedOutput()
	if err == nil {
		return out, nil
	}
	return exec.Command("docker-compose", composeArgs...).CombinedOutput()
}

func ServerUpgradeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method == http.MethodGet {
		username := db.GetSetting("github_username", "")
		token := db.GetSetting("github_token", "")
		yamlPath := os.Getenv("AC_SERVER_YAML_PATH")
		if yamlPath == "" {
			yamlPath = "/app/data/cfg/extra_cfg.yml"
		}
		patreonKey := configparser.ReadYAML(yamlPath, "PatreonHubPlugin.Key")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"githubUsername": username,
			"githubToken":    token,
			"patreonKey":     patreonKey,
		})
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"message": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req UpgradeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request payload"}`, http.StatusBadRequest)
		return
	}

	// Handle Automatic Mode (GitHub Login)
	if req.IsAutoMode {
		if req.GithubUsername == "" || req.GithubToken == "" {
			http.Error(w, `{"message": "GitHub username and token are required for automatic mode"}`, http.StatusBadRequest)
			return
		}

		loginCmd := exec.Command("docker", "login", "ghcr.io", "-u", req.GithubUsername, "--password-stdin")
		loginCmd.Stdin = strings.NewReader(req.GithubToken)
		loginOut, err := loginCmd.CombinedOutput()
		if err != nil {
			http.Error(w, fmt.Sprintf(`{"message": "GitHub authentication failed", "error": %q}`, string(loginOut)), http.StatusUnauthorized)
			return
		}

		// Save credentials for the background task
		db.SetSetting("github_username", req.GithubUsername)
		db.SetSetting("github_token", req.GithubToken)

		// Fetch the latest tag from GitHub API
		latestTag := "latest"
		resp, err := http.Get("https://api.github.com/repos/compujuckel/AssettoServer/releases/latest")
		if err == nil {
			defer resp.Body.Close()
			var release struct {
				TagName string `json:"tag_name"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&release); err == nil && release.TagName != "" {
				latestTag = release.TagName
			}
		}

		// Set the official Patreon registry image for automatic mode
		req.Image = "ghcr.io/compujuckel/assettoserver-patreon:" + latestTag
	}

	// 1. Update the .env file with the new image
	envPath := "/app/host/.env"
	envContent, err := os.ReadFile(envPath)
	if err != nil && !os.IsNotExist(err) {
		http.Error(w, `{"message": "Failed to read .env file"}`, http.StatusInternalServerError)
		return
	}

	envLines := strings.Split(string(envContent), "\n")
	updatedEnv := false
	for i, line := range envLines {
		if strings.HasPrefix(line, "ASSETTOSERVER_IMAGE=") {
			envLines[i] = "ASSETTOSERVER_IMAGE=" + req.Image
			updatedEnv = true
			break
		}
	}
	if !updatedEnv {
		envLines = append(envLines, "ASSETTOSERVER_IMAGE="+req.Image)
	}

	if err := os.WriteFile(envPath, []byte(strings.Join(envLines, "\n")), 0644); err != nil {
		http.Error(w, `{"message": "Failed to write .env file"}`, http.StatusInternalServerError)
		return
	}

	// 2. Inject Patreon Key into extra_cfg.yml if provided
	if req.PatreonKey != "" {
		yamlPath := os.Getenv("AC_SERVER_YAML_PATH")
		if yamlPath == "" {
			yamlPath = "/app/data/cfg/extra_cfg.yml"
		}

		if err := configparser.UpdateYAML(yamlPath, "PatreonHubPlugin.Key", req.PatreonKey); err != nil {
			http.Error(w, `{"message": "Failed to save Patreon Key to extra_cfg.yml"}`, http.StatusInternalServerError)
			return
		}
		if err := configparser.UpdateYAML(yamlPath, "PatreonHubPlugin.Enable", "true"); err != nil {
			http.Error(w, `{"message": "Failed to enable Patreon plugin in extra_cfg.yml"}`, http.StatusInternalServerError)
			return
		}
	}

	// 3. Restart AssettoServer container with docker compose
	out, err := runCompose("up", "-d", "assettoserver")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "Failed to rebuild container", "error": %q}`, string(out)), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Upgrade successful! Server is restarting.",
	})
}

func PerformAutoUpdate() (bool, error) {
	username := db.GetSetting("github_username", "")
	token := db.GetSetting("github_token", "")

	if username != "" && token != "" {
		// 1. Login to GitHub Container Registry
		loginCmd := exec.Command("docker", "login", "ghcr.io", "-u", username, "--password-stdin")
		loginCmd.Stdin = strings.NewReader(token)
		if err := loginCmd.Run(); err != nil {
			return false, fmt.Errorf("GitHub authentication failed")
		}
	}

	// 2. Pull the latest image
	if _, err := runCompose("pull", "assettoserver"); err != nil {
		return false, fmt.Errorf("Failed to pull image: %v", err)
	}

	// Check if the image was actually updated (docker compose pull output often says "Status: Downloaded newer image" or similar, or we just run up -d and it recreates)
	upOut, err := runCompose("up", "-d", "assettoserver")
	if err != nil {
		return false, fmt.Errorf("Failed to restart container: %v, output: %s", err, string(upOut))
	}

	outStr := string(upOut)
	if strings.Contains(outStr, "Recreated") || strings.Contains(outStr, "Started") {
		// It updated!
		return true, nil
	}

	// If it says "Running" and not recreated, it means it was already up to date.
	return false, nil
}

func CheckUpdateHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == "OPTIONS" {
		return
	}

	updated, err := PerformAutoUpdate()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if updated {
		w.Write([]byte(`{"message": "Une nouvelle mise à jour a été trouvée et installée. Le serveur redémarre.", "updated": true}`))
	} else {
		w.Write([]byte(`{"message": "Le serveur est déjà à jour avec la dernière version.", "updated": false}`))
	}
}

func ServerUpgradeZipHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"message": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(100 << 20) // 100 MB max
	if err != nil {
		http.Error(w, `{"message": "Failed to parse form"}`, http.StatusBadRequest)
		return
	}

	patreonKey := r.FormValue("patreonKey")
	if patreonKey != "" {
		yamlPath := os.Getenv("AC_SERVER_YAML_PATH")
		if yamlPath == "" {
			yamlPath = "/app/data/cfg/extra_cfg.yml"
		}
		configparser.UpdateYAML(yamlPath, "PatreonHubPlugin.Key", patreonKey)
		configparser.UpdateYAML(yamlPath, "PatreonHubPlugin.Enable", "true")
	}

	file, _, err := r.FormFile("zipfile")
	if err != nil {
		http.Error(w, `{"message": "Failed to get file from request"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Create temp directory for building the custom docker image
	buildDir := "/app/data/build/patreon"
	os.RemoveAll(buildDir)
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		http.Error(w, `{"message": "Failed to create build directory"}`, http.StatusInternalServerError)
		return
	}

	// Save uploaded zip file
	zipPath := filepath.Join(buildDir, "patreon.zip")
	outZip, err := os.Create(zipPath)
	if err != nil {
		http.Error(w, `{"message": "Failed to create zip file on server"}`, http.StatusInternalServerError)
		return
	}
	if _, err := io.Copy(outZip, file); err != nil {
		outZip.Close()
		http.Error(w, `{"message": "Failed to save zip file"}`, http.StatusInternalServerError)
		return
	}
	outZip.Close()

	// Extract zip file
	archive, err := zip.OpenReader(zipPath)
	if err != nil {
		http.Error(w, `{"message": "Failed to open zip file"}`, http.StatusInternalServerError)
		return
	}
	defer archive.Close()

	for _, f := range archive.File {
		filePath := filepath.Join(buildDir, f.Name)
		if !strings.HasPrefix(filePath, filepath.Clean(buildDir)+string(os.PathSeparator)) {
			continue // Prevent Zip Slip
		}
		if f.FileInfo().IsDir() {
			os.MkdirAll(filePath, os.ModePerm)
			continue
		}
		if err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm); err != nil {
			continue
		}
		outFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			continue
		}
		io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
	}

	// Validate Patreon zip
	if _, err := os.Stat(filepath.Join(buildDir, "AssettoServer")); os.IsNotExist(err) {
		if _, err := os.Stat(filepath.Join(buildDir, "plugins", "PatreonHubPlugin.dll")); os.IsNotExist(err) {
			os.RemoveAll(buildDir)
			http.Error(w, `{"message": "L'archive importée ne semble pas être la version Patreon valide d'AssettoServer (fichier AssettoServer ou PatreonHubPlugin.dll introuvable)."}`, http.StatusBadRequest)
			return
		}
	}

	// Create Dockerfile
	dockerfileContent := `FROM compujuckel/assettoserver:latest
COPY . /app
RUN chmod +x /app/AssettoServer
`
	if err := os.WriteFile(filepath.Join(buildDir, "Dockerfile"), []byte(dockerfileContent), 0644); err != nil {
		http.Error(w, `{"message": "Failed to create Dockerfile"}`, http.StatusInternalServerError)
		return
	}

	// Build custom Docker image
	buildCmd := exec.Command("docker", "build", "-t", "custom-assettoserver-patreon:latest", ".")
	buildCmd.Dir = buildDir
	buildOut, err := buildCmd.CombinedOutput()
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "Failed to build custom image", "error": %q}`, string(buildOut)), http.StatusInternalServerError)
		return
	}

	// Update .env file with new image
	reqImage := "custom-assettoserver-patreon:latest"
	envPath := "/app/host/.env"
	envContent, err := os.ReadFile(envPath)
	if err == nil {
		envLines := strings.Split(string(envContent), "\n")
		updatedEnv := false
		for i, line := range envLines {
			if strings.HasPrefix(line, "ASSETTOSERVER_IMAGE=") {
				envLines[i] = "ASSETTOSERVER_IMAGE=" + reqImage
				updatedEnv = true
				break
			}
		}
		if !updatedEnv {
			envLines = append(envLines, "ASSETTOSERVER_IMAGE="+reqImage)
		}
		os.WriteFile(envPath, []byte(strings.Join(envLines, "\n")), 0644)
	}

	// Restart AssettoServer container
	out, err := runCompose("up", "-d", "assettoserver")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "Failed to rebuild container", "error": %q}`, string(out)), http.StatusInternalServerError)
		return
	}

	// Cleanup
	os.RemoveAll(buildDir)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Archive Patreon installée avec succès ! Le serveur redémarre.",
	})
}
