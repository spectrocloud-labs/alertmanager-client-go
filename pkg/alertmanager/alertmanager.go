package alertmanager

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"net/url"
	"strconv"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

// Alertmanager represents the Alertmanager client.
type Alertmanager struct {
	client Client
	log    logr.Logger

	endpoint string
	username string
	password string

	// base labels and annotations to be applied to all alerts
	labels      map[string]string
	annotations map[string]string
}

var (
	// ErrInvalidEndpoint is returned when an Alertmanager endpoint is invalid.
	ErrInvalidEndpoint = errors.New("invalid Alertmanager config: endpoint scheme and host are required")

	// ErrEndpointRequired is returned when the Alertmanager endpoint is not provided.
	ErrEndpointRequired = errors.New("invalid Alertmanager config: endpoint required")

	// ErrEmissionFailed is returned when alert emission fails.
	ErrEmissionFailed = errors.New("emission failed")
)

// Configure configures the Alertmanager with the provided configuration.
func (a *Alertmanager) Configure(c Client, config map[string][]byte) error {
	// endpoint
	endpoint, ok := config["endpoint"]
	if !ok {
		return ErrEndpointRequired
	}
	u, err := url.Parse(string(endpoint))
	if err != nil {
		return errors.Wrap(err, "invalid Alertmanager config: failed to parse endpoint")
	}
	if u.Scheme == "" || u.Host == "" {
		return ErrInvalidEndpoint
	}
	if u.Path != "" {
		a.log.V(1).Info("stripping path from Alertmanager endpoint", "path", u.Path)
		u.Path = ""
	}
	a.endpoint = fmt.Sprintf("%s/api/v2/alerts", u.String())

	// basic auth
	a.username = string(config["username"])
	a.password = string(config["password"])

	// tls
	var caCertPool *x509.CertPool
	var insecureSkipVerify bool

	insecure, ok := config["insecureSkipVerify"]
	if ok {
		insecureSkipVerify, err = strconv.ParseBool(string(insecure))
		if err != nil {
			return errors.Wrap(err, "invalid Alertmanager config: failed to parse insecureSkipVerify")
		}
	}
	caCert, ok := config["caCert"]
	if ok {
		caCertPool, err = x509.SystemCertPool()
		if err != nil {
			a.log.Error(err, "failed to get system cert pool; using empty pool")
			caCertPool = x509.NewCertPool()
		}
		caCertPool.AppendCertsFromPEM(caCert)
	}
	c.hclient.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify, // #nosec G402
			MinVersion:         tls.VersionTLS12,
			RootCAs:            caCertPool,
		},
	}
	a.client = c

	// Initialize base labels and annotations maps if not already initialized
	if a.labels == nil {
		a.labels = make(map[string]string)
	}
	if a.annotations == nil {
		a.annotations = make(map[string]string)
	}

	return nil
}

// WithLabel adds a base label that will be applied to all alerts sent by this Alertmanager.
func (a *Alertmanager) WithLabel(key, value string) *Alertmanager {
	if a.labels == nil {
		a.labels = make(map[string]string)
	}
	a.labels[key] = value
	return a
}

// WithAnnotation adds a base annotation that will be applied to all alerts sent by this Alertmanager.
func (a *Alertmanager) WithAnnotation(key, value string) *Alertmanager {
	if a.annotations == nil {
		a.annotations = make(map[string]string)
	}
	a.annotations[key] = value
	return a
}

// Emit sends one or more alerts to Alertmanager.
func (a *Alertmanager) Emit(alerts ...*Alert) error {
	if len(alerts) == 0 {
		return nil
	}

	// Merge base labels and annotations with each alert
	finalAlerts := make([]Alert, 0, len(alerts))
	for _, alert := range alerts {
		if alert == nil {
			continue
		}

		// Create a new alert with merged labels and annotations
		mergedAlert := Alert{
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		}

		// merge labels
		maps.Copy(mergedAlert.Labels, a.labels)
		maps.Copy(mergedAlert.Labels, alert.Labels)

		// merge annotations
		maps.Copy(mergedAlert.Annotations, a.annotations)
		maps.Copy(mergedAlert.Annotations, alert.Annotations)

		finalAlerts = append(finalAlerts, mergedAlert)
	}

	body, err := json.Marshal(finalAlerts)
	if err != nil {
		a.log.Error(err, "failed to marshal alerts", "alerts", finalAlerts)
		return err
	}
	a.log.V(1).Info("Alertmanager message", "payload", string(body))

	req, err := http.NewRequest(http.MethodPost, a.endpoint, bytes.NewReader(body))
	if err != nil {
		a.log.Error(err, "failed to create HTTP POST request", "endpoint", a.endpoint)
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	if a.username != "" && a.password != "" {
		req.Header.Add(basicAuthHeader(a.username, a.password))
	}

	resp, err := a.client.hclient.Do(req)
	defer func() {
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()
	if err != nil {
		a.log.Error(err, "failed to post alert", "endpoint", a.endpoint)
		return err
	}
	if resp.StatusCode != 200 {
		a.log.V(0).Info("failed to post alert", "endpoint", a.endpoint, "status", resp.Status, "code", resp.StatusCode)
		return ErrEmissionFailed
	}

	a.log.V(0).Info("Successfully posted alert to Alertmanager", "endpoint", a.endpoint, "status", resp.Status, "code", resp.StatusCode)
	return nil
}

func basicAuthHeader(username, password string) (string, string) {
	auth := base64.StdEncoding.EncodeToString(
		bytes.Join([][]byte{[]byte(username), []byte(password)}, []byte(":")),
	)
	return "Authorization", fmt.Sprintf("Basic %s", auth)
}
