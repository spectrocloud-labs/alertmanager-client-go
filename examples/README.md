# Examples

This directory contains examples demonstrating various features of the alertmanager-client-go library.

## Prerequisites

All examples require Alertmanager to be running. Start it using Docker Compose:

```bash
cd examples
docker-compose up -d
```

Verify Alertmanager is running:
```bash
curl http://localhost:9093/api/v2/status
```

View alerts in the web UI: http://localhost:9093

When finished, stop Alertmanager:
```bash
docker-compose down
```

## Available Examples

### 1. Basic Example (`basic/`)

Demonstrates the basic usage of the library, showing how to create and send simple alerts to Alertmanager.

**What it shows:**
- Creating alerts with different severities (warning, critical)
- Setting labels (alertname, severity, custom labels)
- Adding annotations (summary, description)
- Using base labels and annotations from the client configuration
- Sending multiple alerts in a single API call

**Alerts created:**
1. **HighCPUUsage** - Warning severity alert for CPU usage above 80%
2. **HighMemoryUsage** - Critical severity alert for memory usage above 95%
3. **DiskSpaceLow** - Warning severity alert for low disk space with custom labels

**Run it:**
```bash
go run ./basic
```

**Key concepts:**
- **Labels**: Key-value pairs used to identify and group alerts (e.g., `alertname`, `severity`)
- **Annotations**: Additional metadata for alerts (e.g., `summary`, `description`)
- **Base Labels/Annotations**: Applied to all alerts sent through the client
- **Emit()**: Sends one or more alerts to Alertmanager in a single API call

---

### 2. Audit Log Example (`audit/`)

Demonstrates using Alertmanager as an audit log sink for tracking CRUD operations on Kubernetes resources.

**What it shows:**
- Emitting alerts immediately as audit log entries
- Using unique identifiers to prevent deduplication
- Tracking operations with structured labels
- Setting `endsAt` to keep audit entries visible

**Operations tracked:**
1. **CREATE** - User creates a ConfigMap
2. **UPDATE** - User updates the ConfigMap
3. **DELETE** - User deletes the ConfigMap
4. **CREATE** - User recreates the ConfigMap (demonstrating why unique IDs are needed)

**Run it:**
```bash
go run ./audit
```

**Key insight:**

Alertmanager deduplicates alerts based solely on their label set - timing doesn't matter. Without a unique identifier, the two CREATE operations would be merged into a single alert.

**Solution:** Include an `operation_id` label with a unique timestamp:

```go
alertmanager.WithLabel("operation_id", fmt.Sprintf("%d", time.Now().UnixNano()))
```

This ensures:
- Every audit log entry is unique and won't be deduplicated
- You get a complete audit trail with all operations, even duplicates
- Each alert can be correlated back to a specific operation timestamp

---

### 3. Time-Based Alerts Example (`time/`)

Demonstrates how to use `WithStartsAt` and `WithEndsAt` options to control alert timestamps and auto-resolution.

**What it shows:**
- Using `startsAt` for historical tracking
- Using `endsAt` for explicit auto-resolution timing
- Understanding the difference between alert metadata and resolution behavior
- How global `resolve_timeout` works when `endsAt` is not set

**Alerts created:**
1. **PastStartAlert** - `startsAt` set to 10 minutes ago (historical tracking)
2. **QuickResolveAlert** - Explicit `endsAt` 1 minute in the future
3. **GlobalTimeoutAlert** - No `endsAt` set, uses global `resolve_timeout` (5 minutes)
4. **LongResolveAlert** - Explicit `endsAt` 10 minutes in the future

**Run it:**
```bash
go run ./time
```

**Behavior to observe:**
- All 4 alerts appear immediately (`startsAt` doesn't delay visibility)
- PastStartAlert shows it's been "firing for 10 minutes" (historical metadata)
- QuickResolveAlert resolves in 1 minute
- GlobalTimeoutAlert resolves in 5 minutes
- LongResolveAlert resolves in 10 minutes

**Key concepts:**

- **`WithStartsAt(time.Time)`**: Records when the alert condition began. This is **purely metadata** for historical tracking:
  - Alerts appear in the UI immediately regardless of this timestamp
  - The UI shows "firing for X duration" based on this timestamp
  - If omitted, Alertmanager uses the current time
  - **Does NOT affect** when `resolve_timeout` starts counting

- **`WithEndsAt(time.Time)`**: Controls when an alert auto-resolves:
  - Alerts with explicit `endsAt` times will resolve at that time
  - If omitted, the alert resolves after `resolve_timeout` from when Alertmanager **receives** the alert
  - The `resolve_timeout` is NOT calculated from `startsAt`

- **Important behaviors**:
  - `startsAt` does NOT delay when alerts appear
  - Without `endsAt`, timeout starts from receipt time, not `startsAt`
  - Only `endsAt` controls actual resolution timing

## Running All Examples

To run all examples in sequence:

```bash
cd examples
docker-compose up -d

go run ./basic
go run ./audit
go run ./time

docker-compose down
```

## Common Configuration

All examples use a shared `go.mod` file in the `examples/` directory and share the same Docker Compose configuration for Alertmanager.

The Alertmanager configuration (`alertmanager.yaml`) sets a global `resolve_timeout` of 5 minutes, which affects alerts that don't specify an explicit `endsAt` time.
