package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout          = 30 * time.Second
	defaultMaxRequestBytes  = 1 << 20
	defaultMaxResponseBytes = 1 << 20
	defaultProbeTimeout     = 2 * time.Second
)

type ProviderStatus string

const (
	ProviderReady       ProviderStatus = "ready"
	ProviderUnavailable ProviderStatus = "unavailable"
	ProviderInvalid     ProviderStatus = "invalid"
)

type Config struct {
	BaseURL          string
	APIKey           string
	Timeout          time.Duration
	MaxRequestBytes  int64
	MaxResponseBytes int64
	HTTPClient       *http.Client
}

type Client struct {
	baseURL          *url.URL
	apiKey           string
	timeout          time.Duration
	maxRequestBytes  int64
	maxResponseBytes int64
	httpClient       *http.Client
}

func NewClient(cfg Config) (*Client, error) {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		return nil, fmt.Errorf("alamat penyedia model wajib diisi")
	}
	u, err := url.Parse(base)
	if err != nil || u.Scheme == "" || u.Host == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" {
		return nil, fmt.Errorf("alamat penyedia model tidak valid")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("skema alamat penyedia model harus http atau https")
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	maxRequest := cfg.MaxRequestBytes
	if maxRequest <= 0 {
		maxRequest = defaultMaxRequestBytes
	}
	maxResponse := cfg.MaxResponseBytes
	if maxResponse <= 0 {
		maxResponse = defaultMaxResponseBytes
	}
	if maxRequest < 1 || maxResponse < 1 {
		return nil, fmt.Errorf("batas ukuran harus positif")
	}
	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{
		baseURL:          u,
		apiKey:           cfg.APIKey,
		timeout:          timeout,
		maxRequestBytes:  maxRequest,
		maxResponseBytes: maxResponse,
		httpClient:       httpClient,
	}, nil
}

// Probe memeriksa kesiapan server tanpa mengirim isi catatan atau menjalankan model.
func (c *Client) Probe(ctx context.Context) ProviderStatus {
	probeCtx, cancel := context.WithTimeout(ctx, defaultProbeTimeout)
	defer cancel()

	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/v1/models"
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return ProviderInvalid
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ProviderUnavailable
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, c.maxResponseBytes))
		if resp.StatusCode >= 500 {
			return ProviderUnavailable
		}
		return ProviderInvalid
	}
	if resp.ContentLength > c.maxResponseBytes {
		return ProviderInvalid
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes+1))
	if err != nil || int64(len(body)) > c.maxResponseBytes {
		return ProviderInvalid
	}

	var out struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return ProviderInvalid
	}
	return ProviderReady
}

func (c *Client) Generate(ctx context.Context, in Request) (Response, error) {
	if strings.TrimSpace(in.Model) == "" {
		return Response{}, fmt.Errorf("%w: model wajib diisi", ErrProviderRejected)
	}
	if len(in.Messages) == 0 {
		return Response{}, fmt.Errorf("%w: pesan wajib diisi", ErrProviderRejected)
	}
	for _, message := range in.Messages {
		if message.Role != RoleSystem && message.Role != RoleUser && message.Role != RoleAssistant {
			return Response{}, fmt.Errorf("%w: peran pesan tidak dikenal", ErrProviderRejected)
		}
		if strings.TrimSpace(message.Content) == "" {
			return Response{}, fmt.Errorf("%w: isi pesan kosong", ErrProviderRejected)
		}
	}

	body, err := json.Marshal(struct {
		Model    string    `json:"model"`
		Messages []Message `json:"messages"`
	}{Model: in.Model, Messages: in.Messages})
	if err != nil {
		return Response{}, fmt.Errorf("%w: %v", ErrProviderRejected, err)
	}
	if int64(len(body)) > c.maxRequestBytes {
		return Response{}, ErrRequestTooLarge
	}

	endpoint := *c.baseURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/v1/chat/completions"
	reqCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(reqCtx, http.MethodPost, endpoint.String(), bytes.NewReader(body))
	if err != nil {
		return Response{}, fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(reqCtx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
			return Response{}, ErrTimeout
		}
		var netErr net.Error
		if errors.As(err, &netErr) {
			return Response{}, fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
		}
		return Response{}, fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, c.maxResponseBytes))
		if resp.StatusCode >= 500 {
			return Response{}, fmt.Errorf("%w: HTTP %d", ErrProviderUnavailable, resp.StatusCode)
		}
		return Response{}, fmt.Errorf("%w: HTTP %d", ErrProviderRejected, resp.StatusCode)
	}
	if resp.ContentLength > c.maxResponseBytes {
		return Response{}, ErrResponseTooLarge
	}
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, c.maxResponseBytes+1))
	if err != nil {
		return Response{}, fmt.Errorf("%w: %v", ErrProviderUnavailable, err)
	}
	if int64(len(responseBody)) > c.maxResponseBytes {
		return Response{}, ErrResponseTooLarge
	}

	var out struct {
		Model string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(responseBody, &out); err != nil {
		return Response{}, fmt.Errorf("%w: JSON tidak valid", ErrInvalidResponse)
	}
	if len(out.Choices) == 0 || strings.TrimSpace(out.Choices[0].Message.Content) == "" {
		return Response{}, fmt.Errorf("%w: pilihan teks tidak ditemukan", ErrInvalidResponse)
	}
	return Response{Text: out.Choices[0].Message.Content, Model: out.Model}, nil
}
