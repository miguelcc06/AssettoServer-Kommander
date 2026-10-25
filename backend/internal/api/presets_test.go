package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestApplyPresetConfigSyncsTrackAndKeepsAssists(t *testing.T) {
	root := t.TempDir()
	cfg := filepath.Join(root, "cfg")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	ini := filepath.Join(cfg, "server_cfg.ini")
	if err := os.WriteFile(ini, []byte("[SERVER]\nNAME=Kept Name\nADMIN_PASSWORD=secret\nHTTP_PORT=8081\nTCP_PORT=9600\nUDP_PORT=9600\nTRACK=imola\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmPath := filepath.Join(root, "cm_content", "content.json")
	if err := os.MkdirAll(filepath.Dir(cmPath), 0o755); err != nil {
		t.Fatal(err)
	}
	original := `{
	  "cars": {"ks_alfa": {"url": "http://example.test/downloads/ks_alfa.zip", "version": "1"}},
	  "tracks": {
	    "imola": {"url": "http://example.test/downloads/imola.zip", "version": "9"},
	    "spa": {"url": "http://example.test/downloads/spa.zip", "version": "4", "note": "keep"}
	  },
	  "track": {"url": "http://example.test/downloads/imola.zip", "version": "9"}
	}`
	if err := os.WriteFile(cmPath, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Setenv("AC_SERVER_INI_PATH", ini)
	t.Setenv("AC_CM_CONTENT_JSON", cmPath)

	preset := &EventPreset{
		Name:          "Spa evening",
		Track:         "spa",
		TrackLayout:   "layout_gp",
		AvailableCars: []string{"ks_alfa", "ks_bmw"},
		MaxClients:    18,
		Password:      "race",
		LoopMode:      true,
		Weather: WeatherPreset{
			Graphics:               "Clear",
			BaseTemperatureAmbient: 22,
			BaseTemperatureRoad:    30,
			VariationAmbient:       1,
			VariationRoad:          1,
			WindBaseSpeedMin:       3,
			WindBaseSpeedMax:       15,
			WindBaseDirection:      0,
			WindVariationDirection: 10,
		},
		Sessions: SessionsPreset{
			Practice: SessionConfig{Enabled: true, Time: 30},
		},
		EntryList: []EntryListSlot{{Car: "ks_alfa", Skin: "red", Name: "A"}},
	}
	if err := applyPresetConfig(preset); err != nil {
		t.Fatal(err)
	}

	iniData, err := os.ReadFile(ini)
	if err != nil {
		t.Fatal(err)
	}
	text := string(iniData)
	for _, line := range []string{
		"NAME=Kept Name",
		"ADMIN_PASSWORD=secret",
		"TRACK=spa",
		"CONFIG_TRACK=layout_gp",
		"CARS=ks_alfa;ks_bmw",
		"PASSWORD=race",
		"MAX_CLIENTS=18",
		"LOOP_MODE=1",
		"ABS_ALLOWED=1",
		"TC_ALLOWED=1",
		"STABILITY_ALLOWED=1",
		"AUTOCLUTCH_ALLOWED=1",
		"TYRE_BLANKETS_ALLOWED=1",
		"[PRACTICE]",
		"TIME=30",
	} {
		if !strings.Contains(text, line) {
			t.Fatalf("server_cfg.ini missing %q:\n%s", line, text)
		}
	}

	data, err := os.ReadFile(cmPath)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatal(err)
	}
	track := doc["track"].(map[string]any)
	if track["url"] != "http://example.test/downloads/spa.zip" || track["version"] != "4" || track["note"] != "keep" {
		t.Fatalf("track = %#v", track)
	}
	tracks := doc["tracks"].(map[string]any)
	imola := tracks["imola"].(map[string]any)
	if imola["url"] != "http://example.test/downloads/imola.zip" {
		t.Fatalf("imola dropped: %#v", tracks)
	}
	cars := doc["cars"].(map[string]any)
	if cars["ks_alfa"].(map[string]any)["version"] != "1" {
		t.Fatalf("cars = %#v", cars)
	}

	entry, err := os.ReadFile(filepath.Join(cfg, "entry_list.ini"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(entry), "MODEL=ks_alfa") || !strings.Contains(string(entry), "SKIN=red") {
		t.Fatalf("entry_list.ini = %s", entry)
	}
}

func TestApplyPresetConfigReportsBrokenContentJSON(t *testing.T) {
	root := t.TempDir()
	cfg := filepath.Join(root, "cfg")
	if err := os.MkdirAll(cfg, 0o755); err != nil {
		t.Fatal(err)
	}
	ini := filepath.Join(cfg, "server_cfg.ini")
	t.Setenv("AC_SERVER_INI_PATH", ini)
	cmPath := filepath.Join(root, "cm_content", "content.json")
	t.Setenv("AC_CM_CONTENT_JSON", cmPath)
	if err := os.MkdirAll(filepath.Dir(cmPath), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cmPath, []byte("{"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := applyPresetConfig(&EventPreset{Track: "spa", Name: "broken"})
	if err == nil || !strings.Contains(err.Error(), "Content Manager track") {
		t.Fatalf("expected sync error, got %v", err)
	}
}
