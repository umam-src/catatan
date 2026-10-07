package ai

import "testing"

func TestConfigFromEnv(t *testing.T) {
	t.Setenv("CATATAN_AI_URL", "")
	t.Setenv("CATATAN_AI_API_KEY", "")
	cfg, enabled, err := ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if enabled || cfg != nil {
		t.Fatalf("konfigurasi kosong = %#v, aktif = %v", cfg, enabled)
	}

	t.Setenv("CATATAN_AI_URL", "http://127.0.0.1:11434/v1")
	t.Setenv("CATATAN_AI_API_KEY", "rahasia-uji")
	cfg, enabled, err = ConfigFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if !enabled || cfg == nil {
		t.Fatal("konfigurasi lokal tidak aktif")
	}
	if cfg.BaseURL != "http://127.0.0.1:11434/v1" || cfg.APIKey != "rahasia-uji" {
		t.Fatalf("konfigurasi = %#v", cfg)
	}

	t.Setenv("CATATAN_AI_URL", "file:///rahasia")
	if _, _, err := ConfigFromEnv(); err == nil {
		t.Fatal("alamat dengan skema tidak valid seharusnya ditolak")
	}
}
