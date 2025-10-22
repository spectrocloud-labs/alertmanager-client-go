package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"github.com/go-logr/logr"

	alertmanager "github.com/spectrocloud-labs/alertmanager-client-go"
)

// generateAuditID generates a cryptographically secure random audit ID
func generateAuditID() (string, error) {
	bytes := make([]byte, 16) // 16 bytes = 32 hex characters
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func main() {
	logger := logr.Discard() // Use your preferred logger
	client := &http.Client{Timeout: 30 * time.Second}

	// Create Alertmanager client with base configuration
	am, err := alertmanager.NewAlertmanager(logger, client,
		alertmanager.WithEndpoint("http://localhost:9093"),
		alertmanager.WithBaseLabel("service", "audit-demo"),
		alertmanager.WithBaseLabel("environment", "development"),
	)
	if err != nil {
		fmt.Printf("Failed to create Alertmanager client: %v\n", err)
		return
	}

	// Simulate a series of CRUD operations on a Kubernetes ConfigMap
	resourceKind := "ConfigMap"
	resourceNamespace := "default"
	resourceName := "my-config"
	user := "john@example.com"

	fmt.Println("Simulating CRUD operations on ConfigMap...")

	// Operation 1: Create ConfigMap
	auditID, err := generateAuditID()
	if err != nil {
		fmt.Printf("Failed to generate audit ID: %v\n", err)
		return
	}
	fmt.Printf("[%s] CREATE ConfigMap %s/%s by %s (audit_id: %s)\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user, auditID)
	createAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "create"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("audit_id", auditID),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap created"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s created ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	resp, err := am.Emit(createAlert)
	if err != nil {
		fmt.Printf("Failed to send CREATE alert: %v\n", err)
		return
	}
	resp.Body.Close()

	// Operation 2: Update ConfigMap
	auditID, err = generateAuditID()
	if err != nil {
		fmt.Printf("Failed to generate audit ID: %v\n", err)
		return
	}
	fmt.Printf("[%s] UPDATE ConfigMap %s/%s by %s (audit_id: %s)\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user, auditID)
	updateAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "update"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("audit_id", auditID),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap updated"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s updated ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	resp, err = am.Emit(updateAlert)
	if err != nil {
		fmt.Printf("Failed to send UPDATE alert: %v\n", err)
		return
	}
	resp.Body.Close()

	// Operation 3: Update ConfigMap again
	auditID, err = generateAuditID()
	if err != nil {
		fmt.Printf("Failed to generate audit ID: %v\n", err)
		return
	}
	fmt.Printf("[%s] UPDATE ConfigMap %s/%s by %s (audit_id: %s)\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user, auditID)
	updateAlert2 := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "update"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("audit_id", auditID),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap updated"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s updated ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	resp, err = am.Emit(updateAlert2)
	if err != nil {
		fmt.Printf("Failed to send UPDATE alert: %v\n", err)
		return
	}
	resp.Body.Close()

	// Operation 4: Delete ConfigMap
	auditID, err = generateAuditID()
	if err != nil {
		fmt.Printf("Failed to generate audit ID: %v\n", err)
		return
	}
	fmt.Printf("[%s] DELETE ConfigMap %s/%s by %s (audit_id: %s)\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user, auditID)
	deleteAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "delete"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("audit_id", auditID),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap deleted"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s deleted ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	resp, err = am.Emit(deleteAlert)
	if err != nil {
		fmt.Printf("Failed to send DELETE alert: %v\n", err)
		return
	}
	resp.Body.Close()

	// Operation 5: Re-create ConfigMap
	auditID, err = generateAuditID()
	if err != nil {
		fmt.Printf("Failed to generate audit ID: %v\n", err)
		return
	}
	fmt.Printf("[%s] CREATE ConfigMap %s/%s by %s (audit_id: %s)\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user, auditID)
	recreateAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "create"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("audit_id", auditID),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap created"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s created ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	resp, err = am.Emit(recreateAlert)
	if err != nil {
		fmt.Printf("Failed to send CREATE alert: %v\n", err)
		return
	}
	resp.Body.Close()

	fmt.Println("\nSuccessfully sent 5 audit log alerts to Alertmanager!")
	fmt.Println("\nEach alert has a unique audit_id label to prevent deduplication.")
	fmt.Println("Both CREATE and both UPDATE operations will appear as separate alerts.")
	fmt.Println("Check the Alertmanager web UI at http://localhost:9093")
	fmt.Println("Look for alerts with service=audit-demo - you should see all 5 operations.")
}
