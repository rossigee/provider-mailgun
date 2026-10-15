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

package unsubscribe

import (
	"context"
	"net/http"
	"testing"

	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"
	xpv1 "github.com/crossplane/crossplane/apis/v2/core/v2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	bouncetypes "github.com/rossigee/provider-mailgun/apis/bounce/v1beta1"
	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	mailinglisttypes "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routetypes "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpcredentialtypes "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatetypes "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	v1beta1 "github.com/rossigee/provider-mailgun/apis/unsubscribe/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	webhooktypes "github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
	"github.com/rossigee/provider-mailgun/internal/clients"
)

// MockUnsubscribeClient for testing
type MockUnsubscribeClient struct {
	unsubscribes   map[string]*clients.Unsubscribe
	err            error
	deleteAttempts []string
}

func (m *MockUnsubscribeClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe interface{}) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}

	spec, ok := unsubscribe.(*clients.UnsubscribeSpec)
	if !ok {
		return nil, errors.New("invalid unsubscribe parameter type")
	}

	if m.unsubscribes == nil {
		m.unsubscribes = make(map[string]*clients.Unsubscribe)
	}
	result := &clients.Unsubscribe{
		Address:   spec.Address,
		CreatedAt: "2025-01-01T00:00:00Z",
	}
	if spec.Tags != nil {
		result.Tags = *spec.Tags
	}
	m.unsubscribes[domain+"/"+spec.Address] = result

	return result, nil
}

func (m *MockUnsubscribeClient) GetUnsubscribe(ctx context.Context, domain, address string) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}

	if unsubscribe, exists := m.unsubscribes[domain+"/"+address]; exists {
		return unsubscribe, nil
	}

	return nil, &clients.APIError{StatusCode: http.StatusNotFound, Message: "unsubscribe not found"}
}

func (m *MockUnsubscribeClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	if m.err != nil {
		return m.err
	}

	key := domain + "/" + address
	m.deleteAttempts = append(m.deleteAttempts, key)
	if _, exists := m.unsubscribes[key]; !exists {
		return &clients.APIError{StatusCode: http.StatusNotFound, Message: "unsubscribe not found"}
	}
	delete(m.unsubscribes, key)
	return nil
}

// Implement other required client methods as no-ops with v1beta1 types

// Domain operations
func (m *MockUnsubscribeClient) CreateDomain(ctx context.Context, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) GetDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) UpdateDomain(ctx context.Context, name string, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) DeleteDomain(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *MockUnsubscribeClient) VerifyDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

// MailingList operations
func (m *MockUnsubscribeClient) CreateMailingList(ctx context.Context, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) GetMailingList(ctx context.Context, address string) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) UpdateMailingList(ctx context.Context, address string, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) DeleteMailingList(ctx context.Context, address string) error {
	return errors.New("not implemented")
}

// Route operations
func (m *MockUnsubscribeClient) CreateRoute(ctx context.Context, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) GetRoute(ctx context.Context, id string) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) UpdateRoute(ctx context.Context, id string, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) DeleteRoute(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

// Webhook operations
func (m *MockUnsubscribeClient) CreateWebhook(ctx context.Context, domain string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) GetWebhook(ctx context.Context, domain, eventType string) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return errors.New("not implemented")
}

// SMTPCredential operations
func (m *MockUnsubscribeClient) CreateSMTPCredential(ctx context.Context, domain string, credential *smtpcredentialtypes.SMTPCredentialParameters) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) GetSMTPCredential(ctx context.Context, domain, login string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return errors.New("not implemented")
}

// Template operations
func (m *MockUnsubscribeClient) CreateTemplate(ctx context.Context, domain string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) GetTemplate(ctx context.Context, domain, name string) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) UpdateTemplate(ctx context.Context, domain, name string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return errors.New("not implemented")
}

// Bounce operations
func (m *MockUnsubscribeClient) CreateBounce(ctx context.Context, domain string, bounce *bouncetypes.BounceParameters) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) GetBounce(ctx context.Context, domain, address string) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Complaint operations
func (m *MockUnsubscribeClient) CreateComplaint(ctx context.Context, domain string, complaint interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) GetComplaint(ctx context.Context, domain, address string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockUnsubscribeClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func (m *MockUnsubscribeClient) SendEmail(ctx context.Context, domain, from, to, subject, body string) error {
	return nil
}

func newUnsubscribe(address, domainRef string, tags *string) *v1beta1.Unsubscribe {
	return &v1beta1.Unsubscribe{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-unsubscribe",
			Namespace: "test-namespace",
		},
		Spec: v1beta1.UnsubscribeSpec{
			ForProvider: v1beta1.UnsubscribeParameters{
				Address: address,
				Tags:    tags,
				DomainRef: xpv1.Reference{
					Name: domainRef,
				},
			},
		},
	}
}

func TestUnsubscribeObserve(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, apisv1beta1.SchemeBuilder.AddToScheme(scheme))

	cases := map[string]struct {
		reason  string
		mg      resource.Managed
		mockErr error
		seed    map[string]*clients.Unsubscribe
		want    managed.ExternalObservation
		wantErr string
	}{
		"UnsubscribeExists": {
			reason: "Should return ResourceExists and populate status when the unsubscribe exists",
			mg:     newUnsubscribe("unsub@example.com", "example.com", nil),
			seed: map[string]*clients.Unsubscribe{
				"example.com/unsub@example.com": {
					Address:   "unsub@example.com",
					CreatedAt: "2025-01-01T00:00:00Z",
				},
			},
			want: managed.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: true,
			},
		},
		"UnsubscribeNotFound": {
			reason: "Should return ResourceExists false when the unsubscribe is absent",
			mg:     newUnsubscribe("missing@example.com", "example.com", nil),
			want: managed.ExternalObservation{
				ResourceExists: false,
			},
		},
		"ServiceError": {
			reason:  "Should surface service errors",
			mg:      newUnsubscribe("unsub@example.com", "example.com", nil),
			mockErr: errors.New("boom"),
			wantErr: "cannot get unsubscribe",
		},
		"WrongManagedResource": {
			reason:  "Should reject a managed resource of the wrong type",
			mg:      &bouncetypes.Bounce{},
			wantErr: errNotUnsubscribe,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			mockClient := &MockUnsubscribeClient{unsubscribes: tc.seed, err: tc.mockErr}

			e := &external{service: mockClient, kube: fakeClient}
			got, err := e.Observe(context.Background(), tc.mg)

			if tc.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want.ResourceExists, got.ResourceExists)
			assert.Equal(t, tc.want.ResourceUpToDate, got.ResourceUpToDate)

			if unsubscribe, ok := tc.mg.(*v1beta1.Unsubscribe); ok && tc.want.ResourceExists {
				require.NotNil(t, unsubscribe.Status.AtProvider.CreatedAt)
				assert.Equal(t, "2025-01-01T00:00:00Z", *unsubscribe.Status.AtProvider.CreatedAt)
				assert.Equal(t, "unsub@example.com", unsubscribe.GetAnnotations()["crossplane.io/external-name"])
			}
		})
	}
}

func TestUnsubscribeCreate(t *testing.T) {
	tags := "newsletter,promo"
	mockClient := &MockUnsubscribeClient{}
	e := &external{service: mockClient}
	cr := newUnsubscribe("new@example.com", "example.com", &tags)

	got, err := e.Create(context.Background(), cr)
	require.NoError(t, err)
	assert.Equal(t, managed.ExternalCreation{}, got)

	created, exists := mockClient.unsubscribes["example.com/new@example.com"]
	require.True(t, exists, "unsubscribe should be created in the mock")
	assert.Equal(t, "new@example.com", created.Address)
	assert.Equal(t, tags, created.Tags)
	assert.Equal(t, "new@example.com", cr.GetAnnotations()["crossplane.io/external-name"])
}

func TestUnsubscribeCreateError(t *testing.T) {
	mockClient := &MockUnsubscribeClient{err: errors.New("boom")}
	e := &external{service: mockClient}

	_, err := e.Create(context.Background(), newUnsubscribe("new@example.com", "example.com", nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot create unsubscribe")
}

func TestUnsubscribeUpdate(t *testing.T) {
	e := &external{service: &MockUnsubscribeClient{}}

	got, err := e.Update(context.Background(), newUnsubscribe("existing@example.com", "example.com", nil))
	require.NoError(t, err)
	assert.Equal(t, managed.ExternalUpdate{}, got)
}

func TestUnsubscribeDelete(t *testing.T) {
	cases := map[string]struct {
		reason   string
		seed     map[string]*clients.Unsubscribe
		external string
	}{
		"SuccessfulDelete": {
			reason: "Should delete an existing unsubscribe",
			seed: map[string]*clients.Unsubscribe{
				"example.com/delete@example.com": {Address: "delete@example.com"},
			},
			external: "delete@example.com",
		},
		"MissingIsTolerated": {
			reason:   "A 404 from Mailgun should be treated as already deleted",
			external: "absent@example.com",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockUnsubscribeClient{unsubscribes: tc.seed}
			e := &external{service: mockClient}
			cr := newUnsubscribe(tc.external, "example.com", nil)
			cr.SetAnnotations(map[string]string{"crossplane.io/external-name": tc.external})

			_, err := e.Delete(context.Background(), cr)
			require.NoError(t, err)

			assert.Equal(t, []string{"example.com/" + tc.external}, mockClient.deleteAttempts)
			_, exists := mockClient.unsubscribes["example.com/"+tc.external]
			assert.False(t, exists)
		})
	}
}

func TestUnsubscribeDeleteError(t *testing.T) {
	mockClient := &MockUnsubscribeClient{err: errors.New("boom")}
	e := &external{service: mockClient}

	_, err := e.Delete(context.Background(), newUnsubscribe("delete@example.com", "example.com", nil))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete unsubscribe")
}

func TestUnsubscribeResolveDomainName(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, domaintypes.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, apisv1beta1.SchemeBuilder.AddToScheme(scheme))

	cases := map[string]struct {
		reason        string
		domainRefName string
		domainName    string
		namespace     string
		expected      string
	}{
		"RefIsAlreadyADomainName": {
			reason:        "A ref containing dots is used verbatim, without a lookup",
			domainRefName: "example.com",
			expected:      "example.com",
		},
		"ResolvesDomainResourceInNamespace": {
			reason:        "A non-dotted ref is resolved through the Domain resource's spec",
			domainRefName: "example-domain",
			domainName:    "example-domain",
			namespace:     "test-namespace",
			expected:      "actual.example.com",
		},
		"FallsBackWhenDomainResourceMissing": {
			reason:        "A missing Domain resource falls back to the ref name",
			domainRefName: "example-domain",
			expected:      "example-domain",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			builder := fake.NewClientBuilder().WithScheme(scheme)
			if tc.domainName != "" {
				builder = builder.WithObjects(&domaintypes.Domain{
					ObjectMeta: metav1.ObjectMeta{
						Name:      tc.domainName,
						Namespace: tc.namespace,
					},
					Spec: domaintypes.DomainSpec{
						ForProvider: domaintypes.DomainParameters{
							Name: "actual.example.com",
						},
					},
				})
			}
			fakeClient := builder.Build()

			e := &external{kube: fakeClient}
			cr := newUnsubscribe("unsub@example.com", tc.domainRefName, nil)
			if tc.namespace != "" {
				cr.Namespace = tc.namespace
			}

			result, err := e.resolveDomainName(context.Background(), cr)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
