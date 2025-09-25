package alertmanager

import (
	"reflect"
	"testing"
)

func TestAlert(t *testing.T) {
	tests := []struct {
		name                string
		buildAlert          func() *Alert
		expectedLabels      map[string]string
		expectedAnnotations map[string]string
	}{
		{
			name: "alert with labels and annotations",
			buildAlert: func() *Alert {
				return NewAlert(
					WithLabel("alertname", "test"),
					WithLabel("severity", "critical"),
					WithAnnotation("message", "Test message"),
					WithAnnotation("runbook", "https://example.com"),
				)
			},
			expectedLabels: map[string]string{
				"alertname": "test",
				"severity":  "critical",
			},
			expectedAnnotations: map[string]string{
				"message": "Test message",
				"runbook": "https://example.com",
			},
		},
		{
			name: "empty alert",
			buildAlert: func() *Alert {
				return NewAlert()
			},
			expectedLabels:      map[string]string{},
			expectedAnnotations: map[string]string{},
		},
		{
			name: "alert with only labels",
			buildAlert: func() *Alert {
				return NewAlert(
					WithLabel("service", "api"),
					WithLabel("environment", "production"),
				)
			},
			expectedLabels: map[string]string{
				"service":     "api",
				"environment": "production",
			},
			expectedAnnotations: map[string]string{},
		},
		{
			name: "alert with only annotations",
			buildAlert: func() *Alert {
				return NewAlert(
					WithAnnotation("summary", "Service is down"),
					WithAnnotation("description", "The API service is not responding"),
				)
			},
			expectedLabels: map[string]string{},
			expectedAnnotations: map[string]string{
				"summary":     "Service is down",
				"description": "The API service is not responding",
			},
		},
		{
			name: "alert with overwritten values",
			buildAlert: func() *Alert {
				return NewAlert(
					WithLabel("severity", "warning"),
					WithLabel("severity", "critical"), // should overwrite
					WithAnnotation("message", "first message"),
					WithAnnotation("message", "final message"), // should overwrite
				)
			},
			expectedLabels: map[string]string{
				"severity": "critical",
			},
			expectedAnnotations: map[string]string{
				"message": "final message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alert := tt.buildAlert()

			if !reflect.DeepEqual(alert.Labels, tt.expectedLabels) {
				t.Errorf("expected labels (%v), got (%v)", tt.expectedLabels, alert.Labels)
			}
			if !reflect.DeepEqual(alert.Annotations, tt.expectedAnnotations) {
				t.Errorf("expected annotations (%v), got (%v)", tt.expectedAnnotations, alert.Annotations)
			}
		})
	}
}