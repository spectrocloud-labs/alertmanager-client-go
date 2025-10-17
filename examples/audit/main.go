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
	fmt.Println()

	// Operation 1: Create ConfigMap
	fmt.Printf("[%s] CREATE ConfigMap %s/%s by %s\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user)
	createAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "create"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("operation_id", fmt.Sprintf("%d", time.Now().UnixNano())),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap created"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s created ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	if err := am.Emit(createAlert); err != nil {
		fmt.Printf("Failed to send CREATE alert: %v\n", err)
		return
	}

	// Operation 2: Update ConfigMap
	fmt.Printf("[%s] UPDATE ConfigMap %s/%s by %s\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user)
	updateAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "update"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("operation_id", fmt.Sprintf("%d", time.Now().UnixNano())),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap updated"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s updated ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	if err := am.Emit(updateAlert); err != nil {
		fmt.Printf("Failed to send UPDATE alert: %v\n", err)
		return
	}

	// Operation 3: Delete ConfigMap
	fmt.Printf("[%s] DELETE ConfigMap %s/%s by %s\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user)
	deleteAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "delete"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("operation_id", fmt.Sprintf("%d", time.Now().UnixNano())),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap deleted"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s deleted ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	if err := am.Emit(deleteAlert); err != nil {
		fmt.Printf("Failed to send DELETE alert: %v\n", err)
		return
	}

	// Operation 4: Re-create ConfigMap
	fmt.Printf("[%s] CREATE ConfigMap %s/%s by %s (recreate)\n", time.Now().Format(time.RFC3339), resourceNamespace, resourceName, user)
	recreateAlert := alertmanager.NewAlert(
		alertmanager.WithLabel("alertname", "AuditLog"),
		alertmanager.WithLabel("operation", "create"),
		alertmanager.WithLabel("resource_kind", resourceKind),
		alertmanager.WithLabel("resource_namespace", resourceNamespace),
		alertmanager.WithLabel("resource_name", resourceName),
		alertmanager.WithLabel("user", user),
		alertmanager.WithLabel("operation_id", fmt.Sprintf("%d", time.Now().UnixNano())),
		alertmanager.WithEndsAt(time.Now().Add(1*time.Hour)),
		alertmanager.WithAnnotation("summary", "ConfigMap created"),
		alertmanager.WithAnnotation("description", fmt.Sprintf("User %s created ConfigMap %s/%s", user, resourceNamespace, resourceName)),
	)
	if err := am.Emit(recreateAlert); err != nil {
		fmt.Printf("Failed to send CREATE alert: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("Successfully sent 4 audit log alerts to Alertmanager!")
	fmt.Println()
	fmt.Println("Key test: Will both CREATE operations appear as separate alerts?")
	fmt.Println("They have identical labels but were sent at different times.")
	fmt.Println()
	fmt.Println("Check the Alertmanager web UI at http://localhost:9093")
	fmt.Println("Look for alerts with service=audit-demo and count how many appear.")
}
