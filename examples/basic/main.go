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
		alertmanager.WithBaseLabel("service", "example-service"),
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
