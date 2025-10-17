package alertmanager

import "time"

// Alert is an Alertmanager alert.
type Alert struct {
	// Annotations are arbitrary key-value pairs.
	Annotations map[string]string `json:"annotations"`

	// Labels are key-value pairs that can be used to group and filter alerts.
	Labels map[string]string `json:"labels"`

	// StartsAt is the time the alert started firing. If omitted, the current time is used.
	StartsAt *time.Time `json:"startsAt,omitempty"`

	// EndsAt is the time the alert should be considered resolved.
	// If omitted, the alert will be resolved after the global resolve_timeout.
	EndsAt *time.Time `json:"endsAt,omitempty"`
}

// AlertOption is a functional option for configuring an Alert.
type AlertOption func(*Alert)

// NewAlert creates a new Alert with initialized maps and applies the given options.
func NewAlert(options ...AlertOption) *Alert {
	alert := &Alert{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}

	for _, opt := range options {
		opt(alert)
	}

	return alert
}

// WithLabel adds a label to an Alert.
func WithLabel(key, value string) AlertOption {
	return func(a *Alert) {
		if a.Labels == nil {
			a.Labels = make(map[string]string)
		}
		a.Labels[key] = value
	}
}

// WithAnnotation adds an annotation to an Alert.
func WithAnnotation(key, value string) AlertOption {
	return func(a *Alert) {
		if a.Annotations == nil {
			a.Annotations = make(map[string]string)
		}
		a.Annotations[key] = value
	}
}

// WithStartsAt sets the start time of an Alert.
func WithStartsAt(t time.Time) AlertOption {
	return func(a *Alert) {
		a.StartsAt = &t
	}
}

// WithEndsAt sets the end time of an Alert.
func WithEndsAt(t time.Time) AlertOption {
	return func(a *Alert) {
		a.EndsAt = &t
	}
}
