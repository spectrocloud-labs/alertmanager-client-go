package alertmanager

import (
	"net/http"
	"testing"
	"time"

	"github.com/go-logr/logr"
)

func TestWithEndpoint(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name        string
		endpoint    string
		expectError bool
		expectedURL string
	}{
		{
			name:        "valid http endpoint",
			endpoint:    "http://alertmanager:9093",
			expectError: false,
			expectedURL: "http://alertmanager:9093/api/v2/alerts",
		},
		{
			name:        "valid https endpoint",
			endpoint:    "https://secure-alertmanager.com:9093",
			expectError: false,
			expectedURL: "https://secure-alertmanager.com:9093/api/v2/alerts",
		},
		{
			name:        "endpoint with path stripped",
			endpoint:    "http://alertmanager:9093/some/path",
			expectError: false,
			expectedURL: "http://alertmanager:9093/api/v2/alerts",
		},
		{
			name:        "empty endpoint",
			endpoint:    "",
			expectError: true,
		},
		{
			name:        "invalid endpoint",
			endpoint:    "_not_valid_",
			expectError: true,
		},
		{
			name:        "endpoint without scheme",
			endpoint:    "alertmanager:9093",
			expectError: true,
		},
		{
			name:        "endpoint without host",
			endpoint:    "http://",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAlertmanager(logger, &http.Client{}, WithEndpoint(tt.endpoint))
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for endpoint '%s', but got none", tt.endpoint)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for endpoint '%s': %v", tt.endpoint, err)
				return
			}

			if am.endpoint != tt.expectedURL {
				t.Errorf("expected endpoint to be '%s', got '%s'", tt.expectedURL, am.endpoint)
			}
		})
	}
}

func TestWithBasicAuth(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "valid credentials",
			username: "testuser",
			password: "testpass",
		},
		{
			name:     "empty credentials",
			username: "",
			password: "",
		},
		{
			name:     "username only",
			username: "user",
			password: "",
		},
		{
			name:     "password only",
			username: "",
			password: "pass",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAlertmanager(logger, &http.Client{},
				WithEndpoint("http://example.com"),
				WithBasicAuth(tt.username, tt.password))
			if err != nil {
				t.Fatalf("failed to create alertmanager: %v", err)
			}

			if am.username != tt.username || am.password != tt.password {
				t.Errorf("expected username=%s, password=%s, got username=%s, password=%s",
					tt.username, tt.password, am.username, am.password)
			}
		})
	}
}

func TestWithTimeout(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "30 seconds",
			timeout: 30 * time.Second,
		},
		{
			name:    "1 minute",
			timeout: time.Minute,
		},
		{
			name:    "zero timeout",
			timeout: 0,
		},
		{
			name:    "negative timeout",
			timeout: -5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			am, err := NewAlertmanager(logger, client,
				WithEndpoint("http://example.com"),
				WithTimeout(tt.timeout))
			if err != nil {
				t.Fatalf("failed to create alertmanager: %v", err)
			}

			if am.client.Timeout != tt.timeout {
				t.Errorf("expected timeout to be %v, got %v", tt.timeout, am.client.Timeout)
			}
		})
	}
}

func TestWithBaseLabel(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "service label",
			key:   "service",
			value: "test-service",
		},
		{
			name:  "environment label",
			key:   "environment",
			value: "production",
		},
		{
			name:  "empty key",
			key:   "",
			value: "value",
		},
		{
			name:  "empty value",
			key:   "key",
			value: "",
		},
		{
			name:  "both empty",
			key:   "",
			value: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAlertmanager(logger, &http.Client{},
				WithEndpoint("http://example.com"),
				WithBaseLabel(tt.key, tt.value))
			if err != nil {
				t.Fatalf("failed to create alertmanager: %v", err)
			}

			if am.labels[tt.key] != tt.value {
				t.Errorf("expected label '%s' to be '%s', got '%s'", tt.key, tt.value, am.labels[tt.key])
			}
		})
	}
}

func TestWithBaseAnnotation(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name  string
		key   string
		value string
	}{
		{
			name:  "team annotation",
			key:   "team",
			value: "platform",
		},
		{
			name:  "runbook annotation",
			key:   "runbook",
			value: "https://example.com/runbook",
		},
		{
			name:  "empty key",
			key:   "",
			value: "value",
		},
		{
			name:  "empty value",
			key:   "key",
			value: "",
		},
		{
			name:  "both empty",
			key:   "",
			value: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAlertmanager(logger, &http.Client{},
				WithEndpoint("http://example.com"),
				WithBaseAnnotation(tt.key, tt.value))
			if err != nil {
				t.Fatalf("failed to create alertmanager: %v", err)
			}

			if am.annotations[tt.key] != tt.value {
				t.Errorf("expected annotation '%s' to be '%s', got '%s'", tt.key, tt.value, am.annotations[tt.key])
			}
		})
	}
}

func TestWithCustomCA(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name            string
		caCert          []byte
		expectTransport bool
	}{
		{
			name:            "valid CA certificate",
			caCert:          []byte("-----BEGIN CERTIFICATE-----\nfake cert data\n-----END CERTIFICATE-----"),
			expectTransport: true,
		},
		{
			name:            "empty CA certificate",
			caCert:          []byte{},
			expectTransport: true,
		},
		{
			name:            "nil CA certificate",
			caCert:          nil,
			expectTransport: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			am, err := NewAlertmanager(logger, client,
				WithEndpoint("https://example.com"),
				WithCustomCA(tt.caCert))
			if err != nil {
				t.Fatalf("failed to create alertmanager: %v", err)
			}

			hasTransport := am.client.Transport != nil
			if hasTransport != tt.expectTransport {
				t.Errorf("expected transport set=%v, got transport set=%v", tt.expectTransport, hasTransport)
			}
		})
	}
}

func TestWithInsecure(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name     string
		insecure bool
	}{
		{
			name:     "insecure true",
			insecure: true,
		},
		{
			name:     "insecure false",
			insecure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			am, err := NewAlertmanager(logger, client,
				WithEndpoint("https://example.com"),
				WithInsecure(tt.insecure))
			if err != nil {
				t.Fatalf("failed to create alertmanager: %v", err)
			}

			// Verify transport was set
			if am.client.Transport == nil {
				t.Errorf("expected transport to be set")
			}
		})
	}
}

func TestWithProxyURL(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name        string
		proxyURL    string
		expectError bool
	}{
		{
			name:     "valid proxy URL",
			proxyURL: "http://proxy.example.com:8080",
		},
		{
			name:     "valid https proxy URL",
			proxyURL: "https://secure-proxy.example.com:3128",
		},
		{
			name:     "empty proxy URL",
			proxyURL: "",
		},
		{
			name:        "invalid proxy URL",
			proxyURL:    "://invalid-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &http.Client{}
			am, err := NewAlertmanager(logger, client,
				WithEndpoint("https://example.com"),
				WithProxyURL(tt.proxyURL))

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error for proxy URL '%s', but got none", tt.proxyURL)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error for proxy URL '%s': %v", tt.proxyURL, err)
				return
			}

			// Verify transport was configured
			transport, ok := am.client.Transport.(*http.Transport)
			if !ok {
				if tt.proxyURL == "" {
					return
				}

				t.Errorf("expected http.Transport, got %T", am.client.Transport)
				return
			}

			if transport.Proxy == nil {
				t.Errorf("expected proxy to be set")
			}
		})
	}
}
