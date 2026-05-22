package controllers

import (
	"context"

	"nexus-api-gw/pkg/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"
)

// ExtensionRestAPIEndpointGVR is the GroupVersionResource for ExtensionRestAPIEndpoint CRs.
var ExtensionRestAPIEndpointGVR = schema.GroupVersionResource{
	Group:    "nexus.com",
	Version:  "v1",
	Resource: "extensionrestapiendpoints",
}

// ExtensionRestAPIEndpointReconciler reconciles ExtensionRestAPIEndpoint objects.
type ExtensionRestAPIEndpointReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	StopCh  chan struct{}
	Dynamic dynamic.Interface
}

//+kubebuilder:rbac:groups=nexus.com,resources=extensionrestapiendpoints,verbs=get;list;watch
//+kubebuilder:rbac:groups=nexus.com,resources=extensionrestapiendpoints/status,verbs=get

// Reconcile handles ExtensionRestAPIEndpoint CR events.
func (r *ExtensionRestAPIEndpointReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = ctrllog.FromContext(ctx)

	eventType := model.Upsert

	obj, err := r.Dynamic.Resource(ExtensionRestAPIEndpointGVR).Get(ctx, req.Name, metav1.GetOptions{})
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return ctrl.Result{}, err
		}
		eventType = model.Delete
		log.Info().Msgf("ExtensionRestAPIEndpoint deleted: %s", req.Name)
		model.ConstructExtensionRestAPIEndpoint(eventType, model.ExtensionRestAPIEndpointSpec{Name: req.Name})
		return ctrl.Result{}, nil
	}

	log.Info().Msgf("Received ExtensionRestAPIEndpoint notification for Name %s Type %s", req.Name, eventType)

	spec, err := r.parseExtensionRestAPIEndpoint(obj)
	if err != nil {
		log.Error().Msgf("Failed to parse ExtensionRestAPIEndpoint %s: %v", req.Name, err)
		return ctrl.Result{}, err
	}

	model.ConstructExtensionRestAPIEndpoint(eventType, spec)

	return ctrl.Result{}, nil
}

// parseExtensionRestAPIEndpoint extracts the spec from an unstructured ExtensionRestAPIEndpoint CR.
func (r *ExtensionRestAPIEndpointReconciler) parseExtensionRestAPIEndpoint(obj *unstructured.Unstructured) (model.ExtensionRestAPIEndpointSpec, error) {
	spec := model.ExtensionRestAPIEndpointSpec{
		Name: obj.GetName(),
	}

	// Extract spec fields
	specMap, found, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil || !found {
		return spec, err
	}

	// ExtensionRestAPIRef
	if ref, ok := specMap["extensionRestAPIRef"].(string); ok {
		spec.ExtensionRestAPIRef = ref
	}

	// Service - fully qualified DNS name
	if service, ok := specMap["service"].(string); ok {
		spec.Service = service
	}

	// Port
	if port, ok := specMap["port"].(string); ok {
		spec.Port = port
	}

	log.Debug().Msgf("Parsed ExtensionRestAPIEndpoint: %+v", spec)
	return spec, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExtensionRestAPIEndpointReconciler) SetupWithManager(mgr ctrl.Manager) error {
	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Kind:    "ExtensionRestAPIEndpoint",
		Group:   "nexus.com",
		Version: "v1",
	})

	return ctrl.NewControllerManagedBy(mgr).
		For(u).
		Complete(r)
}
