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

package bounce

import (
	"context"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/rossigee/provider-mailgun/apis/bounce/v1beta1"
	domainv1beta1 "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	clients "github.com/rossigee/provider-mailgun/internal/clients"
)

const (
	errNotBounce    = "managed resource is not a Bounce custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

)

// Setup adds a controller that reconciles Bounce managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.BounceKind)

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.BounceGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			newServiceFn: clients.NewClient,
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Bounce{}).
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
	cr, ok := mg.(*v1beta1.Bounce)
	if !ok {
		return nil, errors.New(errNotBounce)
	}

	if err := c.usage.Track(ctx, mg); err != nil {
		return nil, errors.Wrap(err, errTrackPCUsage)
	}

	pc := &apisv1beta1.ProviderConfig{}
	pcRef := cr.GetProviderConfigReference()

	// Handle case where no providerConfigRef is specified - default to "default"
	pcName := "default"
	if pcRef != nil && pcRef.Name != "" {
		pcName = pcRef.Name
	}

	// Try namespaced lookup first (ProviderConfig CRD is scope: Namespaced)
	pcNamespace := cr.GetNamespace()
	pcErr := c.kube.Get(ctx, types.NamespacedName{Name: pcName, Namespace: pcNamespace}, pc)
	if pcErr != nil {
		// If namespaced lookup fails, try cluster-scoped as fallback
		clusterErr := c.kube.Get(ctx, types.NamespacedName{Name: pcName}, pc)
		if clusterErr != nil {
			// Both lookups failed, return detailed error
			return nil, errors.Wrapf(pcErr, "cannot get ProviderConfig '%s': tried namespaced lookup in '%s' and cluster-scoped lookup", pcName, pcNamespace)
		}
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

	return &external{service: svc, kube: c.kube}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	service clients.Client
	kube    client.Client
}

func (c *external) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Bounce)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotBounce)
	}

	// Get domain name from domainRef
	domainName, err := c.resolveDomainName(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot resolve domain name")
	}

	// Use the address as the external name
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		externalName = cr.Spec.ForProvider.Address
		meta.SetExternalName(cr, externalName)
	}

	bounce, err := c.service.GetBounce(ctx, domainName, externalName)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "cannot get bounce")
	}

	cr.Status.AtProvider = *bounce

	cr.Status.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        true,
		ResourceLateInitialized: false,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Bounce)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotBounce)
	}

	cr.Status.SetConditions(xpv1.Creating())

	// Get domain name from domainRef
	domainName, err := c.resolveDomainName(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot resolve domain name")
	}

	_, err = c.service.CreateBounce(ctx, domainName, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "cannot create bounce")
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Address)

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	// Bounce entries cannot be updated in Mailgun API
	// They can only be created or deleted
	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Bounce)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotBounce)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	// Get domain name from domainRef
	domainName, err := c.resolveDomainName(ctx, cr)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot resolve domain name")
	}

	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		externalName = cr.Spec.ForProvider.Address
	}

	err = c.service.DeleteBounce(ctx, domainName, externalName)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "cannot delete bounce")
	}

	return managed.ExternalDelete{}, nil
}

// resolveDomainName resolves the domain name from the domainRef
func (c *external) resolveDomainName(ctx context.Context, cr *v1beta1.Bounce) (string, error) {
	domainRefName := cr.Spec.ForProvider.DomainRef.Name

	// If the ref name contains dots, it's likely already a domain name
	if strings.Contains(domainRefName, ".") {
		return domainRefName, nil
	}

	// Look up the Domain resource to get its actual domain name
	domain := &domainv1beta1.Domain{}
	domainKey := types.NamespacedName{
		Name:      domainRefName,
		Namespace: cr.GetNamespace(), // Try same namespace first
	}

	err := c.kube.Get(ctx, domainKey, domain)
	if err != nil {
		// If not found in namespace, try cluster-scoped lookup
		domainKey.Namespace = ""
		err = c.kube.Get(ctx, domainKey, domain)
		if err != nil {
			// If Domain resource not found, assume ref name is the domain name
			return domainRefName, nil
		}
	}

	// Extract domain name from the Domain resource
	if domain.Spec.ForProvider.Name != "" {
		return domain.Spec.ForProvider.Name, nil
	}

	// Fallback to the resource name if spec name is not set
	return domainRefName, nil
}
