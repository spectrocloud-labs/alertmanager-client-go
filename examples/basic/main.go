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
		alertmanager.WithBaseLabel("service", "basic-demo"),
		alertmanager.WithBaseLabel("environment", "development"),
		alertmanager.WithBaseAnnotation("team", "platform"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	// Create multiple alerts
	cpuAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "HighCPUUsage"),
		alertmanager.WithLabel("severity", "warning"),
		alertmanager.WithAnnotation("summary", "CPU usage is above 80%"),
		alertmanager.WithAnnotation("description", "The CPU usage on web-01 has been above 80% for more than 5 minutes"),
	)

	memoryAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "HighMemoryUsage"),
		alertmanager.WithLabel("severity", "critical"),
		alertmanager.WithAnnotation("summary", "Memory usage is above 95%"),
		alertmanager.WithAnnotation("description", "The memory usage on web-01 has reached critical levels"),
	)

	diskAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "DiskSpaceLow"),
		alertmanager.WithLabel("severity", "warning"),
		alertmanager.WithLabel("mountpoint", "/var/log"),
		alertmanager.WithAnnotation("summary", "Disk space is running low"),
		alertmanager.WithAnnotation("description", "Only 10% disk space remaining on /var/log"),
	)

	alerts := []*alertmanager.Alert{cpuAlert, memoryAlert, diskAlert}

	fmt.Print("=== Basic Alertmanager Client Example ===\n\n")

	// Emit all alerts at once
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

	fmt.Println("\nAlerts sent:")
	fmt.Println("  1. HighCPUUsage - Warning severity")
	fmt.Println("  2. HighMemoryUsage - Critical severity")
	fmt.Println("  3. DiskSpaceLow - Warning severity with custom mountpoint label")

	fmt.Println("\n=== Summary ===")
	fmt.Println("✓ Created and sent 3 alerts with different severities")
	fmt.Println("✓ Used base labels (service, environment) applied to all alerts")
	fmt.Println("✓ Added custom labels and annotations per alert")
	fmt.Println("\nCheck the Alertmanager web UI at http://localhost:9093")
	fmt.Println("Look for alerts with service=basic-demo")

}
