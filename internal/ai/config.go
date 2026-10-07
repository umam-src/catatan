package ai

import (
	"fmt"
	"os"
	"strings"
)

// ConfigFromEnv membaca konfigurasi provider lokal dari lingkungan proses.
// Jika CATATAN_AI_URL kosong, AI dianggap tidak dikonfigurasi.
func ConfigFromEnv() (*Config, bool, error) {
	baseURL := strings.TrimSpace(os.Getenv("CATATAN_AI_URL"))
	if baseURL == "" {
		return nil, false, nil
	}
	cfg := Config{
		BaseURL: baseURL,
		APIKey:  os.Getenv("CATATAN_AI_API_KEY"),
	}
	if model := strings.TrimSpace(os.Getenv("CATATAN_AI_MODEL")); model != "" {
		_ = model
	}
	if _, err := NewClient(cfg); err != nil {
		return nil, false, fmt.Errorf("konfigurasi penyedia model: %w", err)
	}
	return &cfg, true, nil
}
