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

package mailinglist

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

	"github.com/rossigee/provider-mailgun/apis/mailinglist/v1alpha1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	clients "github.com/rossigee/provider-mailgun/internal/clients"
)

const (
	errNotMailingList = "managed resource is not a MailingList custom resource"
	errTrackPCUsage   = "cannot track ProviderConfig usage"
	errGetPC          = "cannot get ProviderConfig"
	errGetCreds       = "cannot get credentials"

)

// Setup adds a controller that reconciles MailingList managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.MailingListKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.MailingListGroupVersionKind),
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
		For(&v1alpha1.MailingList{}).
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
	cr, ok := mg.(*v1alpha1.MailingList)
	if !ok {
		return nil, errors.New(errNotMailingList)
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

func (c *external) Disconnect(ctx context.Context) error {
	// No persistent connections to clean up
	return nil
}

// ExternalForTesting provides access to external struct for integration tests
type ExternalForTesting struct {
	Client clients.Client
}

// NewExternalForTesting creates a new external struct for testing
func NewExternalForTesting(client clients.Client) *ExternalForTesting {
	return &ExternalForTesting{Client: client}
}

// Observe delegates to the external struct
func (e *ExternalForTesting) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	ext := &external{service: e.Client}
	return ext.Observe(ctx, mg)
}

// Create delegates to the external struct
func (e *ExternalForTesting) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	ext := &external{service: e.Client}
	return ext.Create(ctx, mg)
}

// Update delegates to the external struct
func (e *ExternalForTesting) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	ext := &external{service: e.Client}
	return ext.Update(ctx, mg)
}

// Delete delegates to the external struct
func (e *ExternalForTesting) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	ext := &external{service: e.Client}
	return ext.Delete(ctx, mg)
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.MailingList)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotMailingList)
	}

	mailingList, err := c.service.GetMailingList(ctx, cr.Spec.ForProvider.Address)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get mailing list")
	}

	currentSpec := generateMailingListSpec(cr.Spec.ForProvider)
	upToDate := isMailingListUpToDate(mailingList, currentSpec)

	cr.Status.AtProvider = generateMailingListObservation(mailingList)

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
	cr, ok := mg.(*v1alpha1.MailingList)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotMailingList)
	}

	cr.SetConditions(xpv1.Creating())

	mailingListSpec := generateMailingListSpec(cr.Spec.ForProvider)
	mailingList, err := c.service.CreateMailingList(ctx, mailingListSpec)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create mailing list")
	}

	meta.SetExternalName(cr, mailingList.Address)
	cr.Status.AtProvider = generateMailingListObservation(mailingList)

	return managed.ExternalCreation{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.MailingList)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotMailingList)
	}

	mailingListSpec := generateMailingListSpec(cr.Spec.ForProvider)
	mailingList, err := c.service.UpdateMailingList(ctx, cr.Spec.ForProvider.Address, mailingListSpec)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update mailing list")
	}

	cr.Status.AtProvider = generateMailingListObservation(mailingList)

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.MailingList)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotMailingList)
	}

	cr.SetConditions(xpv1.Deleting())

	err := c.service.DeleteMailingList(ctx, cr.Spec.ForProvider.Address)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete mailing list")
	}

	return managed.ExternalDelete{}, nil
}

// generateMailingListSpec converts the API parameters to client format
func generateMailingListSpec(params v1alpha1.MailingListParameters) *clients.MailingListSpec {
	spec := &clients.MailingListSpec{
		Address: params.Address,
	}

	if params.Name != nil {
		spec.Name = params.Name
	}
	if params.Description != nil {
		spec.Description = params.Description
	}
	if params.AccessLevel != nil {
		spec.AccessLevel = params.AccessLevel
	}
	if params.ReplyPreference != nil {
		spec.ReplyPreference = params.ReplyPreference
	}

	return spec
}

// generateMailingListObservation converts the client response to API format
func generateMailingListObservation(mailingList *clients.MailingList) v1alpha1.MailingListObservation {
	return v1alpha1.MailingListObservation{
		Address:         mailingList.Address,
		Name:            mailingList.Name,
		Description:     mailingList.Description,
		AccessLevel:     mailingList.AccessLevel,
		ReplyPreference: mailingList.ReplyPreference,
		CreatedAt:       mailingList.CreatedAt,
		MembersCount:    mailingList.MembersCount,
	}
}

// isMailingListUpToDate checks if the external resource is up to date
func isMailingListUpToDate(mailingList *clients.MailingList, desired *clients.MailingListSpec) bool {
	// Compare updatable fields
	if desired.Name != nil && mailingList.Name != *desired.Name {
		return false
	}
	if desired.Description != nil && mailingList.Description != *desired.Description {
		return false
	}
	if desired.AccessLevel != nil && mailingList.AccessLevel != *desired.AccessLevel {
		return false
	}
	if desired.ReplyPreference != nil && mailingList.ReplyPreference != *desired.ReplyPreference {
		return false
	}

	return true
}
