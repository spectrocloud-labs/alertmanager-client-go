package alertmanager

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
)

// Option represents a configuration option for Alertmanager.
type Option func(*Alertmanager) error

// WithEndpoint sets the Alertmanager endpoint URL.
func WithEndpoint(endpoint string) Option {
	return func(a *Alertmanager) error {
		if endpoint == "" {
			return ErrEndpointRequired
		}

		u, err := url.Parse(endpoint)
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
		return nil
	}
}

// WithBasicAuth sets basic authentication credentials.
func WithBasicAuth(username, password string) Option {
	return func(a *Alertmanager) error {
		a.username = username
		a.password = password
		return nil
	}
}

// WithCustomCA configures TLS with a custom CA certificate.
func WithCustomCA(caCert []byte) Option {
	return func(a *Alertmanager) error {
		caCertPool, err := x509.SystemCertPool()
		if err != nil {
			a.log.Error(err, "failed to get system cert pool; using empty pool")
			caCertPool = x509.NewCertPool()
		}
		if len(caCert) > 0 {
			caCertPool.AppendCertsFromPEM(caCert)
		}

		transport, ok := a.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}

		transport.TLSClientConfig.RootCAs = caCertPool
		a.client.Transport = transport

		return nil
	}
}

// WithInsecure configures TLS to skip certificate verification.
func WithInsecure(insecureSkipVerify bool) Option {
	return func(a *Alertmanager) error {
		transport, ok := a.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}

		transport.TLSClientConfig.InsecureSkipVerify = insecureSkipVerify // #nosec G402
		a.client.Transport = transport

		return nil
	}
}

// WithTimeout sets the HTTP client timeout on the existing client.
func WithTimeout(timeout time.Duration) Option {
	return func(a *Alertmanager) error {
		a.client.Timeout = timeout
		return nil
	}
}

// WithLabel adds a base label that will be applied to all alerts.
func WithLabel(key, value string) Option {
	return func(a *Alertmanager) error {
		if a.labels == nil {
			a.labels = make(map[string]string)
		}
		a.labels[key] = value
		return nil
	}
}

// WithAnnotation adds a base annotation that will be applied to all alerts.
func WithAnnotation(key, value string) Option {
	return func(a *Alertmanager) error {
		if a.annotations == nil {
			a.annotations = make(map[string]string)
		}
		a.annotations[key] = value
		return nil
	}
}
