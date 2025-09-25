# alertmanager-client-go

A Go client library for sending alerts to Prometheus Alertmanager.

## Installation

```bash
go get github.com/spectrocloud-labs/alertmanager-client-go
```

## Developer Guide

### Running the Example

1. **Start Alertmanager using Docker Compose:**
   ```bash
   cd examples/basic

   # Start Alertmanager with the provided configuration
   docker-compose up -d

   # Check that Alertmanager is running
   curl http://localhost:9093/api/v2/status
   ```

2. **Run the example:**
   ```bash
   # From the examples/basic directory
   go run main.go
   ```

3. **Verify alerts were received:**
   - Open http://localhost:9093 in your browser to see the Alertmanager web UI
   - You should see 3 alerts: HighCPUUsage, HighMemoryUsage, and DiskSpaceLow
   - All alerts will have the base labels (`service=example-service`, `environment=development`) and annotation (`team=platform`)

4. **Clean up:**
   ```bash
   # Stop Alertmanager (from examples/basic directory)
   docker-compose down
   ```
