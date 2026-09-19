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

package complaint

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
	v1beta1 "github.com/rossigee/provider-mailgun/apis/complaint/v1beta1"
	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	mailinglisttypes "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routetypes "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpcredentialtypes "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatetypes "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	webhooktypes "github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
	"github.com/rossigee/provider-mailgun/internal/clients"
)

// MockComplaintClient for testing
type MockComplaintClient struct {
	complaints     map[string]*clients.Complaint
	err            error
	deleteAttempts []string
}

func (m *MockComplaintClient) CreateComplaint(ctx context.Context, domain string, complaint interface{}) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}

	spec, ok := complaint.(*clients.ComplaintSpec)
	if !ok {
		return nil, errors.New("invalid complaint parameter type")
	}

	if m.complaints == nil {
		m.complaints = make(map[string]*clients.Complaint)
	}
	result := &clients.Complaint{
		Address:   spec.Address,
		CreatedAt: "2025-01-01T00:00:00Z",
	}
	m.complaints[domain+"/"+spec.Address] = result

	return result, nil
}

func (m *MockComplaintClient) GetComplaint(ctx context.Context, domain, address string) (interface{}, error) {
	if m.err != nil {
		return nil, m.err
	}

	if complaint, exists := m.complaints[domain+"/"+address]; exists {
		return complaint, nil
	}

	return nil, &clients.APIError{StatusCode: http.StatusNotFound, Message: "complaint not found"}
}

func (m *MockComplaintClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	if m.err != nil {
		return m.err
	}

	key := domain + "/" + address
	m.deleteAttempts = append(m.deleteAttempts, key)
	if _, exists := m.complaints[key]; !exists {
		return &clients.APIError{StatusCode: http.StatusNotFound, Message: "complaint not found"}
	}
	delete(m.complaints, key)
	return nil
}

// Implement other required client methods as no-ops with v1beta1 types

// Domain operations
func (m *MockComplaintClient) CreateDomain(ctx context.Context, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) GetDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) UpdateDomain(ctx context.Context, name string, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) DeleteDomain(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *MockComplaintClient) VerifyDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

// MailingList operations
func (m *MockComplaintClient) CreateMailingList(ctx context.Context, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) GetMailingList(ctx context.Context, address string) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) UpdateMailingList(ctx context.Context, address string, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) DeleteMailingList(ctx context.Context, address string) error {
	return errors.New("not implemented")
}

// Route operations
func (m *MockComplaintClient) CreateRoute(ctx context.Context, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) GetRoute(ctx context.Context, id string) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) UpdateRoute(ctx context.Context, id string, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) DeleteRoute(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

// Webhook operations
func (m *MockComplaintClient) CreateWebhook(ctx context.Context, domain string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) GetWebhook(ctx context.Context, domain, eventType string) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return errors.New("not implemented")
}

// SMTPCredential operations
func (m *MockComplaintClient) CreateSMTPCredential(ctx context.Context, domain string, credential *smtpcredentialtypes.SMTPCredentialParameters) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) GetSMTPCredential(ctx context.Context, domain, login string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return errors.New("not implemented")
}

// Template operations
func (m *MockComplaintClient) CreateTemplate(ctx context.Context, domain string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) GetTemplate(ctx context.Context, domain, name string) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) UpdateTemplate(ctx context.Context, domain, name string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return errors.New("not implemented")
}

// Bounce operations
func (m *MockComplaintClient) CreateBounce(ctx context.Context, domain string, bounce *bouncetypes.BounceParameters) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) GetBounce(ctx context.Context, domain, address string) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Unsubscribe operations
func (m *MockComplaintClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) GetUnsubscribe(ctx context.Context, domain, address string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockComplaintClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func (m *MockComplaintClient) SendEmail(ctx context.Context, domain, from, to, subject, body string) error {
	return nil
}

func newComplaint(address, domainRef string) *v1beta1.Complaint {
	return &v1beta1.Complaint{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-complaint",
			Namespace: "test-namespace",
		},
		Spec: v1beta1.ComplaintSpec{
			ForProvider: v1beta1.ComplaintParameters{
				Address: address,
				DomainRef: xpv1.Reference{
					Name: domainRef,
				},
			},
		},
	}
}

func TestComplaintObserve(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, apisv1beta1.SchemeBuilder.AddToScheme(scheme))

	cases := map[string]struct {
		reason  string
		mg      resource.Managed
		mockErr error
		seed    map[string]*clients.Complaint
		want    managed.ExternalObservation
		wantErr string
	}{
		"ComplaintExists": {
			reason: "Should return ResourceExists and populate status when the complaint exists",
			mg:     newComplaint("complaint@example.com", "example.com"),
			seed: map[string]*clients.Complaint{
				"example.com/complaint@example.com": {
					Address:   "complaint@example.com",
					CreatedAt: "2025-01-01T00:00:00Z",
				},
			},
			want: managed.ExternalObservation{
				ResourceExists:   true,
				ResourceUpToDate: true,
			},
		},
		"ComplaintNotFound": {
			reason: "Should return ResourceExists false when the complaint is absent",
			mg:     newComplaint("missing@example.com", "example.com"),
			want: managed.ExternalObservation{
				ResourceExists: false,
			},
		},
		"ServiceError": {
			reason:  "Should surface service errors",
			mg:      newComplaint("complaint@example.com", "example.com"),
			mockErr: errors.New("boom"),
			wantErr: "cannot get complaint",
		},
		"WrongManagedResource": {
			reason:  "Should reject a managed resource of the wrong type",
			mg:      &bouncetypes.Bounce{},
			wantErr: errNotComplaint,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()
			mockClient := &MockComplaintClient{complaints: tc.seed, err: tc.mockErr}

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

			if complaint, ok := tc.mg.(*v1beta1.Complaint); ok && tc.want.ResourceExists {
				require.NotNil(t, complaint.Status.AtProvider.CreatedAt)
				assert.Equal(t, "2025-01-01T00:00:00Z", *complaint.Status.AtProvider.CreatedAt)
				assert.Equal(t, "complaint@example.com", complaint.GetAnnotations()["crossplane.io/external-name"])
			}
		})
	}
}

func TestComplaintCreate(t *testing.T) {
	mockClient := &MockComplaintClient{}
	e := &external{service: mockClient}
	cr := newComplaint("new@example.com", "example.com")

	got, err := e.Create(context.Background(), cr)
	require.NoError(t, err)
	assert.Equal(t, managed.ExternalCreation{}, got)

	created, exists := mockClient.complaints["example.com/new@example.com"]
	require.True(t, exists, "complaint should be created in the mock")
	assert.Equal(t, "new@example.com", created.Address)
	assert.Equal(t, "new@example.com", cr.GetAnnotations()["crossplane.io/external-name"])
}

func TestComplaintCreateError(t *testing.T) {
	mockClient := &MockComplaintClient{err: errors.New("boom")}
	e := &external{service: mockClient}

	_, err := e.Create(context.Background(), newComplaint("new@example.com", "example.com"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot create complaint")
}

func TestComplaintUpdate(t *testing.T) {
	e := &external{service: &MockComplaintClient{}}

	got, err := e.Update(context.Background(), newComplaint("existing@example.com", "example.com"))
	require.NoError(t, err)
	assert.Equal(t, managed.ExternalUpdate{}, got)
}

func TestComplaintDelete(t *testing.T) {
	cases := map[string]struct {
		reason   string
		seed     map[string]*clients.Complaint
		external string
	}{
		"SuccessfulDelete": {
			reason: "Should delete an existing complaint",
			seed: map[string]*clients.Complaint{
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
			mockClient := &MockComplaintClient{complaints: tc.seed}
			e := &external{service: mockClient}
			cr := newComplaint(tc.external, "example.com")
			cr.SetAnnotations(map[string]string{"crossplane.io/external-name": tc.external})

			_, err := e.Delete(context.Background(), cr)
			require.NoError(t, err)

			assert.Equal(t, []string{"example.com/" + tc.external}, mockClient.deleteAttempts)
			_, exists := mockClient.complaints["example.com/"+tc.external]
			assert.False(t, exists)
		})
	}
}

func TestComplaintDeleteError(t *testing.T) {
	mockClient := &MockComplaintClient{err: errors.New("boom")}
	e := &external{service: mockClient}

	_, err := e.Delete(context.Background(), newComplaint("delete@example.com", "example.com"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot delete complaint")
}

func TestComplaintResolveDomainName(t *testing.T) {
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
			cr := newComplaint("complaint@example.com", tc.domainRefName)
			if tc.namespace != "" {
				cr.Namespace = tc.namespace
			}

			result, err := e.resolveDomainName(context.Background(), cr)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
