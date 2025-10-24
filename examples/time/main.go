package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"

	alertmanager "github.com/spectrocloud-labs/alertmanager-client-go"
)

func main() {
	logger := logr.Discard() // Use your preferred logger
	client := &http.Client{Timeout: 30 * time.Second}

	// Create Alertmanager client with base configuration
	am, err := alertmanager.NewAlertmanager(logger, client,
		alertmanager.WithEndpoint("http://localhost:9093"),
		alertmanager.WithBaseLabel("service", "time-demo"),
		alertmanager.WithBaseLabel("environment", "development"),
		alertmanager.WithBaseAnnotation("team", "platform"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	now := time.Now()

	// Alert 1: Started in the past - demonstrates startsAt is metadata only
	pastStartTime := now.Add(-10 * time.Minute)
	pastStartAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "PastStartAlert"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithStartsAt(pastStartTime),
		alertmanager.WithAnnotation("summary", "This alert started 10 minutes ago"),
		alertmanager.WithAnnotation("description", "StartsAt is metadata for historical tracking - shows 'firing for 10 minutes'"),
	)

	// Alert 2: Explicit endsAt - resolves in 1 minute
	quickResolveTime := now.Add(1 * time.Minute)
	quickResolveAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "QuickResolveAlert"),
		alertmanager.WithLabel("severity", "warning"),
		alertmanager.WithStartsAt(now),
		alertmanager.WithEndsAt(quickResolveTime),
		alertmanager.WithAnnotation("summary", "Resolves in 1 minute via endsAt"),
		alertmanager.WithAnnotation("description", "EndsAt is explicitly set to 1 minute from now"),
	)

	// Alert 3: No endsAt - resolves in 5 minutes (uses global resolve_timeout)
	globalTimeoutResolveTime := now.Add(5 * time.Minute)
	globalTimeoutAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "GlobalTimeoutAlert"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithStartsAt(now),
		alertmanager.WithAnnotation("summary", "No endsAt, uses global timeout"),
		alertmanager.WithAnnotation("description", "Will resolve after 5 minutes (global resolve_timeout from receipt time)"),
	)

	// Alert 4: Explicit endsAt - resolves in 10 minutes
	longResolveTime := now.Add(10 * time.Minute)
	longResolveAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "LongResolveAlert"),
		alertmanager.WithLabel("severity", "critical"),
		alertmanager.WithStartsAt(now),
		alertmanager.WithEndsAt(longResolveTime),
		alertmanager.WithAnnotation("summary", "Resolves in 10 minutes via endsAt"),
		alertmanager.WithAnnotation("description", "EndsAt is set to 10 minutes from now"),
	)

	alerts := []*alertmanager.Alert{pastStartAlert, quickResolveAlert, globalTimeoutAlert, longResolveAlert}

	fmt.Print("=== Time-Based Alerts Example ===\n\n")

	// Emit all alerts at once
	fmt.Printf("Current time: %s\n", now.Format(time.RFC3339))
	fmt.Println("Sending alerts to Alertmanager...")

	resp, err := am.Emit(alerts...)
	if err != nil {
		fmt.Printf("Failed to send alerts: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Alertmanager returned non-OK status: %d %s\n", resp.StatusCode, resp.Status)
		return
	}

	fmt.Printf("  ✓ Successfully sent %d alerts (Status: %d)\n", len(alerts), resp.StatusCode)

	fmt.Print("\nAlert behaviors:\n")
	fmt.Printf("  1. PastStartAlert - Started at %s (10 minutes ago)\n", pastStartTime.Format(time.RFC3339))
	fmt.Println("     Shows 'firing for 10 minutes' in UI (startsAt is metadata)")
	fmt.Printf("  2. QuickResolveAlert - Will resolve at %s (in 1 minute)\n", quickResolveTime.Format(time.RFC3339))
	fmt.Printf("  3. GlobalTimeoutAlert - Will resolve at ~%s (in 5 minutes)\n", globalTimeoutResolveTime.Format(time.RFC3339))
	fmt.Printf("  4. LongResolveAlert - Will resolve at %s (in 10 minutes)\n", longResolveTime.Format(time.RFC3339))

	fmt.Println("\n=== Summary ===")
	fmt.Println("✓ startsAt sets historical metadata (doesn't delay visibility)")
	fmt.Println("✓ endsAt controls when alerts auto-resolve")
	fmt.Println("✓ Without endsAt, resolve_timeout starts from receipt time, not startsAt")
	fmt.Println("\nCheck the Alertmanager web UI at http://localhost:9093")
	fmt.Println("Look for alerts with service=time-demo")
}
