# Basic Alerts Example

This example demonstrates the basic usage of the alertmanager-client-go library, showing how to create and send simple alerts to Alertmanager.

## What This Example Shows

This example creates 3 basic alerts with different severities and configurations:

1. **HighCPUUsage** - Warning severity alert for CPU usage above 80%
2. **HighMemoryUsage** - Critical severity alert for memory usage above 95%
3. **DiskSpaceLow** - Warning severity alert for low disk space with custom labels

Each alert demonstrates:
- Setting labels (alertname, severity, custom labels)
- Adding annotations (summary, description)
- Using base labels and annotations from the Alertmanager client configuration

## Running the Example

1. **Start Alertmanager:**
   ```bash
   docker-compose up -d
   ```

2. **Verify Alertmanager is running:**
   ```bash
   curl http://localhost:9093/api/v2/status
   ```

3. **Run the example:**
   ```bash
   go run main.go
   ```

4. **View the alerts:**
   - Open http://localhost:9093 in your browser to see the Alertmanager web UI
   - You should see all 3 alerts with their labels and annotations
   - Each alert will have the base labels (`service=example-service`, `environment=development`) and base annotation (`team=platform`)

5. **Clean up:**
   ```bash
   docker-compose down
   ```

## Key Concepts

- **Labels**: Key-value pairs used to identify and group alerts (e.g., `alertname`, `severity`)
- **Annotations**: Additional metadata for alerts (e.g., `summary`, `description`)
- **Base Labels/Annotations**: Applied to all alerts sent through the Alertmanager client
- **Emit()**: Sends one or more alerts to Alertmanager in a single API call
