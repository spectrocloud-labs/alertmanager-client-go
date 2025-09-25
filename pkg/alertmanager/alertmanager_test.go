package alertmanager

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"
)

var sinkClient = NewClient(1 * time.Second)

func TestAlertmanagerConfigure(t *testing.T) {
	cs := []struct {
		name     string
		am       Alertmanager
		config   map[string][]byte
		expected error
	}{
		{
			name: "Pass",
			am:   Alertmanager{},
			config: map[string][]byte{
				"endpoint": []byte("http://fake.alertmanager.com:9093/api/v2/alerts"),
				"caCert":   []byte("_fake_ca_cert"),
			},
			expected: nil,
		},
		{
			name:     "Fail (no endpoint)",
			am:       Alertmanager{},
			config:   map[string][]byte{},
			expected: ErrEndpointRequired,
		},
		{
			name: "Fail (invalid endpoint)",
			am:   Alertmanager{},
			config: map[string][]byte{
				"endpoint": []byte("_not_an_endpoint_"),
			},
			expected: ErrInvalidEndpoint,
		},
		{
			name: "Fail (invalid insecureSkipVerify)",
			am:   Alertmanager{},
			config: map[string][]byte{
				"endpoint":           []byte("https://fake.com"),
				"insecureSkipVerify": []byte("_not_a_bool_"),
			},
			expected: errors.New(`invalid Alertmanager config: failed to parse insecureSkipVerify: strconv.ParseBool: parsing "_not_a_bool_": invalid syntax`),
		},
	}
	for _, c := range cs {
		t.Log(c.name)
		err := c.am.Configure(*sinkClient, c.config)
		if err != nil && !reflect.DeepEqual(err.Error(), c.expected.Error()) {
			t.Errorf("expected (%v), got (%v)", c.expected, err)
		}
	}
}

func TestAlertManagerEmit(t *testing.T) {
	cs := []struct {
		name     string
		am       Alertmanager
		alerts   []*Alert
		server   *httptest.Server
		expected error
	}{
		{
			name: "Pass",
			am:   Alertmanager{},
			alerts: []*Alert{
				NewAlert().
					WithLabel("alertname", "test-alert").
					WithLabel("severity", "critical").
					WithAnnotation("message", "Test message"),
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, _ = fmt.Fprintf(w, "ok")
			})),
			expected: nil,
		},
		{
			name: "Fail",
			am:   Alertmanager{},
			alerts: []*Alert{
				NewAlert().WithLabel("alertname", "test-fail"),
			},
			server: httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, "invalid auth", http.StatusUnauthorized)
			})),
			expected: ErrEmissionFailed,
		},
	}
	for _, c := range cs {
		t.Log(c.name)
		defer c.server.Close()
		_ = c.am.Configure(*sinkClient, map[string][]byte{
			"endpoint": []byte(c.server.URL),
		})
		err := c.am.Emit(c.alerts...)
		if err != nil && !reflect.DeepEqual(err.Error(), c.expected.Error()) {
			t.Errorf("expected (%v), got (%v)", c.expected, err)
		}
	}
}

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
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprintf(w, "ok")
	}))
	defer server.Close()

	am := &Alertmanager{}
	_ = am.Configure(*sinkClient, map[string][]byte{
		"endpoint": []byte(server.URL),
	})

	am.WithLabel("service", "test-service").
		WithLabel("environment", "test").
		WithAnnotation("team", "platform")

	alert := NewAlert().
		WithLabel("alertname", "test-alert").
		WithAnnotation("message", "Test message")

	err := am.Emit(alert)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

