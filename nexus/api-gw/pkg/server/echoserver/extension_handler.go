package echoserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"nexus-api-gw/pkg/model"

	"github.com/labstack/echo/v4"
)

const (
	// NexusHierarchyHeader is the header name for passing hierarchy context to backend.
	NexusHierarchyHeader = "X-Nexus-Hierarchy"

	// extensionProxyTimeout is the timeout for proxying requests to backend services.
	extensionProxyTimeout = 30 * time.Second
)

// proxyExtensionRequest handles requests to extension REST APIs by proxying to the backend service.
// It uses pre-resolved labels from the unified Nexus handler for hierarchy header construction.
func (s *EchoServer) proxyExtensionRequest(nc *NexusContext, spec model.ExtensionRestAPISpec, labels map[string]string) error {
	// Resolve backend endpoint for this ExtensionRestAPI
	endpoint, err := resolveBackendEndpoint(spec)
	if err != nil {
		log.Warn().Msgf("Backend not available for %s: %v", spec.URI, err)
		return nc.JSON(http.StatusServiceUnavailable, DefaultResponse{
			Message: "Backend service not available",
		})
	}

	// Build the backend URL
	backendURL, err := buildBackendURL(spec.URI, endpoint, nc)
	if err != nil {
		log.Error().Msgf("Failed to build backend URL: %v", err)
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Failed to build backend URL"})
	}

	log.Debug().Msgf("Proxying extension request to: %s", backendURL)

	// Read the request body
	body, err := io.ReadAll(nc.Request().Body)
	if err != nil {
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Failed to read request body"})
	}
	nc.Request().Body = io.NopCloser(bytes.NewBuffer(body))

	// Create the proxy request
	ctx, cancel := context.WithTimeout(context.Background(), extensionProxyTimeout)
	defer cancel()

	proxyReq, err := http.NewRequestWithContext(ctx, nc.Request().Method, backendURL, bytes.NewBuffer(body))
	if err != nil {
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Failed to create proxy request"})
	}

	// Copy headers from original request
	for key, values := range nc.Request().Header {
		for _, value := range values {
			proxyReq.Header.Add(key, value)
		}
	}

	// Add X-Nexus-Hierarchy header from pre-resolved labels
	if len(labels) > 0 {
		hierarchyHeader, err := buildHierarchyHeaderFromLabels(labels)
		if err != nil {
			log.Warn().Msgf("Failed to build hierarchy header: %v", err)
		} else if hierarchyHeader != "" {
			proxyReq.Header.Set(NexusHierarchyHeader, hierarchyHeader)
			log.Debug().Msgf("Added %s header: %s", NexusHierarchyHeader, hierarchyHeader)
		}
	}

	// Send the proxy request
	httpClient := &http.Client{Timeout: extensionProxyTimeout}
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		log.Error().Msgf("Failed to proxy request: %v", err)
		return nc.JSON(http.StatusBadGateway, DefaultResponse{Message: "Backend service unavailable"})
	}
	defer resp.Body.Close()

	// Read the response body
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nc.JSON(http.StatusInternalServerError, DefaultResponse{Message: "Failed to read backend response"})
	}

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			nc.Response().Header().Add(key, value)
		}
	}

	// Return the response
	return nc.Blob(resp.StatusCode, resp.Header.Get("Content-Type"), respBody)
}

// resolveBackendEndpoint resolves the backend endpoint for an ExtensionRestAPI.
// Returns the endpoint spec if found, or an error if no endpoint is configured.
func resolveBackendEndpoint(spec model.ExtensionRestAPISpec) (model.ExtensionRestAPIEndpointSpec, error) {
	endpoint, found := model.GetExtensionRestAPIEndpoint(spec.Name)
	if found {
		log.Debug().Msgf("Using ExtensionRestAPIEndpoint '%s' for backend resolution", endpoint.Name)
		return endpoint, nil
	}

	return model.ExtensionRestAPIEndpointSpec{}, fmt.Errorf("backend service not configured for endpoint '%s'", spec.URI)
}

// buildBackendURL constructs the full backend URL from the URI, endpoint spec, and request.
func buildBackendURL(uri string, endpoint model.ExtensionRestAPIEndpointSpec, c echo.Context) (string, error) {
	// Service is the fully qualified DNS name (e.g., "metrics-api.hdai-system.svc.cluster.local")
	service := endpoint.Service
	if service == "" {
		return "", fmt.Errorf("endpoint '%s' has no service configured", endpoint.Name)
	}

	port := endpoint.Port
	if port == "" {
		port = "80"
	}

	// Build the URI with path parameters substituted
	resolvedURI := substitutePathParams(uri, c)

	// Build the full URL
	baseURL := fmt.Sprintf("http://%s:%s%s", service, port, resolvedURI)

	// Append query parameters from the original request
	if c.QueryString() != "" {
		baseURL = baseURL + "?" + c.QueryString()
	}

	return baseURL, nil
}

// substitutePathParams replaces path parameter placeholders with actual values.
func substitutePathParams(uriTemplate string, c echo.Context) string {
	result := uriTemplate
	for _, paramName := range c.ParamNames() {
		placeholder := "{" + paramName + "}"
		result = strings.ReplaceAll(result, placeholder, c.Param(paramName))
	}
	return result
}

// buildHierarchyHeaderFromLabels constructs the X-Nexus-Hierarchy JSON header
// from pre-resolved labels. It converts CRD type keys to node names for the header.
func buildHierarchyHeaderFromLabels(labels map[string]string) (string, error) {
	hierarchy := make(map[string]string)
	for crdType, value := range labels {
		if nodeInfo, ok := model.CrdTypeToNodeInfo[crdType]; ok {
			hierarchy[nodeInfo.Name] = value
		}
	}
	if len(hierarchy) == 0 {
		return "", nil
	}
	jsonBytes, err := json.Marshal(hierarchy)
	if err != nil {
		return "", err
	}
	return string(jsonBytes), nil
}
