// Copyright (C) 2025 Intel Corporation
// SPDX-FileCopyrightText: 2025 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package echoserver

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"
	"sync"
	"time"

	"nexus-api-gw/pkg/client"
	"nexus-api-gw/pkg/common"
	"nexus-api-gw/pkg/config"
	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/openapi/api"
	"nexus-api-gw/pkg/openapi/declarative"
	"nexus-api-gw/pkg/utils"

	"github.com/fsnotify/fsnotify"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/vmware-tanzu/graph-framework-for-microservices/nexus/nexus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	corev1client "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/tools/record"

	nexusClient "nexus/admin/api/build/nexus-client"

	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/auditing"
	"github.com/open-edge-platform/infra-core/inventory/v2/pkg/logging"
)

const (
	stopServerTimerTimeout = 30 * time.Second
	stopServerCtxTimeout   = 10 * time.Second
	stopServerWaitTime     = 100 * time.Millisecond
	tmReconcileTime        = 600 * time.Second
)

var (
	appName = "nexus-api-gw-echoserver"
	log     = logging.GetLogger(appName)
)

type TenantData struct {
	TenantName string `json:"tenantName" form:"tenantName"`
	Sku        string `json:"sku,omitempty" form:"sku,omitempty"`
}

type UserLogin struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

var (
	TotalHTTPServerRestartCounter                 = 0
	HTTPServerRestartFromOpenAPISpecUpdateCounter = 0
)

type EchoServer struct {
	Echo        *echo.Echo
	Config      *config.Config
	Client      KubernetesClient
	NexusClient *nexusClient.Clientset
	k8sProxy    *httputil.ReverseProxy
	mu          sync.Mutex
	restartMu   sync.Mutex // Mutex to protect restart counter
	recorder    record.EventRecorder
}

type KubernetesClient interface {
	CoreV1() corev1client.CoreV1Interface
}

func InitEcho(stopCh chan struct{}, conf *config.Config, client KubernetesClient, nc *nexusClient.Clientset) *EchoServer {
	log.Info().Msg("Init Echo")
	e := NewEchoServer(conf, client, nc)

	// Initialize EventRecorder
	eventBroadcaster := record.NewBroadcaster()
	eventBroadcaster.StartRecordingToSink(&corev1client.EventSinkImpl{Interface: client.CoreV1().Events("")})
	// For cluster-scoped objects like CRDs, we need an EventSource
	e.recorder = eventBroadcaster.NewRecorder(clientgoscheme.Scheme, corev1.EventSource{Component: "nexus-api-gw"})

	if conf.EnableNexusRuntime {
		e.RegisterNexusRoutes()
		e.ReplayNexusCRDRoutes()
		e.ReplayExtensionRoutes()
	}

	if conf.BackendService != "" {
		if err := declarative.Setup(declarative.OpenAPISpecFile); err != nil {
			log.Fatal().Msgf("unable to complete setup, %s", err.Error())
		}

		e.RegisterDeclarativeRoutes()
		e.RegisterDeclarativeRouter()
	}

	e.RegisterDebug()
	e.Start(stopCh)

	if config.Cfg.BackendService != "" {
		WatchForOpenAPISpecChanges(stopCh, declarative.OpenAPISpecDir, declarative.OpenAPISpecFile, e)
	}
	return e
}

func (s *EchoServer) StartHTTPServer() {
	port := "80"
	if s.Config.Server.HTTPPort != "" {
		port = s.Config.Server.HTTPPort
	}

	if err := s.Echo.Start(fmt.Sprintf(":%s", port)); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Msgf("Server error %v", err)
	}
}

func (s *EchoServer) Start(stopCh chan struct{}) {
	go func() {
		s.mu.Lock()         // Lock the mutex before starting the server
		defer s.mu.Unlock() // Ensure the mutex is unlocked when the function returns

		if s.Config.EnableNexusRuntime {
			// Start watching URI notification
			go func() {
				log.Debug().Msg("NodeUpdateNotifications.. restarting server")
				if err := s.NodeUpdateNotifications(stopCh); err != nil {
					s.mu.Lock()
					s.StopServer()
					InitEcho(stopCh, s.Config, s.Client, s.NexusClient)
					s.restartMu.Lock()
					TotalHTTPServerRestartCounter++
					log.Info().Msgf("TotalHTTPServerRestartCounter: %d", TotalHTTPServerRestartCounter)
					s.restartMu.Unlock()
					s.mu.Unlock()
				}
			}()
		}

		// Start Server
		go func() {
			log.Info().Msg("Start Echo Server")
			if utils.IsServerConfigValid(s.Config) &&
				utils.IsFileExists(s.Config.Server.CertPath) &&
				utils.IsFileExists(s.Config.Server.KeyPath) {
				log.Info().Msgf("Server Config %v", s.Config.Server)
				log.Info().Msg("Start TLS Server")
				err := s.Echo.StartTLS(s.Config.Server.Address, s.Config.Server.CertPath, s.Config.Server.KeyPath)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Fatal().Msgf("TLS Server error %v", err)
				}
			} else {
				log.Info().Msg("Certificates or TLS port not configured correctly, hence starting the HTTP Server")
				s.StartHTTPServer()
			}
		}()
	}()
}

type NexusContext struct {
	echo.Context
	NexusURI string
	Codes    nexus.HTTPCodesResponse

	// Kube
	CrdType   string
	GroupName string
	Resource  string
}

func (s *EchoServer) RegisterDebug() {
	s.Echo.GET("/debug/all", DebugAllHandler)
}

func (s *EchoServer) RegisterNexusRoutes() {
	// OpenAPI route
	s.Echo.GET("/:datamodel/openapi.json", func(c echo.Context) error {
		return c.JSON(http.StatusOK, api.Schemas[c.Param("datamodel")])
	})

	// Swagger-UI, datamodel is edge-orchestrator.intel.com
	s.Echo.GET("/:datamodel/docs", SwaggerUI)
}

// ReplayNexusCRDRoutes re-registers all known Nexus CRD-derived routes on the current echo server.
// Echo server restarts (triggered by CRD reconcilers via StopCh) destroy all registered routes,
// and waiting for the controller to re-push them via RestURIChan is racy: another restart can
// arrive before the controller drains the channel. Replaying from the cached CrdTypeToRestUris map
// at InitEcho time guarantees every datamodel route is present on the new server.
//
// The route registry is reset before replay so that "already-registered" errors from the previous
// server's bookkeeping do not skip routes on the new server.
func (s *EchoServer) ReplayNexusCRDRoutes() {
	count := 0
	for crdType, uris := range model.CrdTypeToRestUris {
		// Drop the previous server's bookkeeping for this CRD so that the new
		// server's Register() calls don't see "already registered" collisions.
		model.GlobalRouteRegistry.UnregisterByCR(crdType, model.RouteSourceNexusCRD)
		for _, u := range uris {
			if httpCodesResponse, ok := u.Methods[http.MethodPut]; ok {
				u.Methods[http.MethodPatch] = httpCodesResponse
			}
			registered, _ := s.RegisterRouter(u, crdType)
			count += len(registered)
		}
		s.RegisterCrdRouter(crdType)
	}
	if count > 0 {
		log.Info().Msgf("Replaying %d Nexus CRD routes after server restart", count)
	}
}

// ReplayExtensionRoutes re-registers all cached ExtensionRestAPI routes on the current echo server.
// This is needed because echo server restarts (triggered by CRD reconcilers via StopCh) destroy all
// registered routes.
func (s *EchoServer) ReplayExtensionRoutes() {
	specs := model.GetAllExtensionRestAPISpecs()
	if len(specs) == 0 {
		return
	}
	log.Info().Msgf("Replaying %d extension REST API routes after server restart", len(specs))
	for _, spec := range specs {
		s.RegisterExtensionRouter(spec)
	}
}

func (s *EchoServer) RegisterDeclarativeRoutes() {
	s.Echo.GET("/declarative/apis", declarative.ApisHandler)
}

func (s *EchoServer) RegisterRouter(restURI nexus.RestURIs, crdType string) ([]model.RegisteredRoute, []model.CollisionInfo) {
	urlPattern := model.ConstructEchoPathParamURL(restURI.Uri)

	var registeredRoutes []model.RegisteredRoute
	var collisions []model.CollisionInfo

	for method, codes := range restURI.Methods {
		methodStr := string(method)

		// Check for collision before registering
		owner := model.RouteOwner{
			Source: model.RouteSourceNexusCRD,
			CRName: crdType,
		}
		if err := model.GlobalRouteRegistry.Register(restURI.Uri, methodStr, owner); err != nil {
			log.Warn().Msgf("Skipping route %s %s for Nexus CRD %s: %v", methodStr, urlPattern, crdType, err)
			// Record collision
			existingOwner, _ := model.GlobalRouteRegistry.GetOwner(restURI.Uri, methodStr)
			collisions = append(collisions, model.CollisionInfo{
				URI:               restURI.Uri,
				Method:            methodStr,
				ConflictingCR:     existingOwner.CRName,
				ConflictingSource: existingOwner.Source,
			})
			continue
		}

		log.Info().Msgf("Registered Router Path %s Method %s\n", urlPattern, methodStr)
		nexusContext := s.GetNexusContext(restURI, codes)
		s.registerRoute(methodStr, urlPattern, nexusContext)

		registeredRoutes = append(registeredRoutes, model.RegisteredRoute{
			URI:    restURI.Uri,
			Method: methodStr,
		})
	}

	return registeredRoutes, collisions
}

func (s *EchoServer) registerRoute(method, urlPattern string, nexusContext func(next echo.HandlerFunc) echo.HandlerFunc) {
	switch method {
	case "LIST":
		s.registerListRoute(urlPattern, nexusContext)
	case http.MethodGet:
		s.registerGetRoute(urlPattern, nexusContext)
	case http.MethodPut:
		s.registerPutRoute(urlPattern, nexusContext)
	case http.MethodPatch:
		s.registerPatchRoute(urlPattern, nexusContext)
	case http.MethodDelete:
		s.registerDeleteRoute(urlPattern, nexusContext)
	}
}

func (s *EchoServer) registerListRoute(urlPattern string, nexusContext func(next echo.HandlerFunc) echo.HandlerFunc) {
	if common.IsModeAdmin() || common.IsTenancyMode() {
		s.Echo.GET(urlPattern, s.ListHandler, nexusContext)
	}
}

func (s *EchoServer) registerGetRoute(urlPattern string, nexusContext func(next echo.HandlerFunc) echo.HandlerFunc) {
	if common.IsModeAdmin() || common.IsTenancyMode() {
		s.Echo.GET(urlPattern, s.GetHandler, nexusContext)
	}
}

func (s *EchoServer) registerPutRoute(urlPattern string, nexusContext func(next echo.HandlerFunc) echo.HandlerFunc) {
	if common.IsModeAdmin() || common.IsTenancyMode() {
		s.Echo.PUT(urlPattern, s.PutHandler, nexusContext)
	}
}

func (s *EchoServer) registerPatchRoute(urlPattern string, nexusContext func(next echo.HandlerFunc) echo.HandlerFunc) {
	if common.IsModeAdmin() || common.IsTenancyMode() {
		s.Echo.PATCH(urlPattern, s.PatchHandler, nexusContext)
	}
}

func (s *EchoServer) registerDeleteRoute(urlPattern string, nexusContext func(next echo.HandlerFunc) echo.HandlerFunc) {
	if common.IsModeAdmin() || common.IsTenancyMode() {
		s.Echo.DELETE(urlPattern, s.deleteHandler, nexusContext)
	}
}

func (s *EchoServer) GetNexusContext(restURI nexus.RestURIs,
	codes nexus.HTTPCodesResponse,
) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			nc := &NexusContext{
				Context:  c,
				NexusURI: restURI.Uri,
				Codes:    codes,
			}
			return next(nc)
		}
	}
}

func (s *EchoServer) GetNexusCrdContext(crdType, groupName, resource string) func(next echo.HandlerFunc) echo.HandlerFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			nc := &NexusContext{
				Context:   c,
				CrdType:   crdType,
				GroupName: groupName,
				Resource:  resource,
			}
			return next(nc)
		}
	}
}

func (s *EchoServer) RegisterCrdRouter(crdType string) {
	crdParts := strings.Split(crdType, ".")
	groupName := strings.Join(crdParts[1:], ".")
	resourcePattern := fmt.Sprintf("/apis/%s/v1/%s", groupName, crdParts[0])
	resourceNamePattern := resourcePattern + "/:name"
	crdContext := s.GetNexusCrdContext(crdType, groupName, crdParts[0])

	// TODO NPT-313 support authentication for kubectl proxy requests
	s.Echo.GET(resourceNamePattern, KubeGetByNameHandler, crdContext)
	s.Echo.GET(resourcePattern, KubeGetHandler, crdContext)
	s.Echo.POST(resourcePattern, KubePostHandler, crdContext)
	s.Echo.DELETE(resourceNamePattern, KubeDeleteHandler, crdContext)
}

// RegisterExtensionRouter registers routes for an ExtensionRestAPI.
// Extension API routes are handled by the unified Nexus handlers (GetHandler, PutHandler, etc.)
// which detect extension APIs and proxy them to the backend after hierarchy resolution.
// Note: Routes should already be registered in GlobalRouteRegistry by the controller.
func (s *EchoServer) RegisterExtensionRouter(spec model.ExtensionRestAPISpec) {
	urlPattern := model.ConstructEchoPathParamURL(spec.URI)

	// Populate URIToCRDType so getCRDInfoAndName works in the unified handler
	model.SetURIToCRDType(spec.URI, spec.AssociatedNode)

	// Use Methods field if specified, otherwise default to all methods
	methods := spec.Methods
	if len(methods) == 0 {
		// Default: register all common HTTP methods
		methods = []string{http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodDelete}
	}

	nexusContext := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			nc := &NexusContext{
				Context:  c,
				NexusURI: spec.URI,
			}
			return next(nc)
		}
	}

	for _, method := range methods {
		// Verify this route is owned by this ExtensionRestAPI in the registry
		owner, exists := model.GlobalRouteRegistry.GetOwner(spec.URI, method)
		if !exists || owner.CRName != spec.Name || owner.Source != model.RouteSourceExtensionRestAPI {
			log.Debug().Msgf("Skipping extension route %s %s - not owned by %s", method, urlPattern, spec.Name)
			continue
		}

		log.Info().Msgf("Registered Extension Router Path %s Method %s", urlPattern, method)
		s.registerRoute(method, urlPattern, nexusContext)
	}
}

func (s *EchoServer) RegisterDeclarativeRouter() {
	for uri, path := range declarative.Paths {
		if path.Get != nil {
			endpointContext := declarative.SetupContext(uri, http.MethodGet, path.Get)

			if endpointContext.Single {
				s.Echo.GET(endpointContext.URI, declarative.GetHandler, declarative.Middleware(endpointContext, true))
				if endpointContext.ShortURI != "" {
					s.Echo.GET(endpointContext.ShortURI, declarative.GetHandler, declarative.Middleware(endpointContext, true))
					log.Debug().Msgf("Registered declarative short get endpoint: %s for uri: %s", endpointContext.ShortURI, uri)
				}

				declarative.AddApisEndpoint(endpointContext)
				log.Debug().Msgf("Registered declarative get endpoint: %s for uri: %s", endpointContext.URI, uri)
			} else {
				s.Echo.GET(endpointContext.URI, declarative.ListHandler, declarative.Middleware(endpointContext, false))
				if endpointContext.ShortURI != "" {
					s.Echo.GET(endpointContext.ShortURI, declarative.ListHandler, declarative.Middleware(endpointContext, false))
					log.Debug().Msgf("Registered declarative short list endpoint: %s for uri: %s", endpointContext.ShortURI, uri)
				}

				declarative.AddApisEndpoint(endpointContext)
				log.Debug().Msgf("Registered declarative list endpoint: %s for uri: %s", endpointContext.URI, uri)
			}
		}

		if path.Put != nil {
			endpointContext := declarative.SetupContext(uri, http.MethodPut, path.Put)
			s.Echo.PUT(endpointContext.URI, declarative.PutHandler, declarative.Middleware(endpointContext, false))
			if endpointContext.ShortURI != "" {
				s.Echo.PUT(endpointContext.ShortURI, declarative.PutHandler, declarative.Middleware(endpointContext, false))
				log.Debug().Msgf("Registered declarative short put endpoint: %s for uri: %s", endpointContext.ShortURI, uri)
			}

			declarative.AddApisEndpoint(endpointContext)
			log.Debug().Msgf("Registered declarative put endpoint: %s for uri: %s", endpointContext.URI, uri)
		}

		if path.Delete != nil {
			endpointContext := declarative.SetupContext(uri, http.MethodDelete, path.Delete)
			s.Echo.DELETE(endpointContext.URI, declarative.DeleteHandler, declarative.Middleware(endpointContext, true))
			if endpointContext.ShortURI != "" {
				s.Echo.DELETE(endpointContext.ShortURI, declarative.DeleteHandler, declarative.Middleware(endpointContext, true))
				log.Debug().Msgf("Registered declarative short delete endpoint: %s for uri: %s", endpointContext.ShortURI, uri)
			}

			declarative.AddApisEndpoint(endpointContext)
			log.Debug().Msgf("Registered declarative delete endpoint: %s for uri: %s", endpointContext.URI, uri)
		}
	}
}

func (s *EchoServer) NodeUpdateNotifications(stopCh chan struct{}) error {
	for {
		select {
		case <-stopCh:
			return fmt.Errorf("stop signal received")
		case nexusCRDURIs := <-model.RestURIChan:
			log.Debug().Msgf("Rest route notification received for CRD: %s", nexusCRDURIs.CRDType)

			var allRegisteredRoutes []model.RegisteredRoute
			var allCollisions []model.CollisionInfo

			for _, v := range nexusCRDURIs.RestURIs {
				if httpCodesResponse, ok := v.Methods[http.MethodPut]; ok {
					v.Methods[http.MethodPatch] = httpCodesResponse
				}
				registered, collisions := s.RegisterRouter(v, nexusCRDURIs.CRDType)
				allRegisteredRoutes = append(allRegisteredRoutes, registered...)
				allCollisions = append(allCollisions, collisions...)
			}

			// Update Nexus CRD status
			s.updateNexusCRDStatus(nexusCRDURIs.CRDType, allRegisteredRoutes, allCollisions)

			// Construct openapi spec.
			api.Recreate()

		case crdType := <-model.CrdTypeChan:
			log.Debug().Msg("CRD route notification received")
			s.RegisterCrdRouter(crdType)

		case extSpec := <-model.ExtensionURIChan:
			log.Debug().Msgf("Extension REST API notification received: %s", extSpec.URI)
			s.RegisterExtensionRouter(extSpec)

			// Recreate openapi spec to include extension APIs.
			api.Recreate()
			api.RecreateExtension()

		case extURI := <-model.ExtensionAPIDeleteChan:
			log.Debug().Msgf("Extension REST API delete notification received: %s", extURI)
			api.Recreate()
			api.RecreateExtension()
			// Trigger server restart to remove the stale Echo route handler.
			// On restart, ReplayExtensionRoutes re-registers surviving routes from cache.
			// Must send from a separate goroutine since stopCh is unbuffered and
			// this goroutine is also the reader.
			go func() { stopCh <- struct{}{} }()
		}
	}
}

// updateNexusCRDStatus updates the status of a Nexus CRD with route registration results.
func (s *EchoServer) updateNexusCRDStatus(crdType string, registeredRoutes []model.RegisteredRoute, collisions []model.CollisionInfo) {
	if client.Client == nil {
		log.Warn().Msg("Dynamic client not available, skipping Nexus CRD status update")
		return
	}

	// Parse CRD type to get GVR (e.g., "datacenterses.datacenters.hd.cisco.com")
	parts := strings.Split(crdType, ".")
	if len(parts) < 2 {
		log.Warn().Msgf("Invalid CRD type format: %s", crdType)
		return
	}

	// We need to find the CR instances of this CRD type and update their status
	// For now, we'll update the CRD itself (not individual CRs) - this is informational
	// The status will be stored in a way that indicates route registration status

	var status model.RouteStatus
	var eventMessage strings.Builder

	if len(collisions) > 0 {
		status = model.RouteStatus{
			Phase:            model.RouteStatusPhaseRejected,
			RegisteredRoutes: registeredRoutes,
			Collisions:       collisions,
			LastUpdated:      time.Now().UTC().Format(time.RFC3339),
		}

		eventMessage.WriteString(fmt.Sprintf("Routes rejected due to collisions (%d registered, %d collisions).\n", len(registeredRoutes), len(collisions)))
		eventMessage.WriteString("Collisions:\n")
		for _, c := range collisions {
			eventMessage.WriteString(fmt.Sprintf("- %s %s (conflicts with %s %s)\n", c.Method, c.URI, c.ConflictingSource, c.ConflictingCR))
		}
		status.Message = eventMessage.String()
	} else if len(registeredRoutes) > 0 {
		status = model.NewRegisteredStatus(registeredRoutes)

		eventMessage.WriteString(fmt.Sprintf("All %d routes successfully registered.\n", len(registeredRoutes)))
		eventMessage.WriteString("Registered:\n")
		for _, r := range registeredRoutes {
			eventMessage.WriteString(fmt.Sprintf("- %s %s\n", r.Method, r.URI))
		}
		status.Message = eventMessage.String()
	} else {
		// No routes to register
		return
	}

	// Update status on the CRD's status subresource
	// Note: This updates the CRD definition's status, not individual CR instances
	// For individual CR status updates, we'd need to iterate over all CRs of this type
	log.Info().Msgf("Nexus CRD %s route registration: %s (%d routes, %d collisions)",
		crdType, status.Phase, len(registeredRoutes), len(collisions))

	if s.recorder != nil {
		crdObj := &corev1.ObjectReference{
			Kind:       "CustomResourceDefinition",
			APIVersion: "apiextensions.k8s.io/v1",
			Name:       crdType,
			Namespace:  "", // Explicitly empty for cluster-scoped objects
		}

		// Try to fetch the CRD to get its UID for better event association
		if client.Client != nil {
			crdGVR := schema.GroupVersionResource{
				Group:    "apiextensions.k8s.io",
				Version:  "v1",
				Resource: "customresourcedefinitions",
			}
			ctx := context.Background()
			crdUnstructured, err := client.Client.Resource(crdGVR).Get(ctx, crdType, metav1.GetOptions{})
			if err == nil && crdUnstructured != nil {
				crdObj.UID = crdUnstructured.GetUID()
			}
		}

		eventType := corev1.EventTypeNormal
		reason := "RouteRegistrationSuccess"
		if status.Phase == model.RouteStatusPhaseRejected {
			eventType = corev1.EventTypeWarning
			reason = "RouteCollisionRejected"
		}

		s.recorder.Event(crdObj, eventType, reason, status.Message)
		log.Debug().Msgf("Successfully emitted %s event for CRD %s", reason, crdType)
	} else {
		log.Warn().Msg("EventRecorder not initialized, skipping event emission")
	}
}

func (s *EchoServer) StopServer() {
	log.Debug().Msg("StopServer invoked")
	ctx, cancel := context.WithTimeout(context.Background(), stopServerCtxTimeout)
	defer cancel()
	if err := s.Echo.Shutdown(ctx); err != nil {
		log.InfraErr(err).Msg("Shutdown signal received")
		return
	}

	log.Debug().Msg("Server exiting")
	address := ":80"
	if s.Config.Server.HTTPPort != "" {
		address = ":" + s.Config.Server.HTTPPort
	}

	if utils.IsServerConfigValid(s.Config) &&
		utils.IsFileExists(s.Config.Server.CertPath) &&
		utils.IsFileExists(s.Config.Server.KeyPath) {
		address = s.Config.Server.Address
	}

	ok := false
	timeout := time.Now().Add(stopServerTimerTimeout)
	for time.Now().Before(timeout) {
		conn, err := net.DialTimeout("tcp", address, stopServerWaitTime)
		if err != nil {
			// informative log. When port is free then error will occur
			log.Debug().Msgf("StopServer: DialTimeout err: %v\n", err)
		}

		if conn == nil {
			ok = true
			break
		}
		conn.Close()
		time.Sleep(stopServerWaitTime)
	}
	if !ok {
		log.InfraError("Error occurred while stopping echo server. TCP port is busy").Msg("")
		return
	}
	log.Debug().Msg("StopServer: success")
}

func NewEchoServer(conf *config.Config, client KubernetesClient, nc *nexusClient.Clientset) *EchoServer {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(middleware.CORS())
	var k8sProxy *httputil.ReverseProxy
	if conf.EnableNexusRuntime {
		// Setup proxy to api server
		k8sProxy = kubeSetupProxy(e)
	}

	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "ACCESS[${time_rfc3339}] method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.Use(auditing.RestEchoMiddleware)

	return &EchoServer{
		// create a new echo_server instance
		Echo:        e,
		Config:      conf,
		Client:      client,
		NexusClient: nc,
		k8sProxy:    k8sProxy,
	}
}

type AssignedInstance struct {
	URL string `json:"url"`
}
type Services struct {
	AssignedInstance []AssignedInstance `json:"allOrgInstances"`
}
type Results struct {
	Services []Services `json:"services"`
}

func WatchForOpenAPISpecChanges(stopCh chan struct{}, openAPISpecDir, openAPISpecFile string, server *EchoServer) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.InfraErr(err).Msg("NewWatcher failed")
		return
	}
	defer watcher.Close()

	go func() {
		for {
			if err := watchOpenAPISpecFile(watcher, stopCh, openAPISpecDir, openAPISpecFile, server); err != nil {
				log.InfraErr(err).Msg("Error watching OpenAPI spec file")
				return
			}
		}
	}()
}

func watchOpenAPISpecFile(watcher *fsnotify.Watcher, stopCh chan struct{},
	openAPISpecDir, openAPISpecFile string, server *EchoServer,
) error {
	_, err := os.Stat(openAPISpecFile)
	if err != nil {
		return watchDirectory(watcher, stopCh, openAPISpecDir, openAPISpecFile, server)
	}
	return watchFile(watcher, stopCh, openAPISpecFile, server)
}

func watchDirectory(watcher *fsnotify.Watcher, stopCh chan struct{},
	openAPISpecDir, openAPISpecFile string, server *EchoServer,
) error {
	if err := watcher.Add(openAPISpecDir); err != nil {
		log.Panic().Msgf("Unable to add watcher for %v: %v", openAPISpecDir, err.Error())
	}
	log.Debug().Msgf("Watching: %v", openAPISpecDir)

	for {
		select {
		case event := <-watcher.Events:
			if handleDirectoryEvent(event, stopCh, openAPISpecFile, server) {
				return nil
			}
		case err := <-watcher.Errors:
			if err != nil {
				log.InfraErr(err).Msg("")
				return err
			}
		}
	}
}

func watchFile(watcher *fsnotify.Watcher, stopCh chan struct{}, openAPISpecFile string, server *EchoServer) error {
	if err := watcher.Add(openAPISpecFile); err != nil {
		log.Panic().Msgf("Unable to add watcher for %v: %v", openAPISpecFile, err.Error())
	}
	log.Debug().Msgf("Watching: %v", openAPISpecFile)

	for {
		select {
		case event := <-watcher.Events:
			if handleFileEvent(event, stopCh, openAPISpecFile, server) {
				return nil
			}
		case err := <-watcher.Errors:
			if err != nil {
				log.InfraErr(err).Msg("")
				return err
			}
		}
	}
}

func handleDirectoryEvent(event fsnotify.Event, stopCh chan struct{}, openAPISpecFile string, server *EchoServer) bool {
	if event.Op == fsnotify.Create && event.Name == openAPISpecFile {
		log.Debug().Msg("Restarting echo server because openApi spec file is created")
		stopCh <- struct{}{}
		server.restartMu.Lock()
		HTTPServerRestartFromOpenAPISpecUpdateCounter++
		log.Info().Msgf("HTTPServerRestartFromOpenAPISpecUpdateCounter: %d", HTTPServerRestartFromOpenAPISpecUpdateCounter)
		server.restartMu.Unlock()
		return true
	}
	log.Trace().Msgf("Received Event on dir watch: %v on file %v", event.Op.String(), event.Name)
	return false
}

func handleFileEvent(event fsnotify.Event, stopCh chan struct{}, openAPISpecFile string, server *EchoServer) bool {
	if event.Op == fsnotify.Write && event.Name == openAPISpecFile {
		log.Debug().Msg("Restarting echo server because openApi spec file is updated")
		stopCh <- struct{}{}
		server.restartMu.Lock()
		HTTPServerRestartFromOpenAPISpecUpdateCounter++
		log.Info().Msgf("HTTPServerRestartFromOpenAPISpecUpdateCounter: %d", HTTPServerRestartFromOpenAPISpecUpdateCounter)
		server.restartMu.Unlock()
		return true
	}
	if event.Op == fsnotify.Remove && event.Name == openAPISpecFile {
		log.Debug().Msg("Restarting echo server because openApi spec file is removed")
		stopCh <- struct{}{}
		server.restartMu.Lock()
		HTTPServerRestartFromOpenAPISpecUpdateCounter++
		log.Info().Msgf("HTTPServerRestartFromOpenAPISpecUpdateCounter: %d", HTTPServerRestartFromOpenAPISpecUpdateCounter)
		server.restartMu.Unlock()
		return true
	}
	log.Trace().Msgf("Received Event on file watch: %v on file %v", event.Op.String(), event.Name)
	return false
}
