package alertmanager

import (
	"net/http"
	"time"
)

// Client represents an HTTP client interface for Alertmanager operations.
type Client struct {
	hclient *http.Client
}

// NewClient creates a new Client with the specified timeout.
func NewClient(timeout time.Duration) *Client {
	return &Client{
		hclient: &http.Client{
			Timeout: timeout,
		},
	}
}
