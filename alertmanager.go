package alertmanager

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"

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

// Alertmanager represents the Alertmanager client.
type Alertmanager struct {
	client *http.Client
	log    logr.Logger

	endpoint string
	username string
	password string

	// base labels and annotations to be applied to all alerts created by this Alertmanager instance
	labels      map[string]string
	annotations map[string]string
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

	if a.username != "" && a.password != "" {
		req.Header.Add(basicAuthHeader(a.username, a.password))
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to post alert to %s: %w", a.endpoint, err)
	}

	return resp, nil
}

func basicAuthHeader(username, password string) (string, string) {
	auth := base64.StdEncoding.EncodeToString(
		bytes.Join([][]byte{[]byte(username), []byte(password)}, []byte(":")),
	)
	return "Authorization", fmt.Sprintf("Basic %s", auth)
}
