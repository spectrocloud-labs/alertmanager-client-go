# alertmanager-client-go

A Go client library for sending alerts to Prometheus Alertmanager using a clean, functional options API.

## Installation

```bash
go get github.com/spectrocloud-labs/alertmanager-client-go
```

## Quick Start

```go
package main

import (
    "net/http"
    "time"

    "github.com/go-logr/logr"
    "github.com/spectrocloud-labs/alertmanager-client-go/pkg/alertmanager"
)

func main() {
    logger := logr.Discard() // Use your preferred logger
    client := &http.Client{Timeout: 30 * time.Second}

    // Create Alertmanager client with base configuration
    am, err := alertmanager.NewAlertmanager(logger, client,
        alertmanager.WithEndpoint("http://alertmanager:9093"),
        alertmanager.WithBasicAuth("user", "password"),
        alertmanager.WithLabel("service", "my-service"),     // Applied to all alerts
        alertmanager.WithLabel("environment", "production"), // Applied to all alerts
        alertmanager.WithAnnotation("team", "platform"),     // Applied to all alerts
    )
    if err != nil {
        panic(err)
    }

    // Create multiple alerts
    cpuAlert := alertmanager.NewAlert().
        WithLabel("alertname", "HighCPUUsage").             // Specific to this alert
        WithLabel("severity", "warning").                   // Specific to this alert
        WithLabel("instance", "web-01").                    // Specific to this alert
        WithAnnotation("summary", "CPU usage is above 80%") // Specific to this alert

    memoryAlert := alertmanager.NewAlert().
        WithLabel("alertname", "HighMemoryUsage").             // Specific to this alert
        WithLabel("severity", "critical").                     // Specific to this alert
        WithLabel("instance", "web-01").                       // Specific to this alert
        WithAnnotation("summary", "Memory usage is above 95%") // Specific to this alert

    // Emit alerts
    err = am.Emit(cpuAlert, memoryAlert)
    if err != nil {
        panic(err)
    }
}
```

## Developer Guide

### Running the Example

1. **Start Alertmanager using Docker Compose:**
   ```bash
   # Start Alertmanager with the provided configuration
   docker-compose up -d

   # Check that Alertmanager is running
   curl http://localhost:9093/api/v1/status
   ```

2. **Run the example:**
   ```bash
   cd examples/basic
   go run main.go
   ```

3. **Verify alerts were received:**
   - Open http://localhost:9093 in your browser to see the Alertmanager web UI
   - You should see 3 alerts: HighCPUUsage, HighMemoryUsage, and DiskSpaceLow
   - All alerts will have the base labels (`service=example-service`, `environment=development`) and annotation (`team=platform`)

4. **Clean up:**
   ```bash
   # Stop Alertmanager
   docker-compose down
   ```

### Development Setup

1. **Clone the repository:**
   ```bash
   git clone https://github.com/spectrocloud-labs/alertmanager-client-go.git
   cd alertmanager-client-go
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run tests:**
   ```bash
   go test ./...
   ```

### Testing Against Real Alertmanager

The repository includes a complete Docker setup for testing:

