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
    "github.com/spectrocloud-labs/alertmanager-client-go/alertmanager"
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

### Testing with a Local Alertmanager

For development and testing, you can run Alertmanager locally using Docker:

```bash
# Run Alertmanager in Docker
docker run -p 9093:9093 prom/alertmanager

# Test your integration
go run examples/basic/main.go
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

### Creating Examples

Create example programs in the `examples/` directory to test functionality:

```go
// examples/basic/main.go
package main

import (
    "fmt"
    "net/http"
    "time"

    "github.com/go-logr/logr"
    "github.com/spectrocloud-labs/alertmanager-client-go/alertmanager"
)

func main() {
    logger := logr.Discard()
    client := &http.Client{Timeout: 10 * time.Second}

    am, err := alertmanager.NewAlertmanager(logger, client,
        alertmanager.WithEndpoint("http://localhost:9093"),
        alertmanager.WithLabel("environment", "dev"),
    )
    if err != nil {
        panic(err)
    }

    alert := alertmanager.NewAlert().
        WithLabel("alertname", "TestAlert").
        WithLabel("severity", "info").
        WithAnnotation("summary", "This is a test alert")

    fmt.Println("Sending test alert...")
    err = am.Emit(alert)
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Println("Alert sent successfully!")
}
```

### Testing Against Real Alertmanager

For integration testing, you can use the provided Docker Compose setup:

```yaml
# docker-compose.yml
version: '3'
services:
  alertmanager:
    image: prom/alertmanager
    ports:
      - "9093:9093"
    volumes:
      - ./alertmanager.yml:/etc/alertmanager/alertmanager.yml
```

```yaml
# alertmanager.yml
global:
  smtp_smarthost: 'localhost:587'
  smtp_from: 'alertmanager@example.org'

route:
  group_by: ['alertname']
  group_wait: 10s
  group_interval: 10s
  repeat_interval: 1h
  receiver: 'web.hook'

receivers:
- name: 'web.hook'
  webhook_configs:
  - url: 'http://127.0.0.1:5001/'
```

### Debugging

Enable debug logging to see HTTP requests:

```go
import (
    "os"
    "github.com/go-logr/logr"
    "github.com/go-logr/zapr"
    "go.uber.org/zap"
)

// Create a debug logger
zapLog, _ := zap.NewDevelopment()
logger := zapr.NewLogger(zapLog)

am, err := alertmanager.NewAlertmanager(logger, client,
    alertmanager.WithEndpoint("http://localhost:9093"),
)
```

### Contributing

1. **Fork the repository**
2. **Create a feature branch:** `git checkout -b feature/amazing-feature`
3. **Write tests** for your changes
4. **Run the test suite:** `go test ./...`
5. **Commit your changes:** `git commit -m 'Add amazing feature'`
6. **Push to the branch:** `git push origin feature/amazing-feature`
7. **Open a Pull Request**

### Code Style

This project follows standard Go conventions:

- Use `gofmt` to format your code
- Run `go vet` to check for issues
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Write table-driven tests where appropriate
- Include examples in your documentation

## License

This project is licensed under the MIT License - see the LICENSE file for details.

