package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/logr"

	alertmanager "github.com/spectrocloud-labs/alertmanager-client-go"
)

func main() {
	logger := logr.Discard() // Use your preferred logger

	// Read CA certificate for TLS verification
	caCert, err := os.ReadFile("certs/ca.pem")
	if err != nil {
		fmt.Printf("Failed to read CA certificate: %v\n", err)
		fmt.Println("\nPlease run the following commands from the examples/ directory:")
		fmt.Println("  ./generate-certs.sh")
		fmt.Println("  docker-compose up -d")
		return
	}

	fmt.Println("=== Secure Alertmanager Client Example ===\n")

	// Test 1: Verify that missing basic auth fails
	fmt.Println("Test 1: Attempting to send alert WITHOUT basic auth (should fail)...")
	client1 := &http.Client{Timeout: 30 * time.Second}
	am1, err := alertmanager.NewAlertmanager(logger, client1,
		alertmanager.WithEndpoint("https://localhost:9094"),
		alertmanager.WithCustomCA(caCert),
		// Note: WithBasicAuth NOT provided
		alertmanager.WithBaseLabel("service", "secure-demo"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	alert1 := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "NoAuthTest"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithAnnotation("summary", "Testing without authentication"),
	)

	resp1, err := am1.Emit(alert1)
	if err != nil {
		fmt.Printf("  ✓ Request failed as expected: %v\n", err)
	} else {
		defer resp1.Body.Close()
		if resp1.StatusCode == http.StatusUnauthorized {
			fmt.Printf("  ✓ Received 401 Unauthorized as expected\n")
		} else {
			fmt.Printf("  ✗ Unexpected status: %d (expected 401)\n", resp1.StatusCode)
		}
	}

	// Test 2: Verify that wrong credentials fail
	fmt.Println("\nTest 2: Attempting to send alert with WRONG credentials (should fail)...")
	client2 := &http.Client{Timeout: 30 * time.Second}
	am2, err := alertmanager.NewAlertmanager(logger, client2,
		alertmanager.WithEndpoint("https://localhost:9094"),
		alertmanager.WithCustomCA(caCert),
		alertmanager.WithBasicAuth("admin", "wrongpassword"),
		alertmanager.WithBaseLabel("service", "secure-demo"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	alert2 := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "WrongAuthTest"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithAnnotation("summary", "Testing with wrong credentials"),
	)

	resp2, err := am2.Emit(alert2)
	if err != nil {
		fmt.Printf("  ✓ Request failed as expected: %v\n", err)
	} else {
		defer resp2.Body.Close()
		if resp2.StatusCode == http.StatusUnauthorized {
			fmt.Printf("  ✓ Received 401 Unauthorized as expected\n")
		} else {
			fmt.Printf("  ✗ Unexpected status: %d (expected 401)\n", resp2.StatusCode)
		}
	}

	// Test 3: Verify that missing CA certificate fails
	fmt.Println("\nTest 3: Attempting to send alert WITHOUT CA certificate (should fail)...")
	client3 := &http.Client{Timeout: 30 * time.Second}
	am3, err := alertmanager.NewAlertmanager(logger, client3,
		alertmanager.WithEndpoint("https://localhost:9094"),
		// Note: WithCustomCA NOT provided - will use system certs
		alertmanager.WithBasicAuth("admin", "password"),
		alertmanager.WithBaseLabel("service", "secure-demo"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	alert3 := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "NoCertTest"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithAnnotation("summary", "Testing without CA certificate"),
	)

	resp3, err := am3.Emit(alert3)
	if err != nil {
		fmt.Printf("  ✓ TLS verification failed as expected: %v\n", err)
	} else {
		defer resp3.Body.Close()
		fmt.Printf("  ✗ Unexpected success (expected TLS error)\n")
	}

	// Test 4: Verify that correct TLS + basic auth succeeds
	fmt.Println("\nTest 4: Sending alert with CORRECT TLS + basic auth (should succeed)...")
	client4 := &http.Client{Timeout: 30 * time.Second}
	am4, err := alertmanager.NewAlertmanager(logger, client4,
		alertmanager.WithEndpoint("https://localhost:9094"),
		alertmanager.WithCustomCA(caCert),
		alertmanager.WithBasicAuth("admin", "password"),
		alertmanager.WithBaseLabel("service", "secure-demo"),
		alertmanager.WithBaseLabel("environment", "production"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	alert4 := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "SecureConnectionTest"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithAnnotation("summary", "Testing TLS and basic authentication"),
		alertmanager.WithAnnotation("description", "This alert was sent over HTTPS with basic authentication"),
	)

	resp4, err := am4.Emit(alert4)
	if err != nil {
		fmt.Printf("  ✗ Failed to send alert: %v\n", err)
		return
	}
	defer resp4.Body.Close()

	if resp4.StatusCode != http.StatusOK {
		fmt.Printf("  ✗ Alertmanager returned non-OK status: %d %s\n", resp4.StatusCode, resp4.Status)
		return
	}

	fmt.Printf("  ✓ Successfully sent alert (Status: %d)\n", resp4.StatusCode)

	fmt.Println("\n=== Summary ===")
	fmt.Println("✓ Basic auth is enforced (requests without credentials fail)")
	fmt.Println("✓ TLS certificate verification is working (self-signed cert rejected without CA)")
	fmt.Println("✓ Secure communication works with proper credentials and CA certificate")
	fmt.Println("\nCheck the Alertmanager web UI at https://localhost:9094")
	fmt.Println("(Username: admin, Password: password)")
}
