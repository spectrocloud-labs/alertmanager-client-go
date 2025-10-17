# Audit Log Alerts Example

This example demonstrates using Alertmanager as an audit log sink for tracking CRUD operations on Kubernetes resources.

## What This Example Shows

This example simulates a realistic series of operations on a Kubernetes ConfigMap:

1. **CREATE** - User creates a ConfigMap
2. **UPDATE** - User updates the ConfigMap
3. **DELETE** - User deletes the ConfigMap
4. **CREATE** - User recreates the ConfigMap

Each operation:
- Is emitted as a separate alert immediately
- Has labels: `operation`, `resource_kind`, `resource_namespace`, `resource_name`, `user`, `operation_id`
- The `operation_id` label contains a unique timestamp (nanoseconds) to ensure each audit entry is distinct
- Uses Alertmanager's automatic `startsAt` (time when the alert is received)
- Has `endsAt` set to 1 hour in the future, so all alerts stay visible

## Key Insight

**Alertmanager deduplicates alerts based solely on their label set** - timing doesn't matter. Without a unique identifier, the two CREATE operations would be merged into a single alert.

**Solution**: This example adds an `operation_id` label with a unique timestamp (nanoseconds) to each alert, ensuring every audit log entry is distinct, even for repeated operations on the same resource.

## Running the Example

1. **Start Alertmanager (from examples directory):**
   ```bash
   cd ..
   docker-compose up -d
   cd audit
   ```

2. **Run the example:**
   ```bash
   go run main.go
   ```

   The program will emit all 4 alerts immediately.

3. **Check the results:**
   - Open http://localhost:9093 in your browser
   - Look for alerts with `service=audit-demo`
   - You should see all 4 operations as separate alerts
   - Each alert will have a unique `operation_id` label visible in the UI

4. **Clean up:**
   ```bash
   cd ..
   docker-compose down
   ```

## Best Practice for Audit Logs

For audit logging, always include a unique identifier in your alert labels:

```go
alertmanager.WithLabel("operation_id", fmt.Sprintf("%d", time.Now().UnixNano()))
```

This ensures:
- Every audit log entry is unique and won't be deduplicated
- You get a complete audit trail with all operations, even duplicates
- Each alert can be correlated back to a specific operation timestamp
