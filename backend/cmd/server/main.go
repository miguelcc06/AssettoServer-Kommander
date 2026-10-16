package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"

	"web-gui-acserver/internal/api"
	"web-gui-acserver/internal/db"
	"web-gui-acserver/internal/process"
)

type HealthResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
	Role    string `json:"role,omitempty"`
}

func enableCORS(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	(*w).Header().Set("Access-Control-Allow-Credentials", "true")
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := HealthResponse{
		Status:  "ok",
		Message: "Backend is connected!",
	}
	json.NewEncoder(w).Encode(response)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"message": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	user, err := db.CheckUserCredentials(req.Username, req.Password)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LoginResponse{
			Message: "Invalid username or password",
		})
		return
	}

	if user.MustChangePassword {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"require_password_change": true,
			"message":                 "Changement de mot de passe requis",
		})
		return
	}

	token := uuid.New().String()
	expiresAt := time.Now().Add(24 * time.Hour)

	if err := db.CreateSession(token, user.ID, expiresAt); err != nil {
		log.Println("Error creating session:", err)
		http.Error(w, `{"message": "Internal error"}`, http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	json.NewEncoder(w).Encode(LoginResponse{
		Message: "Login successful",
		Role:    user.Role,
	})
	db.LogAction("Logged in successfully", user.Username)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}
	// Expire the cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Logged out"})
}

func verifyAuthHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")

	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Not authenticated"}`))
		return
	}

	user, err := db.GetUserByToken(cookie.Value)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid or expired session"}`))
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":       "Authenticated",
		"authenticated": true,
		"username":      user.Username,
		"role":          user.Role,
	})
}

func changePasswordHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if r.Method != http.MethodPost {
		http.Error(w, `{"message": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"message": "Invalid request"}`, http.StatusBadRequest)
		return
	}

	_, err := db.CheckUserCredentials(req.Username, req.OldPassword)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{"message": "Ancien mot de passe incorrect ou compte introuvable"})
		return
	}

	if len(req.NewPassword) < 3 {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"message": "Nouveau mot de passe trop court"})
		return
	}

	if err := db.UpdateUserPassword(req.Username, req.NewPassword); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"message": "Erreur lors de la mise à jour"})
		return
	}

	db.LogAction("a changé son mot de passe initial", req.Username)
	json.NewEncoder(w).Encode(map[string]string{"message": "Mot de passe mis à jour"})
}

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enableCORS(&w)
		if r.Method == "OPTIONS" {
			return
		}

		// Public exception for live metrics if setting is enabled
		if r.URL.Path == "/api/server/live-metrics" {
			publicMode := db.GetSetting("public_live_timing", "false")
			if publicMode == "true" {
				next.ServeHTTP(w, r)
				return
			}
		}

		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			http.Error(w, `{"message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		user, err := db.GetUserByToken(cookie.Value)
		if err != nil {
			http.Error(w, `{"message": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		// Inject user into context (optional, but good practice). Here we just proceed.
		// For now we just allow all authenticated users (admins, etc.)
		_ = user
		next.ServeHTTP(w, r)
	}
}

func serverStatusHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	status := process.Manager.GetStatus()

	json.NewEncoder(w).Encode(map[string]string{
		"status": string(status),
	})
}

func serverStartHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"message": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := process.Manager.Start(); err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "%v"}`, err), http.StatusBadRequest)
		return
	}
	db.LogAction("Server started", "Admin")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Server started",
	})
}

func serverStopHandler(w http.ResponseWriter, r *http.Request) {
	enableCORS(&w)
	if r.Method == "OPTIONS" {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, `{"message": "Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := process.Manager.Stop(); err != nil {
		http.Error(w, fmt.Sprintf(`{"message": "%v"}`, err), http.StatusBadRequest)
		return
	}
	db.LogAction("Server stopped", "Admin")

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Stop signal sent",
	})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for dev
	},
}

func serverConsoleHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil || cookie.Value == "" || !db.ValidateSession(cookie.Value) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	ch, history := process.Manager.SubscribeLogs()
	defer process.Manager.UnsubscribeLogs(ch)

	for _, line := range history {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			return
		}
	}

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()

	for line := range ch {
		if err := conn.WriteMessage(websocket.TextMessage, []byte(line)); err != nil {
			break
		}
	}
}

func captureFrontendLogs() {
	frontendLogPath := os.Getenv("FRONTEND_LOG_PATH")
	if frontendLogPath == "" {
		log.Println("FRONTEND_LOG_PATH is not set; skipping frontend log tail")
		return
	}
	if _, err := os.Stat(frontendLogPath); err != nil {
		log.Printf("Frontend log file %s is not available; skipping frontend log tail", frontendLogPath)
		return
	}
	cmd := exec.Command("tail", "-F", frontendLogPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Println("Error piping frontend logs:", err)
		return
	}
	if err := cmd.Start(); err != nil {
		log.Println("Error starting tail on frontend logs:", err)
		return
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			log.Printf("[VITE] %s", scanner.Text())
		}
	}()
}

func main() {
	// Start background auto-updater
	go func() {
		for {
			now := time.Now()
			// Calculate time until next 04:00 AM
			next := time.Date(now.Year(), now.Month(), now.Day(), 4, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}

			time.Sleep(time.Until(next))

			if db.GetSetting("auto_update_enabled", "false") == "true" {
				fmt.Println("[AutoUpdater] Running daily update check...")
				api.PerformAutoUpdate()
			}
		}
	}()

	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Hook up logs to broadcast to websockets
	log.SetOutput(process.BackendLogger)

	// Capture frontend logs
	captureFrontendLogs()

	if err := db.InitDB(); err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	process.InitManager()

	http.HandleFunc("/api/health", healthHandler)
	http.HandleFunc("/api/login", loginHandler)
	http.HandleFunc("/api/logout", logoutHandler)
	http.HandleFunc("/api/change-password", changePasswordHandler)
	http.HandleFunc("/api/auth/verify", verifyAuthHandler)
	http.HandleFunc("/api/server/status", AuthMiddleware(serverStatusHandler))
	http.HandleFunc("/api/server/start", AuthMiddleware(serverStartHandler))
	http.HandleFunc("/api/server/stop", AuthMiddleware(serverStopHandler))
	http.HandleFunc("/api/server/console", AuthMiddleware(serverConsoleHandler))
	http.HandleFunc("/api/server/live-metrics", AuthMiddleware(api.LiveMetricsHandler))
	http.HandleFunc("/api/config/general", AuthMiddleware(api.ConfigGeneralHandler))
	http.HandleFunc("/api/config/event", AuthMiddleware(api.ConfigEventHandler))
	http.HandleFunc("/api/config/ai", AuthMiddleware(api.ConfigAIHandler))

	http.HandleFunc("/api/plugins/available", AuthMiddleware(api.GetAvailablePluginsHandler))
	http.HandleFunc("/api/plugins/config", AuthMiddleware(api.PluginConfigHandler))
	http.HandleFunc("/api/plugins/custom", AuthMiddleware(api.CustomPluginsHandler))
	http.HandleFunc("/api/plugins/custom/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			api.CustomPluginDeleteHandler(w, r)
		} else {
			api.CustomPluginsHandler(w, r)
		}
	}))

	http.HandleFunc("/api/server/upgrade", AuthMiddleware(api.ServerUpgradeHandler))
	http.HandleFunc("/api/server/upgrade/zip", AuthMiddleware(api.ServerUpgradeZipHandler))
	http.HandleFunc("/api/upgrade/check", AuthMiddleware(api.CheckUpdateHandler))

	http.HandleFunc("/api/content/upload", AuthMiddleware(api.UploadContentHandler))
	http.HandleFunc("/downloads/", api.DownloadHandler)
	http.HandleFunc("/api/content/tracks/upload-fastlane", AuthMiddleware(api.UploadFastLaneHandler))
	http.HandleFunc("/api/content/cars", AuthMiddleware(api.GetCarsHandler))
	http.HandleFunc("/api/content/cars/image", AuthMiddleware(api.GetCarImageHandler))
	http.HandleFunc("/api/content/cars/skin_image", AuthMiddleware(api.GetCarSkinImageHandler))
	http.HandleFunc("/api/content/tracks", AuthMiddleware(api.GetTracksHandler))
	http.HandleFunc("/api/content/tracks/image", AuthMiddleware(api.GetTrackImageHandler))
	http.HandleFunc("/api/content/tracks/layout_image", AuthMiddleware(api.GetTrackLayoutImageHandler))

	http.HandleFunc("/api/logs/backend", AuthMiddleware(api.StreamBackendLogsHandler))
	http.HandleFunc("/api/logs/acsm", AuthMiddleware(api.StreamACSMLogsHandler))
	http.HandleFunc("/api/logs/audit", AuthMiddleware(api.StreamAuditLogsHandler))

	http.HandleFunc("/api/presets", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			api.CreatePresetHandler(w, r)
		} else {
			api.GetPresetsHandler(w, r)
		}
	}))
	http.HandleFunc("/api/presets/", AuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/launch") {
			api.LaunchPresetHandler(w, r)
			return
		}
		if r.Method == "PUT" {
			api.UpdatePresetHandler(w, r)
		} else if r.Method == "DELETE" {
			api.DeletePresetHandler(w, r)
		}
	}))

	http.HandleFunc("/api/users", AuthMiddleware(api.UsersHandler))
	http.HandleFunc("/api/settings", AuthMiddleware(api.SettingsHandler))

	port := ":8080"
	log.Printf("Starting backend server on port %s\n", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed to start: %v\n", err)
	}
}
