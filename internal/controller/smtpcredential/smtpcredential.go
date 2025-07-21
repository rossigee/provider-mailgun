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

package smtpcredential

import (
	"context"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/crossplane-contrib/provider-mailgun/apis/smtpcredential/v1alpha1"
	apisv1beta1 "github.com/crossplane-contrib/provider-mailgun/apis/v1beta1"
	clients "github.com/crossplane-contrib/provider-mailgun/internal/clients"
)

const (
	errNotSMTPCredential = "managed resource is not a SMTPCredential custom resource"
	errTrackPCUsage      = "cannot track ProviderConfig usage"
	errGetPC             = "cannot get ProviderConfig"
	errGetCreds          = "cannot get credentials"

	errNewClient = "cannot create new Service"
)

// Setup adds a controller that reconciles SMTPCredential managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1alpha1.SMTPCredentialKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1alpha1.SMTPCredentialGroupVersionKind),
		managed.WithExternalConnecter(&connector{
			kube:         mgr.GetClient(),
			usage:        newProviderConfigUsageTracker(mgr.GetClient()),
			newServiceFn: clients.NewClient}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1alpha1.SMTPCredential{}).
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
	cr, ok := mg.(*v1alpha1.SMTPCredential)
	if !ok {
		return nil, errors.New(errNotSMTPCredential)
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
	service clients.Client
}

func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*v1alpha1.SMTPCredential)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errNotSMTPCredential)
	}

	// Use the login as external name
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		externalName = cr.Spec.ForProvider.Login
		meta.SetExternalName(cr, externalName)
	}

	credential, err := c.service.GetSMTPCredential(ctx, cr.Spec.ForProvider.Domain, cr.Spec.ForProvider.Login)
	if err != nil {
		if clients.IsNotFound(err) {
			return managed.ExternalObservation{ResourceExists: false}, nil
		}
		return managed.ExternalObservation{}, errors.Wrap(err, "failed to get SMTP credential")
	}

	// Update observed state
	cr.Status.AtProvider = v1alpha1.SMTPCredentialObservation{
		Login:     credential.Login,
		CreatedAt: credential.CreatedAt,
		State:     credential.State,
	}

	// Credentials are always up to date - we can't compare passwords
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: true,
		ConnectionDetails: managed.ConnectionDetails{
			"smtp_host":     []byte("smtp.mailgun.org"),
			"smtp_port":     []byte("587"),
			"smtp_username": []byte(credential.Login),
			"smtp_password": []byte(credential.Password),
		},
	}, nil
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1alpha1.SMTPCredential)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotSMTPCredential)
	}

	cr.SetConditions(xpv1.Creating())

	spec := &clients.SMTPCredentialSpec{
		Login:    cr.Spec.ForProvider.Login,
		Password: cr.Spec.ForProvider.Password,
	}

	credential, err := c.service.CreateSMTPCredential(ctx, cr.Spec.ForProvider.Domain, spec)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create SMTP credential")
	}

	meta.SetExternalName(cr, credential.Login)

	// Update observed state
	cr.Status.AtProvider = v1alpha1.SMTPCredentialObservation{
		Login:     credential.Login,
		CreatedAt: credential.CreatedAt,
		State:     credential.State,
	}

	// Return connection details including the password
	password := ""
	if credential.Password != "" {
		password = credential.Password
	} else if cr.Spec.ForProvider.Password != nil {
		password = *cr.Spec.ForProvider.Password
	}

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"smtp_host":     []byte("smtp.mailgun.org"),
			"smtp_port":     []byte("587"),
			"smtp_username": []byte(credential.Login),
			"smtp_password": []byte(password),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*v1alpha1.SMTPCredential)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotSMTPCredential)
	}

	// Only update if password is provided
	if cr.Spec.ForProvider.Password != nil {
		_, err := c.service.UpdateSMTPCredential(ctx,
			cr.Spec.ForProvider.Domain,
			cr.Spec.ForProvider.Login,
			*cr.Spec.ForProvider.Password)
		if err != nil {
			return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update SMTP credential")
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) error {
	cr, ok := mg.(*v1alpha1.SMTPCredential)
	if !ok {
		return errors.New(errNotSMTPCredential)
	}

	cr.SetConditions(xpv1.Deleting())

	err := c.service.DeleteSMTPCredential(ctx, cr.Spec.ForProvider.Domain, cr.Spec.ForProvider.Login)
	if err != nil && !clients.IsNotFound(err) {
		return errors.Wrap(err, "failed to delete SMTP credential")
	}

	return nil
}

// providerConfigUsageTracker is a custom tracker that ensures ProviderConfigUsage
// resources are created in the correct namespace.
type providerConfigUsageTracker struct {
	kube client.Client
}

func newProviderConfigUsageTracker(kube client.Client) resource.Tracker {
	return &providerConfigUsageTracker{kube: kube}
}

func (t *providerConfigUsageTracker) Track(ctx context.Context, mg resource.Managed) error {
	// Create ProviderConfigUsage in the same namespace as the managed resource
	pcu := &apisv1beta1.ProviderConfigUsage{}
	pcu.SetName(string(mg.GetUID()))
	pcu.SetNamespace(mg.GetNamespace())
	pcu.SetOwnerReferences([]metav1.OwnerReference{meta.AsOwner(meta.TypedReferenceTo(mg, mg.GetObjectKind().GroupVersionKind()))})

	pcRef := mg.GetProviderConfigReference()
	if pcRef != nil {
		pcu.SetProviderConfigReference(*pcRef)
	}

	resRef := meta.TypedReferenceTo(mg, mg.GetObjectKind().GroupVersionKind())
	if resRef != nil {
		pcu.SetResourceReference(*resRef)
	}

	err := t.kube.Create(ctx, pcu)
	if err != nil && client.IgnoreAlreadyExists(err) != nil {
		return err
	}
	return nil
}
