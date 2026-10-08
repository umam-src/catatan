package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAICompatibleProviderProbe(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       string
		want       ProviderStatus
	}{
		{name: "siap", statusCode: http.StatusOK, body: `{"data":[{"id":"uji-model"}]}`, want: ProviderReady},
		{name: "tidak tersedia", statusCode: http.StatusServiceUnavailable, body: `{}`, want: ProviderUnavailable},
		{name: "respons tidak sah", statusCode: http.StatusOK, body: `{"data":"salah"}`, want: ProviderInvalid},
		{name: "json tidak sah", statusCode: http.StatusOK, body: `bukan-json`, want: ProviderInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet || r.URL.Path != "/v1/models" {
					t.Fatalf("request = %s %s", r.Method, r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			provider, err := NewClient(Config{BaseURL: server.URL})
			if err != nil {
				t.Fatal(err)
			}
			if got := provider.Probe(context.Background()); got != tt.want {
				t.Fatalf("status = %q, ingin %q", got, tt.want)
			}
		})
	}
}

func TestOpenAICompatibleProviderProbeRejectsEmptyModels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/v1/models" {
			t.Fatalf("request = %s %s", r.Method, r.URL.Path)
		}
		if r.ContentLength != 0 {
			t.Fatalf("content length = %d", r.ContentLength)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	provider, err := NewClient(Config{BaseURL: server.URL})
	if err != nil {
		t.Fatal(err)
	}
	if got := provider.Probe(context.Background()); got != ProviderInvalid {
		t.Fatalf("status = %q, ingin %q", got, ProviderReady)
	}
}
