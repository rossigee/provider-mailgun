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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	clients "github.com/rossigee/provider-mailgun/internal/clients"
	"github.com/rossigee/provider-mailgun/internal/metrics"
	"github.com/rossigee/provider-mailgun/internal/tracing"
)

const (
	errNotSMTPCredential = "managed resource is not a SMTPCredential custom resource"
	errTrackPCUsage      = "cannot track ProviderConfig usage"
	errGetPC             = "cannot get ProviderConfig"
	errGetCreds          = "cannot get credentials"

)

// Setup adds a controller that reconciles SMTPCredential managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.SMTPCredentialKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.SMTPCredentialGroupVersionKind),
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
		For(&v1beta1.SMTPCredential{}).
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
	cr, ok := mg.(*v1beta1.SMTPCredential)
	if !ok {
		return nil, errors.New(errNotSMTPCredential)
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

	// Always try crossplane-system namespace first for ProviderConfigs
	// This is the standard location for cluster-wide ProviderConfigs
	pcErr := c.kube.Get(ctx, types.NamespacedName{Name: pcName, Namespace: "crossplane-system"}, pc)
	if pcErr != nil {
		// If not found in crossplane-system, try the managed resource's namespace as fallback
		pcNamespace := cr.GetNamespace()
		if pcNamespace != "crossplane-system" {
			fallbackErr := c.kube.Get(ctx, types.NamespacedName{Name: pcName, Namespace: pcNamespace}, pc)
			if fallbackErr != nil {
				// Both lookups failed, return detailed error
				return nil, errors.Wrapf(pcErr, "cannot get ProviderConfig '%s': tried crossplane-system and namespace '%s'", pcName, pcNamespace)
			}
		} else {
			// We already tried crossplane-system, return the original error
			return nil, errors.Wrapf(pcErr, "cannot get ProviderConfig '%s' in namespace 'crossplane-system'", pcName)
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
	op := tracing.StartOperation(ctx, tracing.SpanResourceObserve,
		tracing.AttrResourceType.String("SMTPCredential"),
		tracing.AttrResourceName.String(mg.GetName()),
		tracing.AttrOperation.String("observe"),
	)
	defer op.End()

	timer := metrics.NewOperationTimer()
	logger := log.FromContext(ctx).WithValues(
		"operation", "observe",
		"resource", mg.GetName(),
		"namespace", mg.GetNamespace(),
	)

	cr, ok := mg.(*v1beta1.SMTPCredential)
	if !ok {
		err := errors.New(errNotSMTPCredential)
		logger.Error(nil, "managed resource is not a SMTPCredential", "type", mg.GetObjectKind())
		timer.RecordResourceOperation("smtpcredential", "observe", "error")
		op.RecordError(err)
		return managed.ExternalObservation{}, err
	}

	logger = logger.WithValues(
		"domain", cr.Spec.ForProvider.Domain,
		"login", cr.Spec.ForProvider.Login,
	)

	op.SetAttribute("domain", cr.Spec.ForProvider.Domain)
	op.SetAttribute("login", cr.Spec.ForProvider.Login)

	logger.Info("starting SMTP credential observation")

	// Use the login as external name
	externalName := meta.GetExternalName(cr)
	if externalName == "" {
		externalName = cr.Spec.ForProvider.Login
		meta.SetExternalName(cr, externalName)
		logger.Info("set external name", "externalName", externalName)
	}

	// SMTP credentials are write-only in Mailgun API - we can't read them back
	// Instead, we check if we have connection details stored (indicating successful creation)
	// If not, we'll trigger recreation to rotate credentials

	// Check if we have stored credentials from previous creation
	secretRef := cr.GetWriteConnectionSecretToReference()
	if secretRef == nil {
		logger.Info("no secret reference configured, treating as new resource")
		timer.RecordResourceOperation("smtpcredential", "observe", "not_found")
		return managed.ExternalObservation{ResourceExists: false}, nil
	}
	secretName := secretRef.Name
	secretNamespace := secretRef.Namespace

	logger = logger.WithValues("secretName", secretName, "secretNamespace", secretNamespace)

	secret := &corev1.Secret{}
	secretKey := types.NamespacedName{Name: secretName, Namespace: secretNamespace}
	err := c.kube.Get(ctx, secretKey, secret)

	// If secret exists and has credentials, consider resource as existing
	if err == nil && secret.Data != nil && len(secret.Data) > 0 {
		logger.Info("SMTP credential exists with stored secret",
			"secretDataKeys", getSecretDataKeys(secret.Data))

		metrics.RecordSecretOperation("get", "success")
		timer.RecordResourceOperation("smtpcredential", "observe", "success")

		op.SetAttribute("secret.found", true)
		op.SetAttribute("secret.keys_count", len(secret.Data))
		op.SetAttribute("resource.exists", true)
		op.SetAttribute("resource.up_to_date", true)

		// Resource exists and we have credentials stored
		cr.Status.AtProvider = v1beta1.SMTPCredentialObservation{
			Login: cr.Spec.ForProvider.Login,
			State: "active", // Assume active since we have stored credentials
		}

		return managed.ExternalObservation{
			ResourceExists:   true,
			ResourceUpToDate: true,
			ConnectionDetails: managed.ConnectionDetails{
				"smtp_host":     []byte("smtp.mailgun.org"),
				"smtp_port":     []byte("587"),
				"smtp_username": []byte(cr.Spec.ForProvider.Login),
				// Password is already stored in the secret, don't overwrite
			},
		}, nil
	}

	// If secret doesn't exist or is empty, we need to create/rotate credentials
	// This handles both new resources and imported ones that need rotation
	if err != nil {
		logger.Info("SMTP credential secret not found, treating as new resource",
			"error", err.Error())
		metrics.RecordSecretOperation("get", "not_found")
		op.SetAttribute("secret.found", false)
		op.SetAttribute("secret.reason", "not_found")
	} else {
		logger.Info("SMTP credential secret exists but is empty, needs rotation")
		metrics.RecordSecretOperation("get", "empty")
		op.SetAttribute("secret.found", true)
		op.SetAttribute("secret.reason", "empty")
	}
	timer.RecordResourceOperation("smtpcredential", "observe", "not_found")
	op.SetAttribute("resource.exists", false)
	return managed.ExternalObservation{ResourceExists: false}, nil
}

// getSecretDataKeys returns the keys present in secret data for logging (without values)
func getSecretDataKeys(data map[string][]byte) []string {
	keys := make([]string, 0, len(data))
	for k := range data {
		keys = append(keys, k)
	}
	return keys
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	op := tracing.StartOperation(ctx, tracing.SpanResourceCreate,
		tracing.AttrResourceType.String("SMTPCredential"),
		tracing.AttrResourceName.String(mg.GetName()),
		tracing.AttrOperation.String("create"),
	)
	defer op.End()

	timer := metrics.NewOperationTimer()
	logger := log.FromContext(ctx).WithValues(
		"operation", "create",
		"resource", mg.GetName(),
		"namespace", mg.GetNamespace(),
	)

	cr, ok := mg.(*v1beta1.SMTPCredential)
	if !ok {
		err := errors.New(errNotSMTPCredential)
		logger.Error(nil, "managed resource is not a SMTPCredential", "type", mg.GetObjectKind())
		timer.RecordResourceOperation("smtpcredential", "create", "error")
		op.RecordError(err)
		return managed.ExternalCreation{}, err
	}

	logger = logger.WithValues(
		"domain", cr.Spec.ForProvider.Domain,
		"login", cr.Spec.ForProvider.Login,
	)

	op.SetAttribute("domain", cr.Spec.ForProvider.Domain)
	op.SetAttribute("login", cr.Spec.ForProvider.Login)

	logger.Info("starting SMTP credential creation")

	cr.SetConditions(xpv1.Creating())

	// Check if this is an imported resource that needs rotation
	// Only apply rotation strategy for imported resources, not new ones
	externalName := meta.GetExternalName(cr)
	isImported := externalName != "" && externalName != cr.Spec.ForProvider.Login

	if isImported {
		// Implement rotation strategy for imported resources: delete existing credential first
		// This ensures we get fresh credentials for imported resources
		logger.Info("deleting existing SMTP credential for imported resource rotation")
		op.SetAttribute("rotation_strategy", true)
		err := c.service.DeleteSMTPCredential(ctx, cr.Spec.ForProvider.Domain, cr.Spec.ForProvider.Login)
		if err != nil && !clients.IsNotFound(err) {
			// If deletion fails for reasons other than "not found", that's an error
			logger.Error(err, "failed to delete existing SMTP credential during rotation")
			op.RecordError(err)
			return managed.ExternalCreation{}, errors.Wrap(err, "failed to delete existing SMTP credential during rotation")
		}

		if clients.IsNotFound(err) {
			logger.Info("no existing credential found during rotation, proceeding with fresh creation")
			op.SetAttribute("rotation.existing_found", false)
		} else if err == nil {
			logger.Info("existing credential deleted successfully during rotation, proceeding with fresh creation")
			op.SetAttribute("rotation.existing_found", true)
			op.SetAttribute("rotation.deleted", true)
		}
	} else {
		logger.Info("creating new SMTP credential (no rotation needed)")
		op.SetAttribute("rotation_strategy", false)
	}

	// Use provided password or let Mailgun generate one
	password := cr.Spec.ForProvider.Password
	if password == nil {
		logger.Info("no password provided, letting Mailgun generate one")
	} else {
		logger.Info("using provided password for SMTP credential")
	}

	spec := &clients.SMTPCredentialSpec{
		Login:    cr.Spec.ForProvider.Login,
		Password: password,
	}

	logger.Info("creating new SMTP credential via Mailgun API")
	apiTimer := metrics.NewOperationTimer()
	credential, err := c.service.CreateSMTPCredential(ctx, cr.Spec.ForProvider.Domain, spec)
	if err != nil {
		logger.Error(err, "failed to create SMTP credential")
		apiTimer.RecordMailgunAPIRequest("create_smtp_credential", cr.Spec.ForProvider.Domain, "error")
		timer.RecordResourceOperation("smtpcredential", "create", "error")
		op.RecordError(err)
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create SMTP credential")
	}

	apiTimer.RecordMailgunAPIRequest("create_smtp_credential", cr.Spec.ForProvider.Domain, "success")
	op.SetAttribute("credential.created", true)
	op.SetAttribute("credential.state", credential.State)
	logger.Info("SMTP credential created successfully",
		"externalName", credential.Login,
		"state", credential.State)

	meta.SetExternalName(cr, credential.Login)

	// Update observed state
	cr.Status.AtProvider = v1beta1.SMTPCredentialObservation{
		Login:     credential.Login,
		CreatedAt: credential.CreatedAt,
		State:     credential.State,
	}

	// Return connection details including the password from Mailgun
	// Use the password returned by Mailgun API (either provided or generated by Mailgun)
	connectionPassword := credential.Password
	if connectionPassword == "" && password != nil {
		// Fallback to provided password if Mailgun didn't return one
		connectionPassword = *password
	}

	logger.Info("returning connection details for secret storage",
		"passwordLength", len(connectionPassword),
		"passwordSource", func() string {
			if password == nil {
				return "mailgun-generated"
			} else {
				return "user-provided"
			}
		}())
	timer.RecordResourceOperation("smtpcredential", "create", "success")

	return managed.ExternalCreation{
		ConnectionDetails: managed.ConnectionDetails{
			"smtp_host":     []byte("smtp.mailgun.org"),
			"smtp_port":     []byte("587"),
			"smtp_username": []byte(credential.Login),
			"smtp_password": []byte(connectionPassword),
		},
	}, nil
}

func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	op := tracing.StartOperation(ctx, tracing.SpanResourceUpdate,
		tracing.AttrResourceType.String("SMTPCredential"),
		tracing.AttrResourceName.String(mg.GetName()),
		tracing.AttrOperation.String("update"),
	)
	defer op.End()

	cr, ok := mg.(*v1beta1.SMTPCredential)
	if !ok {
		err := errors.New(errNotSMTPCredential)
		op.RecordError(err)
		return managed.ExternalUpdate{}, err
	}

	op.SetAttribute("domain", cr.Spec.ForProvider.Domain)
	op.SetAttribute("login", cr.Spec.ForProvider.Login)

	// Only update if password is provided
	if cr.Spec.ForProvider.Password != nil {
		op.SetAttribute("password.provided", true)
		_, err := c.service.UpdateSMTPCredential(ctx,
			cr.Spec.ForProvider.Domain,
			cr.Spec.ForProvider.Login,
			*cr.Spec.ForProvider.Password)
		if err != nil {
			op.RecordError(err)
			return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update SMTP credential")
		}
		op.SetAttribute("credential.updated", true)
	} else {
		op.SetAttribute("password.provided", false)
	}

	return managed.ExternalUpdate{}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	op := tracing.StartOperation(ctx, tracing.SpanResourceDelete,
		tracing.AttrResourceType.String("SMTPCredential"),
		tracing.AttrResourceName.String(mg.GetName()),
		tracing.AttrOperation.String("delete"),
	)
	defer op.End()

	cr, ok := mg.(*v1beta1.SMTPCredential)
	if !ok {
		err := errors.New(errNotSMTPCredential)
		op.RecordError(err)
		return managed.ExternalDelete{}, err
	}

	op.SetAttribute("domain", cr.Spec.ForProvider.Domain)
	op.SetAttribute("login", cr.Spec.ForProvider.Login)

	cr.SetConditions(xpv1.Deleting())

	err := c.service.DeleteSMTPCredential(ctx, cr.Spec.ForProvider.Domain, cr.Spec.ForProvider.Login)
	if err != nil && !clients.IsNotFound(err) {
		op.RecordError(err)
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete SMTP credential")
	}

	if clients.IsNotFound(err) {
		op.SetAttribute("credential.already_deleted", true)
	} else {
		op.SetAttribute("credential.deleted", true)
	}

	return managed.ExternalDelete{}, nil
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
	// Create ProviderConfigUsage - namespaced resource per CRD definition
	pcu := &apisv1beta1.ProviderConfigUsage{}
	pcu.SetName(string(mg.GetUID()))
	// Set namespace to crossplane-system if managed resource has no namespace (cluster-scoped)
	// Otherwise use the managed resource namespace
	namespace := mg.GetNamespace()
	if namespace == "" {
		namespace = "crossplane-system"
	}
	pcu.SetNamespace(namespace)
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
