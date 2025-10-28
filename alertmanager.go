package alertmanager

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"
)

var (
	// ErrInvalidEndpoint is returned when an Alertmanager endpoint is invalid.
	ErrInvalidEndpoint = errors.New("invalid Alertmanager config: endpoint scheme and host are required")

	// ErrEndpointRequired is returned when the Alertmanager endpoint is not provided.
	ErrEndpointRequired = errors.New("invalid Alertmanager config: endpoint required")

	// ErrNilHTTPClient is returned when a nil HTTP client is provided.
	ErrNilHTTPClient = errors.New("HTTP client cannot be nil")
)

// Args contains client configuration for Alertmanager.
type Args struct {
	// Enabled determines whether the Alertmanager client should be created
	Enabled bool

	// AlertmanagerURL is the URL of the Alertmanager instance
	AlertmanagerURL string

	// Username is the username for basic authentication (optional)
	Username string

	// Password is the password for basic authentication (optional)
	Password string

	// TLSCACertPath is the path to the TLS CA certificate (optional)
	TLSCACertPath string

	// TLSInsecureSkipVerify skips TLS certificate verification (optional)
	TLSInsecureSkipVerify bool

	// TLSMinVersion is the minimum TLS version (optional, e.g., "TLS12", "TLS13")
	TLSMinVersion string

	// TLSMaxVersion is the maximum TLS version (optional, e.g., "TLS12", "TLS13")
	TLSMaxVersion string

	// ProxyURL is the HTTP proxy URL (optional)
	ProxyURL string

	// Timeout is the timeout for HTTP requests to Alertmanager
	// If not specified, a default of 2 seconds is used
	Timeout time.Duration
}

// Alertmanager represents the Alertmanager client.
type Alertmanager struct {
	client *http.Client
	log    logr.Logger

	endpoint   string
	authHeader string

	// base labels and annotations to be applied to all alerts created by this Alertmanager instance
	labels      map[string]string
	annotations map[string]string
}

// NewAlertmanagerWithArgs creates a new Alertmanager client configured with the provided args.
// This is a convenience constructor that provides defaults:
//   - Default timeout: 2 seconds (if not specified in Args.Timeout)
//   - TLS validation: Only TLS 1.2 and TLS 1.3 are accepted when TLSMinVersion/TLSMaxVersion are specified
//
// For more control over client configuration, use NewAlertmanager directly with ManagerOptions.
// Returns nil if Enabled is false. Returns an error if configuration is invalid.
func NewAlertmanagerWithArgs(logger logr.Logger, args Args) (*Alertmanager, error) {
	if !args.Enabled {
		return nil, nil
	}

	if args.AlertmanagerURL == "" {
		return nil, fmt.Errorf("alertmanager URL must be provided when enabled")
	}

	timeout := args.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	httpClient := &http.Client{}

	opts := []ManagerOption{
		WithEndpoint(args.AlertmanagerURL),
		WithTimeout(timeout),
	}

	if args.Username != "" && args.Password != "" {
		opts = append(opts, WithBasicAuth(args.Username, args.Password))
	} else if args.Username != "" || args.Password != "" {
		return nil, fmt.Errorf("both basic auth username and password must be provided together")
	}

	if args.TLSCACertPath != "" {
		caCert, err := os.ReadFile(args.TLSCACertPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read CA cert: %w", err)
		}
		opts = append(opts, WithCustomCA(caCert))
	}

	if args.TLSInsecureSkipVerify {
		opts = append(opts, WithInsecure(true))
	}

	if args.ProxyURL != "" {
		opts = append(opts, WithProxyURL(args.ProxyURL))
	}

	if args.TLSMinVersion != "" {
		minVersion, err := stringToSecureTLSVersion(args.TLSMinVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid TLS min version: %w", err)
		}
		opts = append(opts, WithMinTLSVersion(minVersion))
	}

	if args.TLSMaxVersion != "" {
		maxVersion, err := stringToSecureTLSVersion(args.TLSMaxVersion)
		if err != nil {
			return nil, fmt.Errorf("invalid TLS max version: %w", err)
		}
		opts = append(opts, WithMaxTLSVersion(maxVersion))
	}

	client, err := NewAlertmanager(logger, httpClient, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create alertmanager client: %w", err)
	}

	return client, nil
}

// NewAlertmanager creates a new Alertmanager instance with the provided logger, HTTP client, and options.
// The logger and client are required. Use WithEndpoint() to set the endpoint.
func NewAlertmanager(logger logr.Logger, client *http.Client, options ...ManagerOption) (*Alertmanager, error) {
	if client == nil {
		return nil, ErrNilHTTPClient
	}

	am := &Alertmanager{
		client:      client,
		log:         logger,
		labels:      make(map[string]string),
		annotations: make(map[string]string),
	}

	// Apply all options
	for _, opt := range options {
		if err := opt(am); err != nil {
			return nil, err
		}
	}

	return am, nil
}

// Emit sends one or more alerts to Alertmanager.
func (a *Alertmanager) Emit(alerts ...*Alert) (*http.Response, error) {
	if a.endpoint == "" {
		return nil, ErrEndpointRequired
	}

	finalAlerts := make([]Alert, 0, len(alerts))
	for _, alert := range alerts {
		if alert == nil {
			continue
		}

		mergedAlert := Alert{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
			StartsAt:    alert.StartsAt,
			EndsAt:      alert.EndsAt,
		}

		// merge labels and annotations
		maps.Copy(mergedAlert.Labels, a.labels)
		maps.Copy(mergedAlert.Labels, alert.Labels)
		maps.Copy(mergedAlert.Annotations, a.annotations)
		maps.Copy(mergedAlert.Annotations, alert.Annotations)

		finalAlerts = append(finalAlerts, mergedAlert)
	}

	body, err := json.Marshal(finalAlerts)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal alerts: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, a.endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request to %s: %w", a.endpoint, err)
	}
	req.Header.Add("Content-Type", "application/json")

	if a.authHeader != "" {
		req.Header.Add("Authorization", a.authHeader)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to post alert to %s: %w", a.endpoint, err)
	}

	return resp, nil
}

func basicAuthHeader(username, password string) string {
	auth := base64.StdEncoding.EncodeToString(
		bytes.Join([][]byte{[]byte(username), []byte(password)}, []byte(":")),
	)
	return fmt.Sprintf("Basic %s", auth)
}

// stringToSecureTLSVersion converts a string TLS version to the TLSVersion type.
// Only TLS 1.2 and TLS 1.3 are allowed for security reasons.
func stringToSecureTLSVersion(version string) (TLSVersion, error) {
	switch version {
	case "TLS10":
		return 0, fmt.Errorf("TLS 1.0 is not allowed (minimum: TLS 1.2)")
	case "TLS11":
		return 0, fmt.Errorf("TLS 1.1 is not allowed (minimum: TLS 1.2)")
	case "TLS12":
		return TLS12, nil
	case "TLS13":
		return TLS13, nil
	default:
		return 0, fmt.Errorf("must be one of: TLS12, TLS13")
	}
}
