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

package domain

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

	"github.com/rossigee/provider-mailgun/apis/domain/v1alpha1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	clients "github.com/rossigee/provider-mailgun/internal/clients"
)

const (
	errNotDomain    = "managed resource is not a Domain custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

)

// Setup adds a controller that reconciles Domain managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.DomainKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.DomainGroupVersionKind),
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
		For(&v1alpha1.Domain{}).
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
	cr, ok := mg.(*v1alpha1.Domain)
	if !ok {
		return nil, errors.New(errNotDomain)
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
	cr, ok := mg.(*v1alpha1.Domain)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotDomain)
	}

	domain, err := c.service.GetDomain(ctx, cr.Spec.ForProvider.Name)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get domain")
	}

	currentSpec := generateDomainSpec(cr.Spec.ForProvider)
	upToDate := isDomainUpToDate(domain, currentSpec)

	cr.Status.AtProvider = generateDomainObservation(domain)

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
		ConnectionDetails: managed.ConnectionDetails{
			"smtp_login":    []byte(domain.SMTPLogin),
			"smtp_password": []byte(domain.SMTPPassword),
		},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Domain)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDomain)
	}

	cr.SetConditions(xpv1.Creating())

	domainSpec := generateDomainSpec(cr.Spec.ForProvider)
	domain, err := c.service.CreateDomain(ctx, domainSpec)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create domain")
	}

	meta.SetExternalName(cr, domain.Name)
	cr.Status.AtProvider = generateDomainObservation(domain)

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{
			"smtp_login":    []byte(domain.SMTPLogin),
			"smtp_password": []byte(domain.SMTPPassword),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Domain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDomain)
	}

	domainSpec := generateDomainSpec(cr.Spec.ForProvider)
	domain, err := c.service.UpdateDomain(ctx, cr.Spec.ForProvider.Name, domainSpec)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update domain")
	}

	cr.Status.AtProvider = generateDomainObservation(domain)

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{
			"smtp_login":    []byte(domain.SMTPLogin),
			"smtp_password": []byte(domain.SMTPPassword),
		},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.Domain)
	if !ok {
		return errors.New(errNotDomain)
	}

	cr.SetConditions(xpv1.Deleting())

	err := c.service.DeleteDomain(ctx, cr.Spec.ForProvider.Name)
	if err != nil && !clients.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete domain")
	}

	return nil
}

// generateDomainSpec converts the API parameters to client format
func generateDomainSpec(params v1alpha1.DomainParameters) *clients.DomainSpec {
	spec := &clients.DomainSpec{
		Name: params.Name,
	}

	if params.Type != nil {
		spec.Type = params.Type
	}
	if params.ForceDKIMAuthority != nil {
		spec.ForceDKIMAuthority = params.ForceDKIMAuthority
	}
	if params.DKIMKeySize != nil {
		spec.DKIMKeySize = params.DKIMKeySize
	}
	if params.SMTPPassword != nil {
		spec.SMTPPassword = params.SMTPPassword
	}
	if params.SpamAction != nil {
		spec.SpamAction = params.SpamAction
	}
	if params.WebScheme != nil {
		spec.WebScheme = params.WebScheme
	}
	if params.Wildcard != nil {
		spec.Wildcard = params.Wildcard
	}
	if len(params.IPs) > 0 {
		spec.IPs = params.IPs
	}

	return spec
}

// generateDomainObservation converts the client response to API format
func generateDomainObservation(domain *clients.Domain) v1alpha1.DomainObservation {
	obs := v1alpha1.DomainObservation{
		ID:           domain.Name,
		State:        domain.State,
		CreatedAt:    domain.CreatedAt,
		SMTPLogin:    domain.SMTPLogin,
		SMTPPassword: domain.SMTPPassword,
	}

	// Convert DNS records
	if len(domain.RequiredDNSRecords) > 0 {
		obs.RequiredDNSRecords = make([]v1alpha1.DNSRecord, len(domain.RequiredDNSRecords))
		for i, record := range domain.RequiredDNSRecords {
			obs.RequiredDNSRecords[i] = v1alpha1.DNSRecord{
				Name:     record.Name,
				Type:     record.Type,
				Value:    record.Value,
				Priority: record.Priority,
				Valid:    record.Valid,
			}
		}
	}

	if len(domain.ReceivingDNSRecords) > 0 {
		obs.ReceivingDNSRecords = make([]v1alpha1.DNSRecord, len(domain.ReceivingDNSRecords))
		for i, record := range domain.ReceivingDNSRecords {
			obs.ReceivingDNSRecords[i] = v1alpha1.DNSRecord{
				Name:     record.Name,
				Type:     record.Type,
				Value:    record.Value,
				Priority: record.Priority,
				Valid:    record.Valid,
			}
		}
	}

	if len(domain.SendingDNSRecords) > 0 {
		obs.SendingDNSRecords = make([]v1alpha1.DNSRecord, len(domain.SendingDNSRecords))
		for i, record := range domain.SendingDNSRecords {
			obs.SendingDNSRecords[i] = v1alpha1.DNSRecord{
				Name:     record.Name,
				Type:     record.Type,
				Value:    record.Value,
				Priority: record.Priority,
				Valid:    record.Valid,
			}
		}
	}

	return obs
}

// isDomainUpToDate checks if the external resource is up to date
func isDomainUpToDate(domain *clients.Domain, desired *clients.DomainSpec) bool {
	// Compare updatable fields only
	// Note: Most domain fields cannot be updated after creation in Mailgun
	// We only check the fields that can be modified

	if desired.SpamAction != nil && domain.Type != *desired.SpamAction {
		return false
	}
	// WebScheme and Wildcard are write-only fields in Mailgun API
	// They are not returned in the domain response, so we cannot compare them
	// We assume they are up to date since they were set during creation/update
	// Note: These settings can only be verified through separate tracking/subdomain API calls
	// which are not currently implemented in this provider
	if desired.WebScheme != nil {
		_ = desired.WebScheme // prevent unused variable warning
	}
	if desired.Wildcard != nil {
		_ = desired.Wildcard // prevent unused variable warning
	}

	return true
}
