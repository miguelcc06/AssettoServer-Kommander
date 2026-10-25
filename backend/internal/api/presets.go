package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"web-gui-acserver/internal/configparser"
	"web-gui-acserver/internal/db"
	"web-gui-acserver/internal/process"
	"web-gui-acserver/internal/services"
)

type WeatherPreset struct {
	Graphics               string `json:"graphics"`
	BaseTemperatureAmbient int    `json:"baseTemperatureAmbient"`
	BaseTemperatureRoad    int    `json:"baseTemperatureRoad"`
	VariationAmbient       int    `json:"variationAmbient"`
	VariationRoad          int    `json:"variationRoad"`
	WindBaseSpeedMin       int    `json:"windBaseSpeedMin"`
	WindBaseSpeedMax       int    `json:"windBaseSpeedMax"`
	WindBaseDirection      int    `json:"windBaseDirection"`
	WindVariationDirection int    `json:"windVariationDirection"`
}

type EntryListSlot struct {
	Car  string `json:"car"`
	Skin string `json:"skin"`
	Name string `json:"name"`
	Team string `json:"team"`
	GUID string `json:"guid"`
}

type TrafficConfig struct {
	Car   string `json:"car"`
	Skin  string `json:"skin"`
	Count int    `json:"count"`
}

type SessionConfig struct {
	Enabled bool `json:"enabled"`
	Time    int  `json:"time"`
	Laps    int  `json:"laps,omitempty"`
}

type SessionsPreset struct {
	Practice SessionConfig `json:"practice"`
	Qualify  SessionConfig `json:"qualify"`
	Race     SessionConfig `json:"race"`
}

type EventPreset struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	Track           string          `json:"track"`
	TrackLayout     string          `json:"trackLayout"`
	AvailableCars   []string        `json:"availableCars"`
	MaxClients      int             `json:"maxClients"`
	Password        string          `json:"password"`
	LoopMode        bool            `json:"loopMode"`
	Weather         WeatherPreset   `json:"weather"`
	Sessions        SessionsPreset  `json:"sessions"`
	EntryList       []EntryListSlot `json:"entryList"`
	Traffic         []TrafficConfig `json:"traffic"`
	AiMaxCars       int             `json:"aiMaxCars"`
	AiMinDistance   int             `json:"aiMinDistance"`
	CspExtraOptions string          `json:"cspExtraOptions,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`
}

func getPresetsPath() string {
	path := os.Getenv("AC_PRESETS_PATH")
	if path == "" {
		path = "data/presets.json"
	}
	return path
}

func loadPresets() ([]EventPreset, error) {
	path := getPresetsPath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return []EventPreset{}, nil
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var presets []EventPreset
	err = json.Unmarshal(data, &presets)
	if err != nil {
		return nil, err
	}

	return presets, nil
}

func savePresets(presets []EventPreset) error {
	path := getPresetsPath()
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)

	data, err := json.MarshalIndent(presets, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(path, data, 0644)
}

func GetPresetsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if r.Method == "OPTIONS" {
		return
	}

	presets, err := loadPresets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(presets)
}

func CreatePresetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	var preset EventPreset
	err := json.NewDecoder(r.Body).Decode(&preset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	preset.ID = uuid.New().String()
	preset.CreatedAt = time.Now()
	preset.UpdatedAt = time.Now()

	presets, err := loadPresets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	presets = append(presets, preset)

	err = savePresets(presets)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.LogAction("Created event '"+preset.Name+"'", "Admin")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(preset)
}

func UpdatePresetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]

	var updatedPreset EventPreset
	err := json.NewDecoder(r.Body).Decode(&updatedPreset)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	presets, err := loadPresets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, p := range presets {
		if p.ID == id {
			updatedPreset.ID = id
			updatedPreset.CreatedAt = p.CreatedAt
			updatedPreset.UpdatedAt = time.Now()
			presets[i] = updatedPreset
			err = savePresets(presets)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			db.LogAction("Updated event '"+updatedPreset.Name+"'", "Admin")

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedPreset)
			return
		}
	}

	http.Error(w, "Preset not found", http.StatusNotFound)
}

func DeletePresetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-1]

	presets, err := loadPresets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	for i, p := range presets {
		if p.ID == id {
			presets = append(presets[:i], presets[i+1:]...)
			err = savePresets(presets)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			db.LogAction("Deleted event ID '"+id+"'", "Admin")

			w.WriteHeader(http.StatusOK)
			return
		}
	}

	http.Error(w, "Preset not found", http.StatusNotFound)
}

func LaunchPresetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")

	if r.Method == "OPTIONS" {
		return
	}

	parts := strings.Split(r.URL.Path, "/")
	id := parts[len(parts)-2] // /api/presets/:id/launch

	presets, err := loadPresets()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetPreset *EventPreset
	for _, p := range presets {
		if p.ID == id {
			targetPreset = &p
			break
		}
	}

	if targetPreset == nil {
		http.Error(w, "Preset not found", http.StatusNotFound)
		return
	}

	if err := applyPresetConfig(targetPreset); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Restart Assetto Corsa after the configs and the Content Manager track
	// entry match the selected event.
	err = process.Manager.Restart()
	if err != nil {
		http.Error(w, "Failed to restart server: "+err.Error(), http.StatusInternalServerError)
		return
	}

	db.LogAction("Launched event '"+targetPreset.Name+"'", "Admin")

	w.WriteHeader(http.StatusOK)
}

func applyPresetConfig(targetPreset *EventPreset) error {
	// 1. Generate server_cfg.ini
	cfgPath := os.Getenv("AC_SERVER_INI_PATH")
	if cfgPath == "" {
		cfgPath = "data/cfg/server_cfg.ini"
	}

	// Preserve existing settings
	currName := configparser.ReadINI(cfgPath, "SERVER", "NAME")
	if currName == "" {
		currName = "Web GUI Server"
	}
	currAdmin := configparser.ReadINI(cfgPath, "SERVER", "ADMIN_PASSWORD")
	if currAdmin == "" {
		currAdmin = "adminpass"
	}
	currHTTP := configparser.ReadINI(cfgPath, "SERVER", "HTTP_PORT")
	if currHTTP == "" {
		currHTTP = "8081"
	}
	currTCP := configparser.ReadINI(cfgPath, "SERVER", "TCP_PORT")
	if currTCP == "" {
		currTCP = "9600"
	}
	currUDP := configparser.ReadINI(cfgPath, "SERVER", "UDP_PORT")
	if currUDP == "" {
		currUDP = "9600"
	}

	carsStr := strings.Join(targetPreset.AvailableCars, ";")

	var sessionsBuilder strings.Builder
	if targetPreset.Sessions.Practice.Enabled {
		sessionsBuilder.WriteString(fmt.Sprintf("[PRACTICE]\nNAME=Free Practice\nTIME=%d\nIS_OPEN=1\n\n", targetPreset.Sessions.Practice.Time))
	}
	if targetPreset.Sessions.Qualify.Enabled {
		sessionsBuilder.WriteString(fmt.Sprintf("[QUALIFY]\nNAME=Qualify\nTIME=%d\nIS_OPEN=1\n\n", targetPreset.Sessions.Qualify.Time))
	}
	if targetPreset.Sessions.Race.Enabled {
		sessionsBuilder.WriteString(fmt.Sprintf("[RACE]\nNAME=Race\nLAPS=%d\nTIME=%d\nWAIT_TIME=60\nIS_OPEN=1\n\n", targetPreset.Sessions.Race.Laps, targetPreset.Sessions.Race.Time))
	}

	loopModeInt := 0
	if targetPreset.LoopMode {
		loopModeInt = 1
	}

	serverCfgContent := fmt.Sprintf(`[SERVER]
NAME=%s
CARS=%s
TRACK=%s
CONFIG_TRACK=%s
PASSWORD=%s
ADMIN_PASSWORD=%s
UDP_PORT=%s
TCP_PORT=%s
HTTP_PORT=%s
CLIENT_SEND_INTERVAL_HZ=18
SEND_BUFFER_SIZE=0
RECV_BUFFER_SIZE=0
TICK_RATE=18
DISCONNECT_SECONDS=200
REGISTER_TO_LOBBY=1
MAX_CLIENTS=%d
UDP_PLUGIN_LOCAL_PORT=0
UDP_PLUGIN_ADDRESS=
AUTH_PLUGIN_ADDRESS=
LEGAL_TYRES=
RACE_PIT_WINDOW_START=0
RACE_PIT_WINDOW_END=0
REVERSED_GRID_RACE_POSITIONS=0
QUALIFY_MAX_WAIT_PERC=120
RACE_GAS_PENALTY_DISABLED=0
MAX_BALLAST_KG=50
MAX_RESTRICTOR=100
RESULT_SCREEN_TIME=10
RACE_EXTRA_LAP=0
LOCKED_ENTRY_LIST=1
LOOP_MODE=%d
NUM_THREADS=2
ALLOWED_TYRES_OUT=-1
ABS_ALLOWED=1
TC_ALLOWED=1
STABILITY_ALLOWED=1
AUTOCLUTCH_ALLOWED=1
TYRE_BLANKETS_ALLOWED=1

%s
[DYNAMIC_TRACK]
SESSION_START=90
RANDOMNESS=2
LAP_GAIN=100
SESSION_TRANSFER=80

[WEATHER_0]
GRAPHICS=%s
BASE_TEMPERATURE_AMBIENT=%d
BASE_TEMPERATURE_ROAD=%d
VARIATION_AMBIENT=%d
VARIATION_ROAD=%d
WIND_BASE_SPEED_MIN=%d
WIND_BASE_SPEED_MAX=%d
WIND_BASE_DIRECTION=%d
WIND_VARIATION_DIRECTION=%d
`,
		currName, carsStr, targetPreset.Track, targetPreset.TrackLayout, targetPreset.Password, currAdmin, currUDP, currTCP, currHTTP, targetPreset.MaxClients, loopModeInt,
		sessionsBuilder.String(),
		targetPreset.Weather.Graphics, targetPreset.Weather.BaseTemperatureAmbient, targetPreset.Weather.BaseTemperatureRoad,
		targetPreset.Weather.VariationAmbient, targetPreset.Weather.VariationRoad, targetPreset.Weather.WindBaseSpeedMin,
		targetPreset.Weather.WindBaseSpeedMax, targetPreset.Weather.WindBaseDirection, targetPreset.Weather.WindVariationDirection)

	err := ioutil.WriteFile(cfgPath, []byte(serverCfgContent), 0644)
	if err != nil {
		return fmt.Errorf("failed to write server_cfg.ini: %w", err)
	}

	// 2. Generate entry_list.ini
	entryListPath := filepath.Join(filepath.Dir(cfgPath), "entry_list.ini")

	var entryListBuilder strings.Builder
	slotIndex := 0

	// Human slots
	for _, slot := range targetPreset.EntryList {
		entryListBuilder.WriteString(fmt.Sprintf(`[CAR_%d]
MODEL=%s
SKIN=%s
SPECTATOR_MODE=0
DRIVERNAME=%s
TEAM=%s
GUID=%s
BALLAST=0
RESTRICTOR=0
AI=none

`, slotIndex, slot.Car, slot.Skin, slot.Name, slot.Team, slot.GUID))
		slotIndex++
	}

	// Traffic AI slots
	for _, traffic := range targetPreset.Traffic {
		for i := 0; i < traffic.Count; i++ {
			entryListBuilder.WriteString(fmt.Sprintf(`[CAR_%d]
MODEL=%s
SKIN=%s
SPECTATOR_MODE=0
DRIVERNAME=
TEAM=
GUID=
BALLAST=0
RESTRICTOR=0
AI=fixed

`, slotIndex, traffic.Car, traffic.Skin))
			slotIndex++
		}
	}

	err = ioutil.WriteFile(entryListPath, []byte(entryListBuilder.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write entry_list.ini: %w", err)
	}

	// 3. Generate extra_cfg.yml safely (without overwriting plugins)
	extraCfgPath := filepath.Join(filepath.Dir(cfgPath), "extra_cfg.yml")
	enableAiStr := "false"
	if len(targetPreset.Traffic) > 0 {
		enableAiStr = "true"
	}

	existingExtraCfg, err := ioutil.ReadFile(extraCfgPath)
	if err != nil {
		// If file doesn't exist, write the default one
		extraCfgContent := fmt.Sprintf(`EnableSteamAuth: false
ValidateDlcOwnership: []
MandatoryClientSecurityLevel: 0
EnableAntiAfk: false
MaxAfkTimeMinutes: 10
AfkKickBehavior: PlayerInput
MaxPing: 500
MaxPingSeconds: 10
ForceLights: false
NetworkBubbleDistance: 500
OutsideNetworkBubbleRefreshRateHz: 4
EnableServerDetails: true
ServerDescription: 'Web GUI Generated Server'
EnableRealTime: false
EnableWeatherFx: true
EnableClientMessages: true
EnableUdpClientMessages: false
DebugClientMessages: false
EnableCustomUpdate: false
EnableAi: %s
AiParams:
  MaxAiTargetCount: %d
  AiPerPlayerTargetCount: %d
  PlayerRadiusMeters: %d
IgnoreConfigurationErrors:
  MissingCarChecksums: true
  MissingTrackParams: true
`, enableAiStr, targetPreset.AiMaxCars, targetPreset.AiMaxCars, targetPreset.AiMinDistance)
		ioutil.WriteFile(extraCfgPath, []byte(extraCfgContent), 0644)
	} else {
		// Use regex to update AI settings while preserving plugins and comments
		contentStr := string(existingExtraCfg)

		reEnable := regexp.MustCompile(`(?m)^EnableAi:\s*(true|false)`)
		contentStr = reEnable.ReplaceAllString(contentStr, fmt.Sprintf("EnableAi: %s", enableAiStr))

		reMaxAi := regexp.MustCompile(`(?m)^\s*MaxAiTargetCount:\s*\d+`)
		contentStr = reMaxAi.ReplaceAllString(contentStr, fmt.Sprintf("  MaxAiTargetCount: %d", targetPreset.AiMaxCars))

		reAiPerPlayer := regexp.MustCompile(`(?m)^\s*AiPerPlayerTargetCount:\s*\d+`)
		contentStr = reAiPerPlayer.ReplaceAllString(contentStr, fmt.Sprintf("  AiPerPlayerTargetCount: %d", targetPreset.AiMaxCars))

		rePlayerRadius := regexp.MustCompile(`(?m)^\s*PlayerRadiusMeters:\s*\d+`)
		contentStr = rePlayerRadius.ReplaceAllString(contentStr, fmt.Sprintf("  PlayerRadiusMeters: %d", targetPreset.AiMinDistance))

		ioutil.WriteFile(extraCfgPath, []byte(contentStr), 0644)
	}

	// Write csp_extra_options.ini
	cspOptionsPath := filepath.Join(filepath.Dir(cfgPath), "csp_extra_options.ini")
	if targetPreset.CspExtraOptions != "" {
		ioutil.WriteFile(cspOptionsPath, []byte(targetPreset.CspExtraOptions), 0644)
	} else {
		// If empty, we can just write empty or delete. Let's write empty to clear old settings.
		ioutil.WriteFile(cspOptionsPath, []byte(""), 0644)
	}

	if err := services.SyncActiveTrack(targetPreset.Track); err != nil {
		return fmt.Errorf("failed to sync Content Manager track: %w", err)
	}
	return nil
}
