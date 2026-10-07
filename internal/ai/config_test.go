package ai

import (
	"os"
	"path/filepath"
	"testing"
 )

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("CATATAN_AI_URL", "")
	t.Setenv("CATATAN_AI_API_KEY", "")
	cfg, enabled, err := ConfigFromEnv()
	if err != nil { t.Fatal(err) }
	if enabled || cfg != nil { t.Fatalf("konfigurasi kosong = %#v, aktif = %v", cfg, enabled) }
	t.Setenv("CATATAN_AI_URL", "http://127.0.0.1:11434")
	t.Setenv("CATATAN_AI_API_KEY", "rahasia-uji")
	cfg, enabled, err = ConfigFromEnv()
	if err != nil { t.Fatal(err) }
	if !enabled || cfg == nil { t.Fatal("konfigurasi lokal tidak aktif") }
	if cfg.BaseURL != "http://127.0.0.1:11434" || cfg.APIKey != "rahasia-uji" { t.Fatalf("konfigurasi = %#v", cfg) }
	t.Setenv("CATATAN_AI_URL", "http://127.0.0.1:11434/v1")
	if _, _, err := ConfigFromEnv(); err == nil { t.Fatal("alamat dengan /v1 seharusnya ditolak") }
	t.Setenv("CATATAN_AI_URL", "file:///rahasia")
	if _, _, err := ConfigFromEnv(); err == nil { t.Fatal("alamat dengan skema tidak valid seharusnya ditolak") }
}

func TestConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if _, _, enabled, err := ConfigFromFile(path); err != nil || enabled { t.Fatalf("berkas tidak ada: enabled=%v err=%v", enabled, err) }
	if err := os.WriteFile(path, []byte(`{"ai":{"url":"http://127.0.0.1:11434","model":"llama3.2"}}`), 0600); err != nil { t.Fatal(err) }
	cfg, model, enabled, err := ConfigFromFile(path)
	if err != nil || !enabled || cfg == nil { t.Fatalf("konfigurasi berkas: cfg=%#v model=%q enabled=%v err=%v", cfg, model, enabled, err) }
	if cfg.BaseURL != "http://127.0.0.1:11434" || model != "llama3.2" || cfg.APIKey != "" { t.Fatalf("konfigurasi = %#v model=%q", cfg, model) }
	if err := os.WriteFile(path, []byte(`{"ai":{"url":"http://127.0.0.1:11434","rahasia":"jangan"}}`), 0600); err != nil { t.Fatal(err) }
	if _, _, _, err := ConfigFromFile(path); err == nil { t.Fatal("field konfigurasi yang tidak dikenal seharusnya ditolak") }
	if err := os.WriteFile(path, []byte("{"), 0600); err != nil { t.Fatal(err) }
	if _, _, _, err := ConfigFromFile(path); err == nil { t.Fatal("JSON rusak seharusnya ditolak") }
}

func TestModelFromEnv(t *testing.T) { t.Setenv("CATATAN_AI_MODEL", " llama3.2 "); if got := ModelFromEnv(); got != "llama3.2" { t.Fatalf("model=%q", got) } }
