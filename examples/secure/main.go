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

	fmt.Print("=== Secure Alertmanager Client Example ===\n\n")
	fmt.Println("This example demonstrates two ways to configure secure connections:")
	fmt.Print("1. Using NewAlertmanager with functional options\n")
	fmt.Print("2. Using NewAlertmanagerWithArgs with Args struct\n\n")

	fmt.Print("--- Example 1: Using Options Pattern ---\n\n")
	demonstrateOptionsPattern(logger, caCert)

	fmt.Print("\n\n--- Example 2: Using Args Constructor ---\n\n")
	demonstrateArgsConstructor(logger, caCert)

	fmt.Print("\n\n=== Summary ===\n")
	fmt.Println("Both approaches support full security configuration:")
	fmt.Println("✓ TLS 1.3 with custom CA certificates")
	fmt.Println("✓ Basic authentication")
	fmt.Println("✓ Certificate verification")
	fmt.Print("\nCheck the Alertmanager web UI at https://localhost:9094\n")
	fmt.Println("(Username: admin, Password: password)")
	fmt.Println("Look for alerts with service=secure-demo or service=secure-args-demo")
}

func demonstrateOptionsPattern(logger logr.Logger, caCert []byte) {

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

	// Test 4: Verify that TLS 1.2 is rejected by server
	fmt.Println("\nTest 4: Attempting to send alert with TLS 1.2 (should fail)...")
	client4 := &http.Client{Timeout: 30 * time.Second}
	am4, err := alertmanager.NewAlertmanager(logger, client4,
		alertmanager.WithEndpoint("https://localhost:9094"),
		alertmanager.WithCustomCA(caCert),
		alertmanager.WithBasicAuth("admin", "password"),
		alertmanager.WithMinTLSVersion(alertmanager.TLS12),
		alertmanager.WithMaxTLSVersion(alertmanager.TLS12),
		alertmanager.WithBaseLabel("service", "secure-demo"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	alert4 := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "TLS12Test"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithAnnotation("summary", "Testing with TLS 1.2"),
	)

	resp4, err := am4.Emit(alert4)
	if err != nil {
		fmt.Printf("  ✓ TLS 1.2 rejected as expected: %v\n", err)
	} else {
		defer resp4.Body.Close()
		fmt.Printf("  ✗ Unexpected success with TLS 1.2 (server should reject)\n")
	}

	// Test 5: Verify that correct TLS 1.3 + basic auth succeeds
	fmt.Println("\nTest 5: Sending alert with CORRECT TLS 1.3 + basic auth (should succeed)...")
	client5 := &http.Client{Timeout: 30 * time.Second}
	am5, err := alertmanager.NewAlertmanager(logger, client5,
		alertmanager.WithEndpoint("https://localhost:9094"),
		alertmanager.WithCustomCA(caCert),
		alertmanager.WithBasicAuth("admin", "password"),
		alertmanager.WithMinTLSVersion(alertmanager.TLS13),
		alertmanager.WithBaseLabel("service", "secure-demo"),
		alertmanager.WithBaseLabel("environment", "production"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	alert5 := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "SecureConnectionTest"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithAnnotation("summary", "Testing TLS and basic authentication"),
		alertmanager.WithAnnotation("description", "This alert was sent over HTTPS with basic authentication"),
	)

	resp5, err := am5.Emit(alert5)
	if err != nil {
		fmt.Printf("  ✗ Failed to send alert: %v\n", err)
		return
	}
	defer resp5.Body.Close()

	if resp5.StatusCode != http.StatusOK {
		fmt.Printf("  ✗ Alertmanager returned non-OK status: %d %s\n", resp5.StatusCode, resp5.Status)
		return
	}

	fmt.Printf("  ✓ Successfully sent alert (Status: %d)\n", resp5.StatusCode)

	fmt.Println("\nOptions pattern benefits:")
	fmt.Println("✓ Basic auth is enforced (requests without credentials fail)")
	fmt.Println("✓ TLS certificate verification is working (self-signed cert rejected without CA)")
	fmt.Println("✓ TLS 1.2 is rejected by server (only TLS 1.3+ accepted)")
	fmt.Println("✓ Secure communication works with TLS 1.3 + basic auth + CA cert")
}

func demonstrateArgsConstructor(logger logr.Logger, caCert []byte) {
	// Write CA cert to a temporary file for the example
	tmpFile := "/tmp/alertmanager-ca.pem"
	if err := os.WriteFile(tmpFile, caCert, 0600); err != nil {
		fmt.Printf("Failed to write temp CA file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile)

	// Test: Verify that correct TLS 1.3 + basic auth succeeds using Args constructor
	fmt.Println("Test: Sending alert with CORRECT TLS 1.3 + basic auth using Args (should succeed)...")

	// NewAlertmanagerWithArgs provides a convenient way to configure
	// the client with a struct - perfect for loading from config files
	// This demonstrates all the security-related fields
	args := alertmanager.Args{
		Enabled:         true,
		AlertmanagerURL: "https://localhost:9094",
		Username:        "admin",
		Password:        "password",
		TLSCACertPath:   tmpFile,
		TLSMinVersion:   "TLS13",         // Enforce TLS 1.3
		Timeout:         5 * time.Second, // Optional: defaults to 2s
		//TLSMaxVersion:         "TLS13",             // Optional: enforce max TLS version
		//TLSInsecureSkipVerify: false,               // Optional: skip TLS verification (not recommended)
		//ProxyURL:              "http://proxy:8080", // Optional: HTTP proxy
	}

	am, err := alertmanager.NewAlertmanagerWithArgs(logger, args)
	if err != nil {
		fmt.Printf("  ✗ Failed to create Alertmanager client: %v\n", err)
		return
	}

	// Client can be nil if Enabled=false
	if am == nil {
		fmt.Println("  ✗ Alertmanager client is disabled (Enabled=false)")
		return
	}

	// Create and send a test alert
	testAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "ArgsConstructorSecureTest"),
		alertmanager.WithLabel("severity", "info"),
		alertmanager.WithLabel("service", "secure-args-demo"),
		alertmanager.WithLabel("environment", "production"),
		alertmanager.WithAnnotation("summary", "Testing Args constructor with security"),
		alertmanager.WithAnnotation("description", "This alert was sent using NewAlertmanagerWithArgs with TLS 1.3 and basic auth"),
	)

	resp, err := am.Emit(testAlert)
	if err != nil {
		fmt.Printf("  ✗ Failed to send alert: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("  ✗ Alertmanager returned non-OK status: %d %s\n", resp.StatusCode, resp.Status)
		return
	}

	fmt.Printf("  ✓ Successfully sent alert (Status: %d)\n", resp.StatusCode)

	fmt.Println("\nArgs constructor benefits:")
	fmt.Println("✓ Struct-based configuration (easy to unmarshal from YAML/JSON)")
	fmt.Println("✓ All security options available (TLS versions, auth, CA certs)")
	fmt.Println("✓ Secure defaults applied automatically (2s timeout)")
	fmt.Println("✓ Built-in enable/disable flag for easy feature toggling")
	fmt.Println("✓ Perfect for loading credentials from secret managers")
}
