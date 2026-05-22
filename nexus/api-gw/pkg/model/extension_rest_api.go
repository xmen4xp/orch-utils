package model

import (
	"sync"

	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
)

// ExtensionRestAPISpec holds the parsed ExtensionRestAPI CR spec.
type ExtensionRestAPISpec struct {
	Name            string   // CR name (e.g., "datacenters-datacentermetricsapi")
	URI             string   // URI path (e.g., "/v1/datacenters/{datacenters.DataCenters}/metrics")
	Methods         []string // HTTP methods to proxy (e.g., ["GET", "PUT"]). Empty = all methods.
	OpenAPIPathSpec string   // Raw OpenAPI path spec YAML
	Hierarchy       []string // Parent hierarchy node names (e.g., ["config.Config", "gns.Gns"])
	AssociatedNode  string   // Associated Nexus node (e.g., "datacenters.DataCenters")
	Datamodel       string   // Datamodel name for OpenAPI grouping
}

// ExtensionRestAPIAnnotation holds the annotation structure from ExtensionRestAPI CRs.
type ExtensionRestAPIAnnotation struct {
	Name            string            `json:"name,omitempty"`
	Hierarchy       []string          `json:"hierarchy,omitempty"`
	NexusRestAPIGen nexus.RestAPISpec `json:"nexus-rest-api-gen,omitempty"`
}

// defaultExtensionMethods are used when spec.Methods is empty.
var defaultExtensionMethods = []string{"GET", "PUT", "PATCH", "DELETE"}

var (
	// ExtensionURIToSpec maps "METHOD:URI" to ExtensionRestAPISpec.
	// e.g. "GET:/v1/projects/{projects.Project}/metrics" -> spec
	ExtensionURIToSpec      = make(map[string]ExtensionRestAPISpec)
	extensionURIToSpecMutex = &sync.RWMutex{}

	// ExtensionURIChan is used to notify route registration of new extension APIs.
	ExtensionURIChan = make(chan ExtensionRestAPISpec, constDefaultChanSize)

	// ExtensionAPIDeleteChan is used to notify deletion of extension APIs.
	ExtensionAPIDeleteChan = make(chan string, constDefaultChanSize)
)

// extensionRouteKey builds the composite cache key from method and URI.
func extensionRouteKey(method, uri string) string {
	return method + ":" + uri
}

// ConstructExtensionRestAPI adds or updates an ExtensionRestAPI in the cache.
func ConstructExtensionRestAPI(eventType EventType, spec ExtensionRestAPISpec) {
	extensionURIToSpecMutex.Lock()
	if eventType == Delete {
		// Delete all method entries belonging to this CR
		var deletedURI string
		for key, s := range ExtensionURIToSpec {
			if s.Name == spec.Name {
				if deletedURI == "" {
					deletedURI = s.URI
				}
				delete(ExtensionURIToSpec, key)
			}
		}
		extensionURIToSpecMutex.Unlock()
		// Push to chan - MUST be outside mutex to avoid deadlock
		if deletedURI != "" {
			ExtensionAPIDeleteChan <- deletedURI
		}
		return
	}

	methods := spec.Methods
	if len(methods) == 0 {
		methods = defaultExtensionMethods
	}
	for _, method := range methods {
		ExtensionURIToSpec[extensionRouteKey(method, spec.URI)] = spec
	}
	extensionURIToSpecMutex.Unlock()
	// Push to chan - MUST be outside mutex to avoid deadlock
	ExtensionURIChan <- spec
}

// GetExtensionRestAPISpec retrieves an ExtensionRestAPISpec by URI and HTTP method.
func GetExtensionRestAPISpec(uri, method string) (ExtensionRestAPISpec, bool) {
	extensionURIToSpecMutex.RLock()
	defer extensionURIToSpecMutex.RUnlock()

	spec, ok := ExtensionURIToSpec[extensionRouteKey(method, uri)]
	return spec, ok
}

// GetAllExtensionRestAPISpecs returns unique ExtensionRestAPISpecs deduplicated by CR name.
func GetAllExtensionRestAPISpecs() map[string]ExtensionRestAPISpec {
	extensionURIToSpecMutex.RLock()
	defer extensionURIToSpecMutex.RUnlock()

	// Deduplicate by CR name since each CR creates multiple (method:uri) entries
	result := make(map[string]ExtensionRestAPISpec)
	for _, spec := range ExtensionURIToSpec {
		result[spec.Name] = spec
	}
	return result
}
