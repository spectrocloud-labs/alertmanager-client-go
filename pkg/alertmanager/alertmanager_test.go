package alertmanager

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-logr/logr"
)

func TestBasicAuthHeader(t *testing.T) {
	cs := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{
			name:     "Pass",
			username: "bob",
			password: "frogs",
			expected: "Basic Ym9iOmZyb2dz",
		},
	}
	for _, c := range cs {
		t.Log(c.name)
		_, v := basicAuthHeader(c.username, c.password)
		if !reflect.DeepEqual(c.expected, v) {
			t.Errorf("expected (%s), got (%s)", c.expected, v)
		}
	}
}

func TestAlertBuilderMethods(t *testing.T) {
	alert := NewAlert().
		WithLabel("alertname", "test").
		WithLabel("severity", "critical").
		WithAnnotation("message", "Test message").
		WithAnnotation("runbook", "https://example.com")

	expectedLabels := map[string]string{
		"alertname": "test",
		"severity":  "critical",
	}
	expectedAnnotations := map[string]string{
		"message": "Test message",
		"runbook": "https://example.com",
	}

	if !reflect.DeepEqual(alert.Labels, expectedLabels) {
		t.Errorf("expected labels (%v), got (%v)", expectedLabels, alert.Labels)
	}
	if !reflect.DeepEqual(alert.Annotations, expectedAnnotations) {
		t.Errorf("expected annotations (%v), got (%v)", expectedAnnotations, alert.Annotations)
	}
}

func TestAlertmanagerWithBaseFields(t *testing.T) {
	logger := logr.Discard()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "ok")
	}))
	defer server.Close()

	am, err := NewAlertmanager(logger, &http.Client{},
		WithEndpoint(server.URL),
		WithLabel("service", "test-service"),
		WithLabel("environment", "test"),
		WithAnnotation("team", "platform"))
	if err != nil {
		t.Fatalf("failed to create alertmanager: %v", err)
	}

	alert := NewAlert().
		WithLabel("alertname", "test-alert").
		WithAnnotation("message", "Test message")

	err = am.Emit(alert)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

// Tests for the new functional options pattern

func TestNewAlertmanager(t *testing.T) {
	logger := logr.Discard()
	client := &http.Client{}

	tests := []struct {
		name        string
		options     []Option
		expectError bool
	}{
		{
			name:        "No options",
			options:     []Option{},
			expectError: false,
		},
		{
			name: "Valid endpoint",
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
			},
			expectError: false,
		},
		{
			name: "Empty endpoint",
			options: []Option{
				WithEndpoint(""),
			},
			expectError: true,
		},
		{
			name: "Invalid endpoint",
			options: []Option{
				WithEndpoint("_not_an_endpoint_"),
			},
			expectError: true,
		},
		{
			name: "With basic auth and endpoint",
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
				WithBasicAuth("user", "pass"),
			},
			expectError: false,
		},
		{
			name: "With timeout and endpoint",
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
				WithTimeout(45 * time.Second),
			},
			expectError: false,
		},
		{
			name: "With custom CA and endpoint",
			options: []Option{
				WithEndpoint("https://alertmanager:9093"),
				WithCustomCA([]byte("fake-cert")),
			},
			expectError: false,
		},
		{
			name: "With insecure TLS and endpoint",
			options: []Option{
				WithEndpoint("https://alertmanager:9093"),
				WithInsecure(true),
			},
			expectError: false,
		},
		{
			name: "With base labels and annotations",
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
				WithLabel("service", "test"),
				WithAnnotation("team", "platform"),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAlertmanager(logger, client, tt.options...)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if am == nil {
				t.Errorf("expected non-nil Alertmanager")
			}
		})
	}
}

func TestNewAlertmanagerEmit(t *testing.T) {
	logger := logr.Discard()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test basic functionality
	am, err := NewAlertmanager(logger, &http.Client{}, WithEndpoint(server.URL))
	if err != nil {
		t.Fatalf("failed to create alertmanager: %v", err)
	}

	alert := NewAlert().
		WithLabel("alertname", "test").
		WithAnnotation("message", "test message")

	err = am.Emit(alert)
	if err != nil {
		t.Errorf("unexpected error during emit: %v", err)
	}
}

func TestNewAlertmanagerWithBaseFields(t *testing.T) {
	logger := logr.Discard()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// Test with base labels and annotations
	am, err := NewAlertmanager(logger, &http.Client{},
		WithEndpoint(server.URL),
		WithLabel("service", "test-service"),
		WithLabel("environment", "test"),
		WithAnnotation("team", "platform"),
		WithTimeout(10*time.Second),
	)
	if err != nil {
		t.Fatalf("failed to create alertmanager: %v", err)
	}

	alert := NewAlert().
		WithLabel("alertname", "test-alert").
		WithAnnotation("message", "test message")

	err = am.Emit(alert)
	if err != nil {
		t.Errorf("unexpected error during emit: %v", err)
	}

	// Verify base fields were applied (this would need inspection of the actual HTTP request in a real test)
	if am.labels["service"] != "test-service" {
		t.Errorf("expected base label 'service' to be 'test-service', got '%s'", am.labels["service"])
	}
	if am.annotations["team"] != "platform" {
		t.Errorf("expected base annotation 'team' to be 'platform', got '%s'", am.annotations["team"])
	}
}

func TestWithOptionsIndependently(t *testing.T) {
	logger := logr.Discard()

	// Test that options can be applied independently
	am, err := NewAlertmanager(logger, &http.Client{}, WithEndpoint("http://example.com"))
	if err != nil {
		t.Fatalf("failed to create alertmanager: %v", err)
	}

	// Apply timeout option
	timeoutOption := WithTimeout(60 * time.Second)
	err = timeoutOption(am)
	if err != nil {
		t.Errorf("failed to apply timeout option: %v", err)
	}
	if am.client.Timeout != 60*time.Second {
		t.Errorf("expected timeout to be 60s, got %v", am.client.Timeout)
	}

	// Apply basic auth option
	authOption := WithBasicAuth("user", "pass")
	err = authOption(am)
	if err != nil {
		t.Errorf("failed to apply basic auth option: %v", err)
	}
	if am.username != "user" || am.password != "pass" {
		t.Errorf("expected username=user, password=pass, got username=%s, password=%s", am.username, am.password)
	}
}

func TestWithEndpoint(t *testing.T) {
	logger := logr.Discard()

	// Test WithEndpoint option
	am, err := NewAlertmanager(logger, &http.Client{}, WithEndpoint("http://initial-endpoint.com"))
	if err != nil {
		t.Fatalf("failed to create alertmanager: %v", err)
	}

	// Change endpoint using WithEndpoint
	newEndpoint := "https://new-alertmanager.com:9093"
	endpointOption := WithEndpoint(newEndpoint)
	err = endpointOption(am)
	if err != nil {
		t.Errorf("failed to apply endpoint option: %v", err)
	}

	expectedEndpoint := "https://new-alertmanager.com:9093/api/v2/alerts"
	if am.endpoint != expectedEndpoint {
		t.Errorf("expected endpoint to be '%s', got '%s'", expectedEndpoint, am.endpoint)
	}

	// Test invalid endpoint
	invalidEndpointOption := WithEndpoint("_not_valid_")
	err = invalidEndpointOption(am)
	if err == nil {
		t.Errorf("expected error for invalid endpoint, but got none")
	}

	// Test empty endpoint
	emptyEndpointOption := WithEndpoint("")
	err = emptyEndpointOption(am)
	if err == nil {
		t.Errorf("expected error for empty endpoint, but got none")
	}
}

func TestEmitWithoutEndpoint(t *testing.T) {
	logger := logr.Discard()

	// Create alertmanager without endpoint
	am, err := NewAlertmanager(logger, &http.Client{})
	if err != nil {
		t.Fatalf("failed to create alertmanager: %v", err)
	}

	alert := NewAlert().
		WithLabel("alertname", "test").
		WithAnnotation("message", "test message")

	// Should fail because no endpoint is set
	err = am.Emit(alert)
	if err == nil {
		t.Errorf("expected error when emitting without endpoint, but got none")
	}
	if err != ErrEndpointRequired {
		t.Errorf("expected ErrEndpointRequired, got %v", err)
	}
}
