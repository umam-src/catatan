package http

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/umam-src/catatan/internal/ai"
	"github.com/umam-src/catatan/internal/db"
)

type fakeAIStatusProvider struct {
	status ai.ProviderStatus
}

func (f fakeAIStatusProvider) Probe(context.Context) ai.ProviderStatus {
	return f.status
}

func TestAIStatusDisabledWithoutProvider(t *testing.T) {
	handler := serverUji(t)
	res := requestUji(t, handler, http.MethodGet, "/api/ai/status", nil)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d", res.Code, http.StatusOK)
	}
	var out struct {
		Enabled bool   `json:"enabled"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out.Enabled || out.Status != "disabled" {
		t.Fatalf("response = %#v", out)
	}
}

func TestAIStatusUsesConfiguredProvider(t *testing.T) {
	path := filepath.Join(t.TempDir(), "catatan.db")
	d, err := db.Open(context.Background(), path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = d.Close() })

	handler := NewWithAI(d, fakeAIStatusProvider{status: ai.ProviderReady})
	setup := requestUji(t, handler, http.MethodPost, "/api/auth/setup", map[string]string{
		"username": "pengguna",
		"password": "kata-sandi-uji-aman",
	})
	if setup.Code != http.StatusCreated {
		t.Fatalf("setup pengguna: status = %d", setup.Code)
	}
	cookie := setup.Result().Cookies()[0]
	res := requestDenganCookie(t, handler, http.MethodGet, "/api/ai/status", nil, cookie)
	if res.Code != http.StatusOK {
		t.Fatalf("status = %d, ingin %d", res.Code, http.StatusOK)
	}
	var out struct {
		Enabled bool   `json:"enabled"`
		Status  string `json:"status"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if !out.Enabled || out.Status != string(ai.ProviderReady) {
		t.Fatalf("response = %#v", out)
	}
}
