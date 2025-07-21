/*
Copyright 2025 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package route

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-mailgun/apis/route/v1alpha1"
	apisv1beta1 "github.com/crossplane-contrib/provider-mailgun/apis/v1beta1"
	clients "github.com/crossplane-contrib/provider-mailgun/internal/clients"
)

const (
	errNotRoute     = "managed resource is not a Route custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// Setup adds a controller that reconciles Route managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.RouteKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.RouteGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1beta1.ProviderConfigUsage{}),
			newServiceFn: clients.NewClient}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Route{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(config *clients.Config) clients.Client
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1alpha1.Route)
	if !ok {
		return nil, errors.New(errNotRoute)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1beta1.ProviderConfig{}
	if err := c.kube.Get(ctx, types.NamespacedName{Name: cr.GetProviderConfigReference().Name}, pc); err != nil {
		return nil, errors.Wrap(err, errGetPC)
	}

	cd := pc.Spec.Credentials
	_, err := resource.CommonCredentialExtractor(ctx, cd.Source, c.kube, cd.CommonCredentialSelectors)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	config, err := clients.GetConfig(ctx, c.kube, mg)
	if err != nil {
		return nil, errors.Wrap(err, errGetCreds)
	}

	svc := c.newServiceFn(config)

	return &external{service: svc}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service clients.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Route)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotRoute)
	}

	// For routes, we need to get the route by ID if it exists
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return managed.ExternalObservation{ResourceExists: false}, nil
	}

	route, err := c.service.GetRoute(ctx, externalName)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get route")
	}

	currentSpec := generateRouteSpec(cr.Spec.ForProvider)
	upToDate := isRouteUpToDate(route, currentSpec)

	cr.Status.AtProvider = generateRouteObservation(route)

	return managed.ExternalObservation{
		// Return false when the external resource does not exist. This lets
		// the managed resource reconciler know that it needs to call Create to
		// (re)create the resource, or that it has successfully been deleted.
		ResourceExists: true,

		// Return false when the external resource exists, but it not up to date
		// with the desired managed resource state. This lets the managed
		// resource reconciler know that it needs to call Update.
		ResourceUpToDate: upToDate,

		// Return any details that may be required to connect to the external
		// resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Route)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotRoute)
	}

	cr.SetConditions(xpv1.Creating())

	routeSpec := generateRouteSpec(cr.Spec.ForProvider)
	route, err := c.service.CreateRoute(ctx, routeSpec)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create route")
	}

	meta.SetExternalName(cr, route.ID)
	cr.Status.AtProvider = generateRouteObservation(route)

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Route)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotRoute)
	}

	externalName := meta.GetExternalName(cr)
	routeSpec := generateRouteSpec(cr.Spec.ForProvider)
	route, err := c.service.UpdateRoute(ctx, externalName, routeSpec)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update route")
	}

	cr.Status.AtProvider = generateRouteObservation(route)

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Route)
	if !ok {
		return errors.New(errNotRoute)
	}

	cr.SetConditions(xpv1.Deleting())

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		return nil // Already deleted
	}

	err := c.service.DeleteRoute(ctx, externalName)
	if err != nil && !clients.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete route")
	}

	return nil
}

// generateRouteSpec converts the API parameters to client format
func generateRouteSpec(params v1alpha1.RouteParameters) *clients.RouteSpec {
	spec := &clients.RouteSpec{
		Expression: params.Expression,
	}

	if params.Priority != nil {
		spec.Priority = params.Priority
	}
	if params.Description != nil {
		spec.Description = params.Description
	}

	// Convert actions
	if len(params.Actions) > 0 {
		spec.Actions = make([]clients.RouteAction, len(params.Actions))
		for i, action := range params.Actions {
			spec.Actions[i] = clients.RouteAction{
				Type: action.Type,
			}
			if action.Destination != nil {
				spec.Actions[i].Destination = action.Destination
			}
		}
	}

	return spec
}

// generateRouteObservation converts the client response to API format
func generateRouteObservation(route *clients.Route) v1alpha1.RouteObservation {
	obs := v1alpha1.RouteObservation{
		ID:          route.ID,
		Priority:    route.Priority,
		Description: route.Description,
		Expression:  route.Expression,
		CreatedAt:   route.CreatedAt,
	}

	// Convert actions
	if len(route.Actions) > 0 {
		obs.Actions = make([]v1alpha1.RouteAction, len(route.Actions))
		for i, action := range route.Actions {
			obs.Actions[i] = v1alpha1.RouteAction{
				Type: action.Type,
			}
			if action.Destination != nil {
				obs.Actions[i].Destination = action.Destination
			}
		}
	}

	return obs
}

// isRouteUpToDate checks if the external resource is up to date
func isRouteUpToDate(route *clients.Route, desired *clients.RouteSpec) bool {
	// Compare updatable fields
	if route.Expression != desired.Expression {
		return false
	}
	if desired.Priority != nil && route.Priority != *desired.Priority {
		return false
	}
	if desired.Description != nil && route.Description != *desired.Description {
		return false
	}

	// Compare actions
	if len(route.Actions) != len(desired.Actions) {
		return false
	}
	for i, action := range route.Actions {
		if i >= len(desired.Actions) {
			return false
		}
		desiredAction := desired.Actions[i]
		if action.Type != desiredAction.Type {
			return false
		}
		if desiredAction.Destination != nil && action.Destination != nil {
			if *action.Destination != *desiredAction.Destination {
				return false
			}
		} else if desiredAction.Destination != action.Destination {
			return false
		}
	}

	return true
}
