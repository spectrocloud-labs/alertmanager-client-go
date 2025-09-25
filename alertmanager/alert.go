package alertmanager

// Alert is an Alertmanager alert.
type Alert struct {
	// Annotations are arbitrary key-value pairs.
	Annotations map[string]string `json:"annotations"`

	// Labels are key-value pairs that can be used to group and filter alerts.
	Labels map[string]string `json:"labels"`
}

// NewAlert creates a new Alert with initialized maps.
func NewAlert() *Alert {
	return &Alert{
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
}

// WithLabel adds a label to this Alert.
func (a *Alert) WithLabel(key, value string) *Alert {
	a.Labels[key] = value
	return a
}

// WithAnnotation adds an annotation to this Alert.
func (a *Alert) WithAnnotation(key, value string) *Alert {
	a.Annotations[key] = value
	return a
}

