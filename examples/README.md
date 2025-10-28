# Examples

This directory contains examples demonstrating various features of the alertmanager-client-go library.

## Prerequisites

All examples require Alertmanager to be running. The setup includes two instances:
- **HTTP** instance on port 9093 (for basic, audit, and time examples)
- **HTTPS** instance on port 9094 with TLS and basic auth (for secure example)

### Setup

Generate TLS certificates (required for the secure example):
```bash
cd examples
./generate-certs.sh
```

Start both Alertmanager instances:
```bash
docker-compose up -d
```

Verify both instances are running:
```bash
# HTTP instance
curl http://localhost:9093/api/v2/status

# HTTPS instance (with basic auth)
curl -k -u admin:password https://localhost:9094/api/v2/status
```

View alerts in the web UIs:
- HTTP: <http://localhost:9093>
- HTTPS: <https://localhost:9094> (Username: admin, Password: password)

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
2. **UPDATE** - User updates the ConfigMap (first update)
3. **UPDATE** - User updates the ConfigMap again (second update)
4. **DELETE** - User deletes the ConfigMap
5. **CREATE** - User recreates the ConfigMap (demonstrating why unique IDs are needed)

**Run it:**
```bash
go run ./audit
```

**Key insight:**

Alertmanager deduplicates alerts based solely on their label set - timing doesn't matter. Without a unique identifier, the two CREATE operations would be merged into a single alert.

**Solution:** Include an `audit_id` label with a unique value (cryptographically secure random ID):

```go
alertmanager.WithLabel("audit_id", auditID)
```

This ensures:
- Every audit log entry is unique and won't be deduplicated
- You get a complete audit trail with all operations, even duplicates
- Each alert can be correlated back to a specific operation via its unique audit ID

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

---

### 4. Secure Connection Example (`secure/`)

Demonstrates how to use TLS and basic authentication for secure communication with Alertmanager. This example shows **two approaches** for configuring secure connections:
1. Using functional options with `NewAlertmanager` (maximum flexibility)
2. Using the Args struct with `NewAlertmanagerWithArgs` (convenient for config files)

**What it shows:**
- Configuring TLS with a custom CA certificate
- Setting up basic authentication credentials
- Validating that authentication is enforced
- Validating that TLS certificate verification is working
- Enforcing TLS 1.3 minimum version
- Two different configuration approaches for the same secure setup

#### Example 1: Using Options Pattern

**Tests performed:**
1. **Test 1** - Attempts to send alert WITHOUT basic auth credentials (expects 401 Unauthorized)
2. **Test 2** - Attempts to send alert with WRONG password (expects 401 Unauthorized)
3. **Test 3** - Attempts to send alert WITHOUT CA certificate (expects TLS verification failure)
4. **Test 4** - Attempts to send alert with TLS 1.2 only (expects TLS protocol version error)
5. **Test 5** - Sends alert with CORRECT TLS 1.3 + basic auth + CA cert (expects success)

#### Example 2: Using Args Constructor

Demonstrates the same secure connection using the `NewAlertmanagerWithArgs` constructor:
- Configures all security options through the `Args` struct
- Perfect for loading configuration from YAML/JSON files or environment variables
- Shows how to configure username, password, TLS CA cert path, and TLS version requirements

**Run it:**
```bash
go run ./secure
```

**Expected output:**
All tests should pass, validating that:
- Basic auth is enforced (Tests 1 & 2 fail with 401)
- TLS certificate verification is working (Test 3 fails with TLS error)
- TLS 1.2 is rejected by server (Test 4 fails with protocol version error)
- Secure communication succeeds with TLS 1.3 + proper credentials (Test 5 & Args test succeed)

**Key concepts:**

#### Options Pattern (NewAlertmanager)

- **`WithCustomCA(caCert []byte)`**: Configures TLS to trust a custom CA certificate
  - Use this when Alertmanager uses certificates signed by a private CA
  - The client will verify the server's certificate against this CA
  - More secure than `WithInsecure(true)` which skips verification entirely
  - Without this option, self-signed certificates are rejected by the system cert pool

- **`WithBasicAuth(username, password string)`**: Sets HTTP basic authentication credentials
  - Credentials are sent in the Authorization header with base64 encoding
  - Should always be used with TLS to prevent credential leakage
  - Common in production Alertmanager deployments
  - Requests without credentials or with wrong credentials receive 401 Unauthorized

- **`WithMinTLSVersion(minVersion TLSVersion)`**: Sets the minimum TLS version for connections
  - Use constants like `TLS12`, `TLS13`
  - If not specified, TLS 1.2 is the default minimum

- **`WithMaxTLSVersion(maxVersion TLSVersion)`**: Sets the maximum TLS version for connections
  - Use constants like `TLS12`, `TLS13`
  - Useful for enforcing specific TLS versions or preventing negotiation to higher versions

#### Args Pattern (NewAlertmanagerWithArgs)

- **`Args` struct**: Convenient struct-based configuration
  - `Enabled`: Toggle client on/off without changing config
  - `AlertmanagerURL`: The Alertmanager endpoint
  - `Username` / `Password`: Basic auth credentials
  - `TLSCACertPath`: Path to CA certificate file (loaded from disk)
  - `TLSMinVersion` / `TLSMaxVersion`: String versions like "TLS12", "TLS13"
  - `TLSInsecureSkipVerify`: Skip TLS verification (not recommended)
  - `ProxyURL`: HTTP proxy URL
  - `Timeout`: Request timeout (defaults to 2 seconds)

**Note:** The example uses port 9094 which runs a separate Alertmanager instance configured with TLS and basic auth, while other examples use port 9093 with plain HTTP.

## Running All Examples

To run all examples in sequence:

```bash
cd examples

# Generate certificates and start Alertmanager instances
./generate-certs.sh
docker-compose up -d

# Run all examples
go run ./basic
go run ./audit
go run ./time
go run ./secure

# Stop Alertmanager
docker-compose down
```

## Common Configuration

All examples use a shared `go.mod` file in the `examples/` directory and share the same Docker Compose configuration for Alertmanager.

The Alertmanager configuration (`alertmanager.yaml`) sets a global `resolve_timeout` of 5 minutes, which affects alerts that don't specify an explicit `endsAt` time.
