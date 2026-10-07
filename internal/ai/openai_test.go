package ai

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestOpenAICompatibleProviderGeneratesWithoutExternalNetwork(t *testing.T) {
	var gotAuthorization string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" {
			t.Fatalf("path = %q", r.URL.Path)
		}
		gotAuthorization = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"uji-model","choices":[{"message":{"content":"jawaban lokal"}}]}`))
	}))
	defer server.Close()

	provider, err := NewClient(Config{BaseURL: server.URL, APIKey: "test-key"})
	if err != nil { t.Fatal(err) }
	response, err := provider.Generate(context.Background(), Request{
		Model: "uji-model",
		Messages: []Message{{Role: RoleUser, Content: "Halo"}},
	})
	if err != nil { t.Fatal(err) }
	if response.Text != "jawaban lokal" || response.Model != "uji-model" {
		t.Fatalf("response = %#v", response)
	}
	if gotAuthorization != "Bearer test-key" {
		t.Fatalf("authorization = %q", gotAuthorization)
	}
}

func TestOpenAICompatibleProviderRejectsInvalidResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"uji-model","choices":[]}`))
	}))
	defer server.Close()

	provider, err := NewClient(Config{BaseURL: server.URL})
	if err != nil { t.Fatal(err) }
	_, err = provider.Generate(context.Background(), Request{
		Model: "uji-model",
		Messages: []Message{{Role: RoleUser, Content: "Halo"}},
	})
	if !errors.Is(err, ErrInvalidResponse) {
		t.Fatalf("error = %v, ingin ErrInvalidResponse", err)
	}
}

func TestOpenAICompatibleProviderClassifiesUnavailableAndTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test-Mode") == "timeout" {
			<-r.Context().Done()
			return
		}
		http.Error(w, "gagal", http.StatusBadGateway)
	}))
	defer server.Close()

	provider, err := NewClient(Config{BaseURL: server.URL, Timeout: 20 * time.Millisecond})
	if err != nil { t.Fatal(err) }
	_, err = provider.Generate(context.Background(), Request{Model: "uji", Messages: []Message{{Role: RoleUser, Content: "Halo"}}})
	if !errors.Is(err, ErrProviderUnavailable) {
		t.Fatalf("error gateway = %v", err)
	}

	provider.httpClient = &http.Client{Transport: roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.Header.Set("X-Test-Mode", "timeout")
		return http.DefaultTransport.RoundTrip(req)
	})}
	_, err = provider.Generate(context.Background(), Request{Model: "uji", Messages: []Message{{Role: RoleUser, Content: "Halo"}}})
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("error timeout = %v", err)
	}
}

func TestOpenAICompatibleProviderEnforcesSizeLimits(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"model":"uji","choices":[{"message":{"content":"jawaban panjang"}}]}`))
	}))
	defer server.Close()

	provider, err := NewClient(Config{BaseURL: server.URL, MaxRequestBytes: 64})
	if err != nil { t.Fatal(err) }
	_, err = provider.Generate(context.Background(), Request{
		Model: "uji", Messages: []Message{{Role: RoleUser, Content: strings.Repeat("x", 100)}},
	})
	if !errors.Is(err, ErrRequestTooLarge) {
		t.Fatalf("error request = %v", err)
	}

	provider.maxRequestBytes = 1 << 20
	provider.maxResponseBytes = 16
	_, err = provider.Generate(context.Background(), Request{
		Model: "uji", Messages: []Message{{Role: RoleUser, Content: "Halo"}},
	})
	if !errors.Is(err, ErrResponseTooLarge) {
		t.Fatalf("error response = %v", err)
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
