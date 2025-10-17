# Time-Based Alerts Example

This example demonstrates how to use the `WithStartsAt` and `WithEndsAt` options to control alert timestamps and auto-resolution.

## What This Example Shows

This example creates 4 alerts demonstrating different timing behaviors:

### startsAt Example (Historical Tracking)
1. **PastStartAlert** - `startsAt` set to 10 minutes ago, demonstrates historical tracking

### endsAt Examples (Auto-Resolution)
2. **QuickResolveAlert** - Explicit `endsAt` 1 minute in the future
3. **GlobalTimeoutAlert** - No `endsAt` set, uses global `resolve_timeout` (5 minutes)
4. **LongResolveAlert** - Explicit `endsAt` 10 minutes in the future

## Running the Example

1. **Start Alertmanager:**
   ```bash
   docker-compose up -d
   ```

2. **Run the example:**
   ```bash
   go run main.go
   ```

3. **Watch the behavior:**
   - Open http://localhost:9093 in your browser
   - All 4 alerts appear immediately (startsAt doesn't delay visibility)
   - PastStartAlert shows it's been "firing for 10 minutes" (historical tracking)
   - QuickResolveAlert resolves in 1 minute
   - GlobalTimeoutAlert resolves in 5 minutes
   - LongResolveAlert resolves in 10 minutes

4. **Clean up:**
   ```bash
   docker-compose down
   ```

## Key Concepts

- **`WithStartsAt(time.Time)`**: Records when the alert condition began. This is **purely metadata** for historical tracking:
  - Alerts appear in the UI immediately regardless of this timestamp
  - The UI shows "firing for X duration" based on this timestamp
  - If omitted, Alertmanager uses the current time
  - **Does NOT affect** when `resolve_timeout` starts counting

- **`WithEndsAt(time.Time)`**: Controls when an alert auto-resolves:
  - Alerts with explicit `endsAt` times will resolve at that time
  - If omitted, the alert resolves after `resolve_timeout` from when Alertmanager **receives** the alert
  - The `resolve_timeout` is NOT calculated from `startsAt`

- **Important Behaviors**:
  - `startsAt` does NOT delay when alerts appear
  - Without `endsAt`, timeout starts from receipt time, not `startsAt`
  - Only `endsAt` controls actual resolution timing
