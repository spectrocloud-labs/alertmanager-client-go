package main

import (
	"fmt"
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
		alertmanager.WithEndpoint("http://localhost:9093"),
		alertmanager.WithLabel("service", "example-service"),
		alertmanager.WithLabel("environment", "development"),
		alertmanager.WithAnnotation("team", "platform"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	// Create multiple alerts
	cpuAlert := alertmanager.NewAlert().
		WithLabel("alertname", "HighCPUUsage").
		WithLabel("severity", "warning").
		WithLabel("instance", "web-01").
		WithAnnotation("summary", "CPU usage is above 80%").
		WithAnnotation("description", "The CPU usage on web-01 has been above 80% for more than 5 minutes")

	memoryAlert := alertmanager.NewAlert().
		WithLabel("alertname", "HighMemoryUsage").
		WithLabel("severity", "critical").
		WithLabel("instance", "web-01").
		WithAnnotation("summary", "Memory usage is above 95%").
		WithAnnotation("description", "The memory usage on web-01 has reached critical levels")

	diskAlert := alertmanager.NewAlert().
		WithLabel("alertname", "DiskSpaceLow").
		WithLabel("severity", "warning").
		WithLabel("instance", "web-01").
		WithLabel("mountpoint", "/var/log").
		WithAnnotation("summary", "Disk space is running low").
		WithAnnotation("description", "Only 10% disk space remaining on /var/log")

	// Emit all alerts at once
	fmt.Println("Sending alerts to Alertmanager...")
	err = am.Emit(cpuAlert, memoryAlert, diskAlert)
	if err != nil {
		fmt.Printf("Failed to send alerts: %v\n", err)
		return
	}

	fmt.Println("✅ Successfully sent 3 alerts to Alertmanager!")
	fmt.Println("Check the Alertmanager web UI at http://localhost:9093 to see your alerts.")
}