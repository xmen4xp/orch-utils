package controllers

import (
	"context"
	"encoding/json"

	"nexus-api-gw/pkg/model"
	"nexus-api-gw/pkg/utils"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ExtensionRestAPIGVR is the GroupVersionResource for ExtensionRestAPI CRs.
var ExtensionRestAPIGVR = schema.GroupVersionResource{
	Group:    "nexus.com",
	Version:  "v1",
	Resource: "extensionrestapis",
}

// ExtensionRestAPIReconciler reconciles ExtensionRestAPI objects.
type ExtensionRestAPIReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	StopCh  chan struct{}
	Dynamic dynamic.Interface
}

//+kubebuilder:rbac:groups=nexus.com,resources=extensionrestapis,verbs=get;list;watch
//+kubebuilder:rbac:groups=nexus.com,resources=extensionrestapis/status,verbs=get;update;patch

// Reconcile handles ExtensionRestAPI CR events.
func (r *ExtensionRestAPIReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = ctrllog.FromContext(ctx)

	eventType := model.Upsert

	obj, err := r.Dynamic.Resource(ExtensionRestAPIGVR).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		eventType = model.Delete
		log.Info().Msgf("ExtensionRestAPI deleted: %s", req.Name)

		// Unregister routes from the registry
		model.GlobalRouteRegistry.UnregisterByCR(req.Name, model.RouteSourceExtensionRestAPI)

		// For delete, we need to construct a minimal spec with just the name
		// The actual URI will be looked up from the cache
		model.ConstructExtensionRestAPI(eventType, model.ExtensionRestAPISpec{Name: req.Name})
		return ctrl.Result{}, nil
	}

	log.Info().Msgf("Received ExtensionRestAPI notification for Name %s Type %s", req.Name, eventType)

	spec, err := r.parseExtensionRestAPI(obj)
	if err != nil {
		log.Error().Msgf("Failed to parse ExtensionRestAPI %s: %v", req.Name, err)
		return ctrl.Result{}, err
	}

	// Determine methods to register
	methods := spec.Methods
	if len(methods) == 0 {
		// Default to all common HTTP methods
		methods = []string{"GET", "PUT", "PATCH", "DELETE"}
	}

	// Atomically check for collisions and register routes
	owner := model.RouteOwner{
		Source: model.RouteSourceExtensionRestAPI,
		CRName: req.Name,
		GVR:    ExtensionRestAPIGVR,
	}
	registeredRoutes, collisions := model.GlobalRouteRegistry.CheckAndRegister(spec.URI, methods, owner)
	if len(collisions) > 0 {
		log.Warn().Msgf("ExtensionRestAPI %s rejected due to route collisions: %+v", req.Name, collisions)

		// Update CR status with rejection
		status := model.NewRejectedStatus(collisions)
		if err := model.UpdateCRStatusWithRouteStatus(ctx, r.Dynamic, req.Name, status); err != nil {
			log.Error().Msgf("Failed to update status for rejected ExtensionRestAPI %s: %v", req.Name, err)
		}

		// Do NOT register routes or trigger restart
		return ctrl.Result{}, nil
	}

	// Update CR status with success
	status := model.NewRegisteredStatus(registeredRoutes)
	if err := model.UpdateCRStatusWithRouteStatus(ctx, r.Dynamic, req.Name, status); err != nil {
		log.Error().Msgf("Failed to update status for ExtensionRestAPI %s: %v", req.Name, err)
	}

	model.ConstructExtensionRestAPI(eventType, spec)

	return ctrl.Result{}, nil
}

// parseExtensionRestAPI extracts the spec from an unstructured ExtensionRestAPI CR.
func (r *ExtensionRestAPIReconciler) parseExtensionRestAPI(obj *unstructured.Unstructured) (model.ExtensionRestAPISpec, error) {
	spec := model.ExtensionRestAPISpec{
		Name: obj.GetName(),
	}

	// Extract spec fields
	specMap, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil || !found {
		return spec, err
	}

	// URI
	if uri, ok := specMap["uri"].(string); ok {
		spec.URI = uri
	}

	// Methods
	if methods, ok := specMap["methods"].([]interface{}); ok {
		for _, m := range methods {
			if method, ok := m.(string); ok {
				spec.Methods = append(spec.Methods, method)
			}
		}
	}

	// OpenAPIPathSpec
	if openAPIPathSpec, ok := specMap["openAPIPathSpec"].(string); ok {
		spec.OpenAPIPathSpec = openAPIPathSpec
	}

	// Parse nexus annotation for hierarchy
	annotations := obj.GetAnnotations()
	if nexusAnnotation, ok := annotations["nexus"]; ok {
		var annotation model.ExtensionRestAPIAnnotation
		if err := json.Unmarshal([]byte(nexusAnnotation), &annotation); err != nil {
			log.Warn().Msgf("Failed to parse nexus annotation for %s: %v", obj.GetName(), err)
		} else {
			spec.Hierarchy = annotation.Hierarchy
			spec.AssociatedNode = annotation.Name
		}
	}

	// Determine datamodel from associated node or CR name
	if spec.AssociatedNode != "" {
		spec.Datamodel = utils.GetDatamodelName(spec.AssociatedNode)
	} else {
		// Use a default datamodel name based on CR name prefix
		spec.Datamodel = utils.GetDatamodelNameFromCRName(obj.GetName())
	}

	log.Debug().Msgf("Parsed ExtensionRestAPI: %+v", spec)
	return spec, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExtensionRestAPIReconciler) SetupWithManager(mgr ctrl.Manager) error {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    "ExtensionRestAPI",
		Group:   "nexus.com",
		Version: "v1",
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(u).
		Complete(r)
}
