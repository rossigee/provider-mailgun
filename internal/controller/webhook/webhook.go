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

package webhook

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

	"github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	clients "github.com/rossigee/provider-mailgun/internal/clients"
	"github.com/rossigee/provider-mailgun/internal/features"
)

const (
	errNotWebhook     = "managed resource is not a Webhook custom resource"
	errTrackPCUsage   = "cannot track ProviderConfig usage"
	errGetPC          = "cannot get ProviderConfig"
	errGetCreds       = "cannot get credentials"
	errResolveDomain  = "cannot resolve domain reference"

)

// Setup adds a controller that reconciles Webhook managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.WebhookKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		cps = append(cps, managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme()))
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.WebhookGroupVersionKind),
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
		For(&v1beta1.Webhook{}).
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
	cr, ok := mg.(*v1beta1.Webhook)
	if !ok {
		return nil, errors.New(errNotWebhook)
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
	// A 'client' used to connect to the external resource API. In practice this
	// would be something like an AWS SDK client.
	service clients.Client
	kube    client.Client
}

func (c *external) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// ExternalForTesting provides access to external struct for integration tests
type ExternalForTesting struct {
	Client clients.Client
	Kube   client.Client
}

// NewExternalForTesting creates a new external struct for testing
func NewExternalForTesting(clientAPI clients.Client, kube client.Client) *ExternalForTesting {
	return &ExternalForTesting{Client: clientAPI, Kube: kube}
}

// Observe delegates to the external struct
func (e *ExternalForTesting) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	ext := &external{service: e.Client, kube: e.Kube}
	return ext.Observe(ctx, mg)
}

// Create delegates to the external struct
func (e *ExternalForTesting) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	ext := &external{service: e.Client, kube: e.Kube}
	return ext.Create(ctx, mg)
}

// Update delegates to the external struct
func (e *ExternalForTesting) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	ext := &external{service: e.Client, kube: e.Kube}
	return ext.Update(ctx, mg)
}

// Delete delegates to the external struct
func (e *ExternalForTesting) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	ext := &external{service: e.Client, kube: e.Kube}
	return ext.Delete(ctx, mg)
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1beta1.Webhook)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotWebhook)
	}

	// Resolve domain reference
	domainName, err := c.resolveDomainReference(ctx, cr)
	if err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, errResolveDomain)
	}

	webhook, err := c.service.GetWebhook(ctx, domainName, cr.Spec.ForProvider.EventType)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get webhook")
	}

	currentSpec := generateWebhookSpec(cr.Spec.ForProvider)
	upToDate := isWebhookUpToDate(webhook, currentSpec)

	cr.Status.AtProvider = generateWebhookObservation(webhook, domainName)

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
	cr, ok := mg.(*v1beta1.Webhook)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotWebhook)
	}

	cr.SetConditions(xpv1.Creating())

	// Resolve domain reference
	domainName, err := c.resolveDomainReference(ctx, cr)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errResolveDomain)
	}

	webhookSpec := generateWebhookSpec(cr.Spec.ForProvider)
	webhook, err := c.service.CreateWebhook(ctx, domainName, webhookSpec)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create webhook")
	}

	// Use domain:eventType as external name
	externalName := domainName + ":" + cr.Spec.ForProvider.EventType
	meta.SetExternalName(cr, externalName)
	cr.Status.AtProvider = generateWebhookObservation(webhook, domainName)

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1beta1.Webhook)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotWebhook)
	}

	// Resolve domain reference
	domainName, err := c.resolveDomainReference(ctx, cr)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errResolveDomain)
	}

	webhookSpec := generateWebhookSpec(cr.Spec.ForProvider)
	webhook, err := c.service.UpdateWebhook(ctx, domainName, cr.Spec.ForProvider.EventType, webhookSpec)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update webhook")
	}

	cr.Status.AtProvider = generateWebhookObservation(webhook, domainName)

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Webhook)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotWebhook)
	}

	cr.SetConditions(xpv1.Deleting())

	// Resolve domain reference
	domainName, err := c.resolveDomainReference(ctx, cr)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errResolveDomain)
	}

	err = c.service.DeleteWebhook(ctx, domainName, cr.Spec.ForProvider.EventType)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete webhook")
	}

	return managed.ExternalDelete{}, nil
}

// resolveDomainReference resolves the domain reference to get the domain name
func (c *external) resolveDomainReference(ctx context.Context, cr *v1beta1.Webhook) (string, error) {
	// For now, use the domain reference name as the domain name
	// In a full implementation, this would look up the actual Domain resource
	// and get its external name or domain name
	if cr.Spec.ForProvider.DomainRef.Name != "" {
		return cr.Spec.ForProvider.DomainRef.Name, nil
	}

	// If no reference name, this is an error
	return "", errors.New("domain reference name is required")
}

// generateWebhookSpec converts the API parameters to client format
func generateWebhookSpec(params v1beta1.WebhookParameters) *clients.WebhookSpec {
	spec := &clients.WebhookSpec{
		URL:       params.URL,
		EventType: params.EventType,
	}

	if params.Username != nil {
		spec.Username = params.Username
	}
	if params.Password != nil {
		spec.Password = params.Password
	}

	return spec
}

// generateWebhookObservation converts the client response to API format
func generateWebhookObservation(webhook *clients.Webhook, domainName string) v1beta1.WebhookObservation {
	return v1beta1.WebhookObservation{
		ID:        webhook.ID,
		EventType: webhook.EventType,
		URL:       webhook.URL,
		Username:  webhook.Username,
		CreatedAt: webhook.CreatedAt,
		Domain:    domainName,
	}
}

// isWebhookUpToDate checks if the external resource is up to date
func isWebhookUpToDate(webhook *clients.Webhook, desired *clients.WebhookSpec) bool {
	// Compare updatable fields
	if webhook.URL != desired.URL {
		return false
	}
	if desired.Username != nil && webhook.Username != *desired.Username {
		return false
	}
	// Note: Password is not returned by the API, so we can't compare it

	return true
}
