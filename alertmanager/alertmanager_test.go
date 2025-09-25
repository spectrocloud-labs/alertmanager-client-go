package alertmanager

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
)

func TestNewAlertmanager(t *testing.T) {
	logger := logr.Discard()
	client := &http.Client{}

	tests := []struct {
		name        string
		logger      logr.Logger
		client      *http.Client
		options     []Option
		expectedErr error
	}{
		{
			name:    "no options",
			logger:  logger,
			client:  client,
			options: []Option{},
		},
		{
			name:   "nil client",
			logger: logger,
			client: nil,
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
			},
			expectedErr: ErrNilHTTPClient,
		},
		{
			name:   "empty endpoint",
			logger: logger,
			client: client,
			options: []Option{
				WithEndpoint(""),
			},
			expectedErr: ErrEndpointRequired,
		},
		{
			name:   "invalid endpoint",
			logger: logger,
			client: client,
			options: []Option{
				WithEndpoint("_not_an_endpoint_"),
			},
			expectedErr: ErrInvalidEndpoint,
		},
		{
			name:   "valid endpoint",
			logger: logger,
			client: client,
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
			},
		},
		{
			name:   "with basic auth and endpoint",
			logger: logger,
			client: client,
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
				WithBasicAuth("user", "pass"),
			},
		},
		{
			name:   "with timeout and endpoint",
			logger: logger,
			client: client,
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
				WithTimeout(45 * time.Second),
			},
		},
		{
			name:   "with custom CA and endpoint",
			logger: logger,
			client: client,
			options: []Option{
				WithEndpoint("https://alertmanager:9093"),
				WithCustomCA([]byte("fake-cert")),
			},
		},
		{
			name:   "with insecure TLS and endpoint",
			logger: logger,
			client: client,
			options: []Option{
				WithEndpoint("https://alertmanager:9093"),
				WithInsecure(true),
			},
		},
		{
			name:   "with base labels and annotations",
			logger: logger,
			client: client,
			options: []Option{
				WithEndpoint("http://alertmanager:9093"),
				WithLabel("service", "test"),
				WithAnnotation("team", "platform"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			am, err := NewAlertmanager(tt.logger, tt.client, tt.options...)

			if tt.expectedErr != nil {
				if err == nil {
					t.Errorf("expected error %v but got none", tt.expectedErr)
					return
				}
				if err.Error() != tt.expectedErr.Error() {
					t.Errorf("expected error %v, got %v", tt.expectedErr, err)
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

func TestEmit(t *testing.T) {
	logger := logr.Discard()

	tests := []struct {
		name          string
		serverStatus  int // 0 = no server, 200 = OK, 500 = error
		options       []Option
		alerts        []*Alert
		expectedError error
	}{
		{
			name:         "successful emit",
			serverStatus: http.StatusOK,
			alerts: []*Alert{
				NewAlert().WithLabel("alertname", "test").WithAnnotation("message", "test message"),
			},
		},
		{
			name:         "emit with base fields",
			serverStatus: http.StatusOK,
			options: []Option{
				WithLabel("service", "test-service"),
				WithLabel("environment", "test"),
				WithAnnotation("team", "platform"),
				WithTimeout(10 * time.Second),
			},
			alerts: []*Alert{
				NewAlert().WithLabel("alertname", "test-alert").WithAnnotation("message", "test message"),
			},
		},
		{
			name:         "emit empty alerts",
			serverStatus: http.StatusOK,
			alerts:       []*Alert{},
		},
		{
			name:         "emit with nil alerts",
			serverStatus: http.StatusOK,
			alerts:       []*Alert{nil, NewAlert().WithLabel("alertname", "test"), nil},
		},
		{
			name: "emit without endpoint",
			alerts: []*Alert{
				NewAlert().WithLabel("alertname", "test").WithAnnotation("message", "test message"),
			},
			expectedError: ErrEndpointRequired,
		},
		{
			name:         "server returns error",
			serverStatus: http.StatusInternalServerError,
			alerts: []*Alert{
				NewAlert().WithLabel("alertname", "test"),
			},
			expectedError: ErrEmissionFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.serverStatus > 0 {
				server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(tt.serverStatus)
				}))
				defer server.Close()
			}

			var options []Option
			if server != nil {
				options = append([]Option{WithEndpoint(server.URL)}, tt.options...)
			} else {
				options = tt.options
			}

			am, err := NewAlertmanager(logger, &http.Client{}, options...)
			if err != nil {
				t.Fatalf("failed to create alertmanager: %v", err)
			}

			err = am.Emit(tt.alerts...)

			if tt.expectedError != nil {
				if err == nil {
					t.Errorf("expected error %v but got none", tt.expectedError)
					return
				}
				if err != tt.expectedError {
					t.Errorf("expected error %v, got %v", tt.expectedError, err)
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBasicAuthHeader(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		expected string
	}{
		{
			name:     "valid credentials",
			username: "bob",
			password: "frogs",
			expected: "Basic Ym9iOmZyb2dz",
		},
		{
			name:     "empty credentials",
			username: "",
			password: "",
			expected: "Basic Og==",
		},
		{
			name:     "username only",
			username: "user",
			password: "",
			expected: "Basic dXNlcjo=",
		},
		{
			name:     "password only",
			username: "",
			password: "pass",
			expected: "Basic OnBhc3M=",
		},
		{
			name:     "special characters",
			username: "user@domain.com",
			password: "p@ssw0rd!",
			expected: "Basic dXNlckBkb21haW4uY29tOnBAc3N3MHJkIQ==",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, got := basicAuthHeader(tt.username, tt.password)
			if got != tt.expected {
				t.Errorf("expected (%s), got (%s)", tt.expected, got)
			}
		})
	}
}
