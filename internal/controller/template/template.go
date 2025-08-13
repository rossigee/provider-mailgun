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

package template

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-mailgun/apis/template/v1alpha1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	"github.com/rossigee/provider-mailgun/internal/clients"
)

const (
	errNotTemplate     = "managed resource is not a Template custom resource"
	errTrackPCUsage    = "cannot track ProviderConfig usage"
	errGetPC           = "cannot get ProviderConfig"
	errGetCreds        = "cannot get credentials"
	errNewClient       = "cannot create new Service"
	errCreateTemplate  = "cannot create template"
	errGetTemplate     = "cannot get template"
	errUpdateTemplate  = "cannot update template"
	errDeleteTemplate  = "cannot delete template"
)

// Setup adds a controller that reconciles Template managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.TemplateGroupKind.String())

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.TemplateGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.NewProviderConfigUsageTracker(mgr.GetClient(), &apisv1beta1.ProviderConfigUsage{}),
			newServiceFn: clients.NewClient,
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.Template{}).
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
	cr, ok := mg.(*v1alpha1.Template)
	if !ok {
		return nil, errors.New(errNotTemplate)
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

	service := c.newServiceFn(config)
	if service == nil {
		return nil, errors.New(errNewClient)
	}

	return &external{client: service}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	client clients.Client
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
	ext := &external{client: e.Client}
	return ext.Observe(ctx, mg)
}

// Create delegates to the external struct
func (e *ExternalForTesting) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	ext := &external{client: e.Client}
	return ext.Create(ctx, mg)
}

// Update delegates to the external struct
func (e *ExternalForTesting) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	ext := &external{client: e.Client}
	return ext.Update(ctx, mg)
}

// Delete delegates to the external struct
func (e *ExternalForTesting) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	ext := &external{client: e.Client}
	return ext.Delete(ctx, mg)
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.Template)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotTemplate)
	}

	template, err := c.client.GetTemplate(ctx, cr.Spec.ForProvider.Domain, cr.Spec.ForProvider.Name)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, errGetTemplate)
	}

	// Update the status with observed values
	cr.Status.AtProvider.Name = template.Name
	cr.Status.AtProvider.Description = template.Description
	cr.Status.AtProvider.CreatedAt = template.CreatedAt
	cr.Status.AtProvider.CreatedBy = template.CreatedBy

	// Count versions and find active version
	if len(template.Versions) > 0 {
		cr.Status.AtProvider.VersionCount = len(template.Versions)

		// Find the active version
		for _, version := range template.Versions {
			if version.Active {
				cr.Status.AtProvider.ActiveVersion = &v1alpha1.TemplateVersion{
					Tag:       version.Tag,
					Engine:    version.Engine,
					CreatedAt: version.CreatedAt,
					Comment:   version.Comment,
					Active:    version.Active,
				}
				break
			}
		}
	}

	// Check if resource is up to date
	upToDate := cr.Spec.ForProvider.Description == nil || *cr.Spec.ForProvider.Description == template.Description

	cr.SetConditions(xpv1.Available())

	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: upToDate,
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.Template)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotTemplate)
	}

	cr.SetConditions(xpv1.Creating())

	templateSpec := &clients.TemplateSpec{
		Name:        cr.Spec.ForProvider.Name,
		Description: cr.Spec.ForProvider.Description,
		Template:    cr.Spec.ForProvider.Template,
		Engine:      cr.Spec.ForProvider.Engine,
		Comment:     cr.Spec.ForProvider.Comment,
		Tag:         cr.Spec.ForProvider.Tag,
	}

	_, err := c.client.CreateTemplate(ctx, cr.Spec.ForProvider.Domain, templateSpec)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, errCreateTemplate)
	}

	return managed.ExternalCreation{}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.Template)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotTemplate)
	}

	// Only description can be updated for templates
	templateSpec := &clients.TemplateSpec{
		Description: cr.Spec.ForProvider.Description,
	}

	_, err := c.client.UpdateTemplate(ctx, cr.Spec.ForProvider.Domain, cr.Spec.ForProvider.Name, templateSpec)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, errUpdateTemplate)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1alpha1.Template)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotTemplate)
	}

	cr.SetConditions(xpv1.Deleting())

	err := c.client.DeleteTemplate(ctx, cr.Spec.ForProvider.Domain, cr.Spec.ForProvider.Name)
	if err != nil {
		return managed.ExternalDelete{}, errors.Wrap(err, errDeleteTemplate)
	}

	return managed.ExternalDelete{}, nil
}
