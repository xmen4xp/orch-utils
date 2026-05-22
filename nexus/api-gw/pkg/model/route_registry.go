package model

import (
	"fmt"
	"strings"
	"sync"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RouteSource identifies the source of a route registration.
type RouteSource string

const (
	// RouteSourceNexusCRD indicates the route comes from a Nexus datamodel CRD.
	RouteSourceNexusCRD RouteSource = "nexus-crd"
	// RouteSourceExtensionRestAPI indicates the route comes from an ExtensionRestAPI CR.
	RouteSourceExtensionRestAPI RouteSource = "extension-rest-api"
)

// RouteOwner contains information about who owns a registered route.
type RouteOwner struct {
	Source RouteSource                 // Source type of the route
	CRName string                      // Name of the CR that owns this route
	GVR    schema.GroupVersionResource // GVR for status updates
}

// CollisionInfo contains details about a route collision.
type CollisionInfo struct {
	URI               string      `json:"uri"`
	Method            string      `json:"method"`
	ConflictingCR     string      `json:"conflictingCR"`
	ConflictingSource RouteSource `json:"conflictingSource"`
}

// RegisteredRoute represents a successfully registered route.
type RegisteredRoute struct {
	URI    string `json:"uri"`
	Method string `json:"method"`
}

// RouteRegistry tracks all registered URI+Method combinations to detect collisions.
type RouteRegistry struct {
	// routes maps normalized URI -> method -> owner
	routes map[string]map[string]RouteOwner
	mutex  sync.RWMutex
}

// GlobalRouteRegistry is the singleton route registry instance.
var GlobalRouteRegistry = NewRouteRegistry()

// NewRouteRegistry creates a new RouteRegistry.
func NewRouteRegistry() *RouteRegistry {
	return &RouteRegistry{
		routes: make(map[string]map[string]RouteOwner),
	}
}

// normalizeURI normalizes a URI for consistent comparison.
// Converts Echo-style params (:param) to template-style ({param}).
func normalizeURI(uri string) string {
	// Already normalized if using {param} style
	if strings.Contains(uri, "{") {
		return uri
	}
	// Convert :param to {param} for consistency
	result := uri
	parts := strings.Split(uri, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	result = strings.Join(parts, "/")
	return result
}

// Register registers a route with the given URI, method, and owner.
// Returns an error if the route is already registered by a different owner.
func (r *RouteRegistry) Register(uri, method string, owner RouteOwner) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	normalizedURI := normalizeURI(uri)
	upperMethod := strings.ToUpper(method)

	if r.routes[normalizedURI] == nil {
		r.routes[normalizedURI] = make(map[string]RouteOwner)
	}

	if existing, exists := r.routes[normalizedURI][upperMethod]; exists {
		// Allow re-registration by the same owner (idempotent)
		if existing.CRName == owner.CRName && existing.Source == owner.Source {
			return nil
		}
		return fmt.Errorf("route %s %s already registered by %s (%s)",
			upperMethod, normalizedURI, existing.CRName, existing.Source)
	}

	r.routes[normalizedURI][upperMethod] = owner
	return nil
}

// Unregister removes a specific route registration.
func (r *RouteRegistry) Unregister(uri, method string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	normalizedURI := normalizeURI(uri)
	upperMethod := strings.ToUpper(method)

	if methods, exists := r.routes[normalizedURI]; exists {
		delete(methods, upperMethod)
		if len(methods) == 0 {
			delete(r.routes, normalizedURI)
		}
	}
}

// UnregisterByCR removes all routes registered by a specific CR.
func (r *RouteRegistry) UnregisterByCR(crName string, source RouteSource) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	for uri, methods := range r.routes {
		for method, owner := range methods {
			if owner.CRName == crName && owner.Source == source {
				delete(methods, method)
			}
		}
		if len(methods) == 0 {
			delete(r.routes, uri)
		}
	}
}

// CheckCollision checks if any of the given methods for a URI would collide
// with existing registrations. Returns collision info for any conflicts.
// excludeCR allows excluding a specific CR from collision detection (for re-registration).
func (r *RouteRegistry) CheckCollision(uri string, methods []string, excludeCR string, excludeSource RouteSource) []CollisionInfo {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	return r.checkCollisionLocked(uri, methods, excludeCR, excludeSource)
}

// checkCollisionLocked performs collision detection while the caller already holds the lock.
func (r *RouteRegistry) checkCollisionLocked(uri string, methods []string, excludeCR string, excludeSource RouteSource) []CollisionInfo {
	normalizedURI := normalizeURI(uri)
	var collisions []CollisionInfo

	existingMethods, exists := r.routes[normalizedURI]
	if !exists {
		return nil
	}

	for _, method := range methods {
		upperMethod := strings.ToUpper(method)
		if owner, exists := existingMethods[upperMethod]; exists {
			// Skip if this is the same CR (re-registration)
			if owner.CRName == excludeCR && owner.Source == excludeSource {
				continue
			}
			collisions = append(collisions, CollisionInfo{
				URI:               normalizedURI,
				Method:            upperMethod,
				ConflictingCR:     owner.CRName,
				ConflictingSource: owner.Source,
			})
		}
	}

	return collisions
}

// CheckAndRegister atomically checks for collisions and registers routes if none found.
// Returns (registeredRoutes, collisions). If collisions is non-empty, no routes are registered.
func (r *RouteRegistry) CheckAndRegister(uri string, methods []string, owner RouteOwner) ([]RegisteredRoute, []CollisionInfo) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	collisions := r.checkCollisionLocked(uri, methods, owner.CRName, owner.Source)
	if len(collisions) > 0 {
		return nil, collisions
	}

	normalizedURI := normalizeURI(uri)
	var registered []RegisteredRoute
	for _, method := range methods {
		upperMethod := strings.ToUpper(method)
		if r.routes[normalizedURI] == nil {
			r.routes[normalizedURI] = make(map[string]RouteOwner)
		}
		r.routes[normalizedURI][upperMethod] = owner
		registered = append(registered, RegisteredRoute{URI: uri, Method: upperMethod})
	}

	return registered, nil
}

// GetOwner returns the owner of a specific route, if registered.
func (r *RouteRegistry) GetOwner(uri, method string) (RouteOwner, bool) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	normalizedURI := normalizeURI(uri)
	upperMethod := strings.ToUpper(method)

	if methods, exists := r.routes[normalizedURI]; exists {
		if owner, exists := methods[upperMethod]; exists {
			return owner, true
		}
	}
	return RouteOwner{}, false
}

// GetAllRoutes returns a copy of all registered routes (for debugging/status).
func (r *RouteRegistry) GetAllRoutes() map[string]map[string]RouteOwner {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	result := make(map[string]map[string]RouteOwner)
	for uri, methods := range r.routes {
		result[uri] = make(map[string]RouteOwner)
		for method, owner := range methods {
			result[uri][method] = owner
		}
	}
	return result
}

// Clear removes all registered routes (useful for testing or server restart).
func (r *RouteRegistry) Clear() {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	r.routes = make(map[string]map[string]RouteOwner)
}
