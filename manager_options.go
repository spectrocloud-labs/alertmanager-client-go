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

// TLSVersion represents a TLS protocol version.
type TLSVersion uint16

// TLS version constants for use with WithMinTLSVersion.
// These map directly to crypto/tls version constants.
// TLS 1.2 is the recommended minimum for production use.
const (
	TLS10 TLSVersion = tls.VersionTLS10
	TLS11 TLSVersion = tls.VersionTLS11
	TLS12 TLSVersion = tls.VersionTLS12
	TLS13 TLSVersion = tls.VersionTLS13
)

// ManagerOption represents a configuration option for Alertmanager.
type ManagerOption func(*Alertmanager) error

// WithEndpoint sets the Alertmanager endpoint URL.
func WithEndpoint(endpoint string) ManagerOption {
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
func WithBasicAuth(username, password string) ManagerOption {
	return func(a *Alertmanager) error {
		a.username = username
		a.password = password
		return nil
	}
}

// WithCustomCA configures TLS with a custom CA certificate.
func WithCustomCA(caCert []byte) ManagerOption {
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
			transport = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			}
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
func WithInsecure(insecureSkipVerify bool) ManagerOption {
	return func(a *Alertmanager) error {
		transport, ok := a.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			}
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}

		transport.TLSClientConfig.InsecureSkipVerify = insecureSkipVerify
		a.client.Transport = transport

		return nil
	}
}

// WithMinTLSVersion sets the minimum TLS version for HTTPS connections.
// Use the TLS* constants (e.g., TLS12, TLS13).
// If not specified, TLS 1.2 is used as the default minimum.
func WithMinTLSVersion(minVersion TLSVersion) ManagerOption {
	return func(a *Alertmanager) error {
		transport, ok := a.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			}
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{}
		}

		transport.TLSClientConfig.MinVersion = uint16(minVersion)
		a.client.Transport = transport

		return nil
	}
}

// WithMaxTLSVersion sets the maximum TLS version for HTTPS connections.
// Use the TLS* constants (e.g., TLS12, TLS13).
func WithMaxTLSVersion(maxVersion TLSVersion) ManagerOption {
	return func(a *Alertmanager) error {
		transport, ok := a.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{
				Proxy: http.ProxyFromEnvironment,
			}
		}

		if transport.TLSClientConfig == nil {
			transport.TLSClientConfig = &tls.Config{
				MinVersion: tls.VersionTLS12,
			}
		}

		transport.TLSClientConfig.MaxVersion = uint16(maxVersion)
		a.client.Transport = transport

		return nil
	}
}

// WithProxyURL configures an HTTP proxy for the client.
func WithProxyURL(proxyURL string) ManagerOption {
	return func(a *Alertmanager) error {
		if proxyURL == "" {
			return nil
		}

		parsedURL, err := url.Parse(proxyURL)
		if err != nil {
			return errors.Wrap(err, "invalid proxy URL")
		}

		transport, ok := a.client.Transport.(*http.Transport)
		if !ok {
			transport = &http.Transport{}
		}

		transport.Proxy = http.ProxyURL(parsedURL)
		a.client.Transport = transport

		return nil
	}
}

// WithTimeout sets the HTTP client timeout on the existing client.
func WithTimeout(timeout time.Duration) ManagerOption {
	return func(a *Alertmanager) error {
		a.client.Timeout = timeout
		return nil
	}
}

// WithBaseLabel adds a base label that will be applied to all alerts.
func WithBaseLabel(key, value string) ManagerOption {
	return func(a *Alertmanager) error {
		if a.labels == nil {
			a.labels = make(map[string]string)
		}
		a.labels[key] = value
		return nil
	}
}

// WithBaseAnnotation adds a base annotation that will be applied to all alerts.
func WithBaseAnnotation(key, value string) ManagerOption {
	return func(a *Alertmanager) error {
		if a.annotations == nil {
			a.annotations = make(map[string]string)
		}
		a.annotations[key] = value
		return nil
	}
}
