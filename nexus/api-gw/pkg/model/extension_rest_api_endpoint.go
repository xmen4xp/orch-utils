package model

import (
	"sync"
)

// ExtensionRestAPIEndpointSpec holds the parsed ExtensionRestAPIEndpoint CR spec.
type ExtensionRestAPIEndpointSpec struct {
	Name                string // CR name
	ExtensionRestAPIRef string // Reference to ExtensionRestAPI CR name
	Service             string // Fully qualified service DNS name (e.g., "metrics-api.hdai-system.svc.cluster.local")
	Port                string // Service port (e.g., "8080" or "http")
}

var (
	// ExtensionEndpointByAPIRef maps ExtensionRestAPI CR name to its endpoint spec.
	ExtensionEndpointByAPIRef      = make(map[string]ExtensionRestAPIEndpointSpec)
	extensionEndpointByAPIRefMutex = &sync.RWMutex{}
)

// ConstructExtensionRestAPIEndpoint adds or updates an ExtensionRestAPIEndpoint in the cache.
func ConstructExtensionRestAPIEndpoint(eventType EventType, spec ExtensionRestAPIEndpointSpec) {
	extensionEndpointByAPIRefMutex.Lock()
	if eventType == Delete {
		for apiRef, endpoint := range ExtensionEndpointByAPIRef {
			if endpoint.Name == spec.Name {
				delete(ExtensionEndpointByAPIRef, apiRef)
				break
			}
		}
		extensionEndpointByAPIRefMutex.Unlock()
		return
	}

	ExtensionEndpointByAPIRef[spec.ExtensionRestAPIRef] = spec
	extensionEndpointByAPIRefMutex.Unlock()
}

// GetExtensionRestAPIEndpoint retrieves an ExtensionRestAPIEndpointSpec by ExtensionRestAPI CR name.
func GetExtensionRestAPIEndpoint(extensionRestAPIRef string) (ExtensionRestAPIEndpointSpec, bool) {
	extensionEndpointByAPIRefMutex.RLock()
	defer extensionEndpointByAPIRefMutex.RUnlock()

	spec, ok := ExtensionEndpointByAPIRef[extensionRestAPIRef]
	return spec, ok
}

// GetAllExtensionRestAPIEndpoints returns all ExtensionRestAPIEndpointSpecs.
func GetAllExtensionRestAPIEndpoints() map[string]ExtensionRestAPIEndpointSpec {
	extensionEndpointByAPIRefMutex.RLock()
	defer extensionEndpointByAPIRefMutex.RUnlock()

	// Return a copy to avoid race conditions
	result := make(map[string]ExtensionRestAPIEndpointSpec)
	for k, v := range ExtensionEndpointByAPIRef {
		result[k] = v
	}
	return result
}
