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
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/crossplane/crossplane-runtime/v2/pkg/controller"
	"github.com/crossplane/crossplane-runtime/v2/pkg/event"
	"github.com/crossplane/crossplane-runtime/v2/pkg/meta"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	v1beta1 "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	"github.com/rossigee/provider-mailgun/internal/clients"
)

const (
	errNotDomain    = "managed resource is not a Domain custom resource"
	errTrackPCUsage = "cannot track ProviderConfig usage"
	errGetPC        = "cannot get ProviderConfig"
	errGetCreds     = "cannot get credentials"

	conditionTypeDNSVerified xpv1.ConditionType = "DNSVerified"

	eventReasonDNSVerified        = "DNSVerified"
	eventReasonDNSInvalid         = "DNSInvalid"
	eventReasonDNSReverify        = "DNSReverifyRequested"
	eventReasonDNSRecordsRequired = "DNSRecordsRequired"
	eventReasonDNSNotPropagated   = "DNSNotPropagated"
	eventReasonDNSValueMismatch   = "DNSValueMismatch"
	eventReasonDNSRecordMatches   = "DNSRecordMatches"
	eventReasonDNSProbeError      = "DNSProbeError"
	eventReasonSPFNeedsMerge      = "SPFNeedsMerge"
	eventReasonDomainCreate       = "DomainCreated"
	eventReasonDomainUpdate       = "DomainUpdated"
	eventReasonDomainDelete       = "DomainDeleted"
	eventReasonDomainStateChange  = "DomainStateChange"

	// dnsRequeueInterval is how long we wait before the next reconcile when
	// DNS records are not yet verified. Mailgun's verification cycle can
	// take several minutes for new DNS records to propagate, so requeueing
	// immediately just burns API quota.
	dnsRequeueInterval = 5 * time.Minute
)

// Setup adds a controller that reconciles Domain managed resources.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(v1beta1.DomainKind)
	recorder := event.NewAPIRecorder(mgr.GetEventRecorder(name))

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(v1beta1.DomainGroupVersionKind),
		managed.WithExternalConnector(&connector{
			kube:         mgr.GetClient(),
			usage:        resource.TrackerFn(func(ctx context.Context, mg resource.Managed) error { return nil }),
			newServiceFn: clients.NewClient,
			recorder:     recorder,
		}),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithPollInterval(o.PollInterval),
		managed.WithRecorder(recorder))

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&v1beta1.Domain{}).
		Complete(r)
}

// A connector is expected to produce an ExternalClient when its Connect method
// is called.
type connector struct {
	kube         client.Client
	usage        resource.Tracker
	newServiceFn func(config *clients.Config) clients.Client
	recorder     event.Recorder
}

// Connect typically produces an ExternalClient by:
// 1. Tracking that the managed resource is using a ProviderConfig.
// 2. Getting the managed resource's ProviderConfig.
// 3. Getting the credentials specified by the ProviderConfig.
// 4. Using the credentials to form a client.
func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*v1beta1.Domain)
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

	return &external{
		service:  svc,
		recorder: c.recorder,
		kube:     c.kube,
	}, nil
}

// An ExternalClient observes, then either creates, updates, or deletes an
// external resource to ensure it reflects the managed resource's desired state.
type external struct {
	service  clients.Client
	recorder event.Recorder
	kube     client.Client
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

// convertAPIRecordsToClient converts API-typed DNSRecord slices into the
// client-typed form expected by the ConfigMap builder and annotation
// helpers. The two structs share field types; we copy values verbatim
// because the API package owns the CRD type and must not be imported
// transitively into helper signatures that have nothing to do with CRD
// generation.
func convertAPIRecordsToClient(api []v1beta1.DNSRecord) []clients.DNSRecord {
	if api == nil {
		return nil
	}
	out := make([]clients.DNSRecord, len(api))
	for i, r := range api {
		out[i] = clients.DNSRecord{
			Name:     r.Name,
			Type:     r.Type,
			Value:    r.Value,
			Priority: r.Priority,
			Valid:    r.Valid,
		}
	}
	return out
}

// formatAnnotationBlock renders a stable annotation map as a multi-line
// string for inclusion in event messages. Keys appear in sorted order so
// the output is reproducible across reconciles (and so `kubectl describe`
// diffs are meaningful).
func formatAnnotationBlock(annotations map[string]string) string {
	if len(annotations) == 0 {
		return "  (no annotations)"
	}
	keys := make([]string, 0, len(annotations))
	for k := range annotations {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "  %s=%s\n", k, annotations[k])
	}
	return b.String()
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
	cr, ok := mg.(*v1beta1.Domain)
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

	// If DNS records are not yet verified, ask Mailgun to re-check them now so
	// the next reconcile sees a fresh validity reading. We do this only on the
	// slow path (records still unknown or state not active) to avoid hitting
	// Mailgun's rate limits in the happy path.
	needsReverify := domain.State != "active" ||
		(domain.DNSVerified != nil && !*domain.DNSVerified)
	if needsReverify {
		if verified, vErr := c.service.VerifyDomain(ctx, cr.Spec.ForProvider.Name); vErr == nil && verified != nil {
			domain = verified
			if c.recorder != nil {
				c.recorder.Event(cr, event.Normal(eventReasonDNSReverify,
					"Triggered Mailgun DNS re-verification; results will reflect on the next reconcile"))
			}
		}
	}

	upToDate := isDomainUpToDate(domain, &cr.Spec.ForProvider)

	cr.Status.AtProvider = *domain
	meta.SetExternalName(cr, cr.Spec.ForProvider.Name)

	previousState := ""
	for _, cond := range cr.Status.Conditions {
		if cond.Reason == xpv1.ReasonAvailable {
			previousState = "active"
			break
		}
	}

	if domain.State == "active" {
		cr.SetConditions(xpv1.Available())
	} else {
		cr.SetConditions(xpv1.Creating())
	}

	setDNSVerifiedCondition(cr, domain.DNSVerified)

	if c.recorder != nil {
		// Emit event for state transition
		if domain.State != previousState && previousState != "" {
			c.recorder.Event(cr, event.Normal(eventReasonDomainStateChange,
				fmt.Sprintf("Domain state changed from %s to %s", previousState, domain.State)))
		}

		// Emit events for DNS verification status
		if domain.DNSVerified != nil && !*domain.DNSVerified {
			invalidRecords := getInvalidRecordNames(cr.Status.AtProvider.RequiredDNSRecords)
			c.recorder.Event(cr, event.Warning(eventReasonDNSInvalid,
				errors.Errorf("DNS records not properly configured: %s", invalidRecords)))
		} else if domain.DNSVerified != nil && *domain.DNSVerified {
			c.recorder.Event(cr, event.Normal(eventReasonDNSVerified,
				"All DNS records are properly configured"))
		}
	}

	obs := managed.ExternalObservation{
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
	}

	// When DNS records are not yet verified, slow the next reconcile down to
	// dnsRequeueInterval by setting the crossplane.io/poll-interval
	// annotation. Mailgun's verification cycle takes several minutes for new
	// DNS records to propagate, so the controller-runtime default poll
	// cadence would just burn API quota. The DNSReverifyRequested event above
	// ensures the user has a signal that a fresh check was triggered in the
	// meantime. Clear the annotation once verification succeeds so the default
	// cadence resumes.
	//
	// Also publish the DNS automation side effects on the slow path:
	//   - emit a DNSRecordsRequired event with per-record annotations so the
	//     user (or a Composition) can `kubectl describe domain` and see the
	//     exact value of every record Mailgun is waiting for;
	//   - write the external-dns hostname annotation unless opted out;
	//   - create/update the DNS-records ConfigMap if opted in.
	if domain.DNSVerified == nil || !*domain.DNSVerified {
		meta.AddAnnotations(cr, map[string]string{
			meta.AnnotationKeyPollInterval: dnsRequeueInterval.String(),
		})
		c.publishDNSRecordsRequired(ctx, cr, convertAPIRecordsToClient(cr.Status.AtProvider.RequiredDNSRecords))
	} else {
		if _, ok := cr.GetAnnotations()[meta.AnnotationKeyPollInterval]; ok {
			meta.AddAnnotations(cr, map[string]string{
				meta.AnnotationKeyPollInterval: "",
			})
		}
	}

	return obs, nil
}

// publishDNSRecordsRequired emits a Normal event describing what records
// Mailgun is waiting on. The event carries one annotation per record so
// `kubectl describe` shows the exact values in the message field.
//
// Side effects (external-dns annotation, ConfigMap, DNS probe) are
// wired here too when the user has opted in; each is best-effort and
// never blocks the reconcile if it fails — they only affect the
// diagnostic experience.
func (c *external) publishDNSRecordsRequired(ctx context.Context, cr *v1beta1.Domain, records []clients.DNSRecord) {
	if c.recorder == nil {
		return
	}

	annotations := DNSRecordAnnotations(records)

	if len(records) == 0 {
		c.recorder.Event(cr, event.Normal(eventReasonDNSRecordsRequired,
			"Mailgun returned no DNS records to configure; domain is still unverified"))
		return
	}

	c.recorder.Event(cr, event.Normal(eventReasonDNSRecordsRequired,
		fmt.Sprintf("Mailgun requires %d DNS record(s) to verify this domain.\n"+
			"Each record below appears as `<type>-<hash>=<expected value>` so you can pipe them straight into a Composition:\n%s",
			len(records), formatAnnotationBlock(annotations)),
	))

	if c.kube == nil {
		return
	}
	helper := newK8sExternalDNSHelper(c.kube)
	if err := helper.setExternalDNSHostname(ctx, cr, cr.Spec.ForProvider.Name); err != nil {
		log := ctrl.Log.WithName("domain")
		log.Info("failed to set external-dns hostname annotation",
			"domain", cr.Spec.ForProvider.Name, "error", err.Error())
	}
	if err := helper.ensureDNSRecordsConfigMap(ctx, cr, records); err != nil {
		log := ctrl.Log.WithName("domain")
		log.Info("failed to ensure DNS records ConfigMap",
			"domain", cr.Spec.ForProvider.Name, "error", err.Error())
	}

	// DNS probe: opt-in only, because it requires outbound DNS egress
	// from the cluster. We do not gate the event on success because
	// probe failures (timeouts, NXDOMAIN on a not-yet-propagated record)
	// are exactly what the user wants to see in the event stream.
	if cr.GetAnnotations()[v1beta1.AnnotationDNSProbeEnabled] == "true" {
		prober := NewMiekDNSProber()
		probeCtx, cancel := context.WithTimeout(ctx, prober.Timeout)
		defer cancel()
		results := ProbeAllRecords(probeCtx, prober, records)
		EmitProbeResults(c.recorder, cr, results)
	}
}

func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errNotDomain)
	}

	cr.SetConditions(xpv1.Creating())

	domain, err := c.service.CreateDomain(ctx, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalCreation{}, errors.Wrap(err, "failed to create domain")
	}

	meta.SetExternalName(cr, cr.Spec.ForProvider.Name)
	cr.Status.AtProvider = *domain

	if c.recorder != nil {
		c.recorder.Event(cr, event.Normal(eventReasonDomainCreate,
			fmt.Sprintf("Domain %s created in Mailgun", cr.Spec.ForProvider.Name)))
	}

	if domain.State == "active" {
		cr.SetConditions(xpv1.Available())
	} else {
		cr.SetConditions(xpv1.Creating())
	}

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
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errNotDomain)
	}

	domain, err := c.service.UpdateDomain(ctx, cr.Spec.ForProvider.Name, &cr.Spec.ForProvider)
	if err != nil {
		return managed.ExternalUpdate{}, errors.Wrap(err, "failed to update domain")
	}

	cr.Status.AtProvider = *domain

	if c.recorder != nil {
		c.recorder.Event(cr, event.Normal(eventReasonDomainUpdate,
			fmt.Sprintf("Domain %s updated in Mailgun", cr.Spec.ForProvider.Name)))
	}

	if domain.State == "active" {
		cr.SetConditions(xpv1.Available())
	} else {
		cr.SetConditions(xpv1.Creating())
	}

	return managed.ExternalUpdate{
		// Optionally return any details that may be required to connect to the
		// external resource. These will be stored as the connection secret.
		ConnectionDetails: managed.ConnectionDetails{
			"smtp_login":    []byte(domain.SMTPLogin),
			"smtp_password": []byte(domain.SMTPPassword),
		},
	}, nil
}

func (c *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*v1beta1.Domain)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errNotDomain)
	}

	cr.SetConditions(xpv1.Deleting())

	err := c.service.DeleteDomain(ctx, cr.Spec.ForProvider.Name)
	if err != nil && !clients.IsNotFound(err) {
		return managed.ExternalDelete{}, errors.Wrap(err, "failed to delete domain")
	}

	if c.recorder != nil {
		c.recorder.Event(cr, event.Normal(eventReasonDomainDelete,
			fmt.Sprintf("Domain %s deleted from Mailgun", cr.Spec.ForProvider.Name)))
	}

	return managed.ExternalDelete{}, nil
}

// isDomainUpToDate checks if the external resource is up to date
func isDomainUpToDate(domain *v1beta1.DomainObservation, desired *v1beta1.DomainParameters) bool {
	// Compare updatable fields only
	// Note: Most domain fields cannot be updated after creation in Mailgun
	// We only check the fields that can be modified

	// SpamAction is write-only and cannot be read back from Mailgun API
	// so we cannot compare it in the observation
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

// setDNSVerifiedCondition sets the DNSVerified condition based on DNS record validation
func setDNSVerifiedCondition(cr *v1beta1.Domain, dnsVerified *bool) {
	if dnsVerified == nil {
		cr.SetConditions(xpv1.Condition{
			Type:               conditionTypeDNSVerified,
			Status:             corev1.ConditionUnknown,
			LastTransitionTime: metav1.Now(),
			Reason:             "Unknown",
			Message:            "DNS verification status unknown",
		})
		return
	}

	if *dnsVerified {
		cr.SetConditions(xpv1.Condition{
			Type:               conditionTypeDNSVerified,
			Status:             corev1.ConditionTrue,
			LastTransitionTime: metav1.Now(),
			Reason:             "AllDNSRecordsValid",
			Message:            "All required DNS records are properly configured",
		})
	} else {
		invalidRecords := getInvalidRecordNames(cr.Status.AtProvider.RequiredDNSRecords)
		msg := "One or more required DNS records are not properly configured"
		if len(invalidRecords) > 0 {
			msg += ": " + invalidRecords
		}
		cr.SetConditions(xpv1.Condition{
			Type:               conditionTypeDNSVerified,
			Status:             corev1.ConditionFalse,
			LastTransitionTime: metav1.Now(),
			Reason:             "DNSRecordsNotValid",
			Message:            msg,
		})
	}
}

// getInvalidRecordNames returns a comma-separated list of invalid record names
func getInvalidRecordNames(records []v1beta1.DNSRecord) string {
	var invalid []string
	for _, r := range records {
		if !r.Valid.IsVerified() {
			invalid = append(invalid, r.Name+" ("+r.Type+")")
		}
	}
	return strings.Join(invalid, ", ")
}
