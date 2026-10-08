package ai

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
 )

// FileConfig adalah konfigurasi pengguna yang dapat disimpan di config.json.
// Rahasia seperti kunci API sengaja tidak termasuk dalam format ini.
type FileConfig struct {
	AI FileAIConfig `json:"ai"`
}

type FileAIConfig struct {
	URL string `json:"url"`
	Model string `json:"model"`
}

// ConfigFromFile membaca konfigurasi pengguna dari path.
// Berkas yang belum ada berarti AI belum dikonfigurasi.
func ConfigFromFile(path string) (*Config, string, bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) { return nil, "", false, nil }
		return nil, "", false, fmt.Errorf("membaca konfigurasi: %w", err)
	}
	var file FileConfig
	decoder := json.NewDecoder(strings.NewReader(string(data)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&file); err != nil { return nil, "", false, fmt.Errorf("konfigurasi JSON tidak valid: %w", err) }
	if strings.TrimSpace(file.AI.URL) == "" { return nil, strings.TrimSpace(file.AI.Model), false, nil }
	cfg := Config{BaseURL: file.AI.URL}
	if _, err := NewClient(cfg); err != nil { return nil, "", false, fmt.Errorf("konfigurasi penyedia model: %w", err) }
	return &cfg, strings.TrimSpace(file.AI.Model), true, nil
}

// ConfigFromEnv membaca konfigurasi provider lokal dari lingkungan proses.
// Jika CATATAN_AI_URL kosong, AI dianggap tidak dikonfigurasi.
func ConfigFromEnv() (*Config, bool, error) {
	baseURL := strings.TrimSpace(os.Getenv("CATATAN_AI_URL"))
	if baseURL == "" { return nil, false, nil }
	cfg := Config{BaseURL: baseURL, APIKey: os.Getenv("CATATAN_AI_API_KEY")}
	if _, err := NewClient(cfg); err != nil { return nil, false, fmt.Errorf("konfigurasi penyedia model: %w", err) }
	return &cfg, true, nil
}

// ModelFromEnv membaca nama model dari lingkungan proses.
func ModelFromEnv() string { return strings.TrimSpace(os.Getenv("CATATAN_AI_MODEL")) }
