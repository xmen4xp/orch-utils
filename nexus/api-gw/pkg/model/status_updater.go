package model

import (
	"context"
	"encoding/json"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
)

// RouteStatusPhase represents the registration status phase.
type RouteStatusPhase string

const (
	// RouteStatusPhaseRegistered indicates routes were successfully registered.
	RouteStatusPhaseRegistered RouteStatusPhase = "Registered"
	// RouteStatusPhaseRejected indicates routes were rejected due to collision.
	RouteStatusPhaseRejected RouteStatusPhase = "Rejected"
	// RouteStatusPhasePending indicates routes are pending registration.
	RouteStatusPhasePending RouteStatusPhase = "Pending"
)

// RouteStatus represents the status of route registration for a CR.
type RouteStatus struct {
	Phase            RouteStatusPhase  `json:"phase"`
	Message          string            `json:"message"`
	RegisteredRoutes []RegisteredRoute `json:"registeredRoutes,omitempty"`
	Collisions       []CollisionInfo   `json:"collisions,omitempty"`
	LastUpdated      string            `json:"lastUpdated"`
}

// NewRegisteredStatus creates a RouteStatus for successful registration.
func NewRegisteredStatus(routes []RegisteredRoute) RouteStatus {
	return RouteStatus{
		Phase:            RouteStatusPhaseRegistered,
		Message:          "All routes successfully registered",
		RegisteredRoutes: routes,
		LastUpdated:      time.Now().UTC().Format(time.RFC3339),
	}
}

// NewRejectedStatus creates a RouteStatus for rejected registration due to collisions.
func NewRejectedStatus(collisions []CollisionInfo) RouteStatus {
	return RouteStatus{
		Phase:       RouteStatusPhaseRejected,
		Message:     "Route registration rejected due to collisions",
		Collisions:  collisions,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewPendingStatus creates a RouteStatus for pending registration.
func NewPendingStatus() RouteStatus {
	return RouteStatus{
		Phase:       RouteStatusPhasePending,
		Message:     "Route registration pending",
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}
}

// UpdateCRStatus updates the status subresource of a CR using the dynamic client.
func UpdateCRStatus(ctx context.Context, dynamicClient dynamic.Interface,
	gvr schema.GroupVersionResource, name string, status RouteStatus) error {

	// Build the status patch
	statusMap := map[string]interface{}{
		"phase":       string(status.Phase),
		"message":     status.Message,
		"lastUpdated": status.LastUpdated,
	}

	if len(status.RegisteredRoutes) > 0 {
		routes := make([]interface{}, len(status.RegisteredRoutes))
		for i, r := range status.RegisteredRoutes {
			routes[i] = map[string]interface{}{
				"uri":    r.URI,
				"method": r.Method,
			}
		}
		statusMap["registeredRoutes"] = routes
	}

	if len(status.Collisions) > 0 {
		collisions := make([]interface{}, len(status.Collisions))
		for i, c := range status.Collisions {
			collisions[i] = map[string]interface{}{
				"uri":               c.URI,
				"method":            c.Method,
				"conflictingCR":     c.ConflictingCR,
				"conflictingSource": string(c.ConflictingSource),
			}
		}
		statusMap["collisions"] = collisions
	}

	patch := map[string]interface{}{
		"status": statusMap,
	}

	patchBytes, err := json.Marshal(patch)
	if err != nil {
		return err
	}

	_, err = dynamicClient.Resource(gvr).Patch(ctx, name, types.MergePatchType, patchBytes, metav1.PatchOptions{}, "status")
	return err
}

// UpdateCRStatusWithRouteStatus is a convenience wrapper that updates status for ExtensionRestAPI CRs.
func UpdateCRStatusWithRouteStatus(ctx context.Context, dynamicClient dynamic.Interface,
	name string, status RouteStatus) error {

	gvr := schema.GroupVersionResource{
		Group:    "nexus.com",
		Version:  "v1",
		Resource: "extensionrestapis",
	}

	return UpdateCRStatus(ctx, dynamicClient, gvr, name, status)
}
