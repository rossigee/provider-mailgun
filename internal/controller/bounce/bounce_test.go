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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-mailgun/apis/bounce/v1alpha1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	"github.com/rossigee/provider-mailgun/internal/clients"
)

// MockBounceClient for testing
type MockBounceClient struct {
	bounces map[string]*clients.Bounce
	err     error
}

func (m *MockBounceClient) CreateBounce(ctx context.Context, domain string, bounce *clients.BounceSpec) (*clients.Bounce, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + bounce.Address
	result := &clients.Bounce{
		Address:   bounce.Address,
		Code:      stringPtrValue(bounce.Code),
		Error:     stringPtrValue(bounce.Error),
		CreatedAt: "2025-01-01T00:00:00Z",
	}

	if m.bounces == nil {
		m.bounces = make(map[string]*clients.Bounce)
	}
	m.bounces[key] = result

	return result, nil
}

func (m *MockBounceClient) GetBounce(ctx context.Context, domain, address string) (*clients.Bounce, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + address
	if bounce, exists := m.bounces[key]; exists {
		return bounce, nil
	}

	return nil, errors.New("bounce not found (404)")
}

func (m *MockBounceClient) DeleteBounce(ctx context.Context, domain, address string) error {
	if m.err != nil {
		return m.err
	}

	key := domain + "/" + address
	delete(m.bounces, key)
	return nil
}

// Implement other required client methods as no-ops
func (m *MockBounceClient) CreateSMTPCredential(ctx context.Context, domain string, credential *clients.SMTPCredentialSpec) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) GetSMTPCredential(ctx context.Context, domain, login string) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return errors.New("not implemented")
}

func (m *MockBounceClient) CreateTemplate(ctx context.Context, domain string, template *clients.TemplateSpec) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) GetTemplate(ctx context.Context, domain, name string) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) UpdateTemplate(ctx context.Context, domain, name string, template *clients.TemplateSpec) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return errors.New("not implemented")
}

func (m *MockBounceClient) CreateDomain(ctx context.Context, domain *clients.DomainSpec) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) GetDomain(ctx context.Context, name string) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) UpdateDomain(ctx context.Context, name string, domain *clients.DomainSpec) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) DeleteDomain(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *MockBounceClient) CreateMailingList(ctx context.Context, list *clients.MailingListSpec) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) GetMailingList(ctx context.Context, address string) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) UpdateMailingList(ctx context.Context, address string, list *clients.MailingListSpec) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) DeleteMailingList(ctx context.Context, address string) error {
	return errors.New("not implemented")
}

func (m *MockBounceClient) CreateRoute(ctx context.Context, route *clients.RouteSpec) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) GetRoute(ctx context.Context, id string) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) UpdateRoute(ctx context.Context, id string, route *clients.RouteSpec) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) DeleteRoute(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *MockBounceClient) CreateWebhook(ctx context.Context, domain string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) GetWebhook(ctx context.Context, domain, eventType string) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return errors.New("not implemented")
}

func (m *MockBounceClient) CreateComplaint(ctx context.Context, domain string, complaint *clients.ComplaintSpec) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) GetComplaint(ctx context.Context, domain, address string) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func (m *MockBounceClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe *clients.UnsubscribeSpec) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) GetUnsubscribe(ctx context.Context, domain, address string) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockBounceClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func TestBounceObserve(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1alpha1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, apisv1beta1.SchemeBuilder.AddToScheme(scheme))

	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalObservation
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"BounceExists": {
			reason: "Should return ResourceExists when bounce exists",
			args: args{
				mg: &v1alpha1.Bounce{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-bounce",
						Namespace: "test-namespace",
						Annotations: map[string]string{
							"crossplane.io/external-name": "bounce@example.com",
						},
					},
					Spec: v1alpha1.BounceSpec{
						ForProvider: v1alpha1.BounceParameters{
							Address: "bounce@example.com",
							Code:    stringPtr("550"),
							Error:   stringPtr("User unknown"),
							DomainRef: xpv1.Reference{
								Name: "example.com",
							},
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"BounceNotFound": {
			reason: "Should return ResourceExists false when bounce not found",
			args: args{
				mg: &v1alpha1.Bounce{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-bounce",
						Namespace: "test-namespace",
					},
					Spec: v1alpha1.BounceSpec{
						ForProvider: v1alpha1.BounceParameters{
							Address: "notfound@example.com",
							Code:    stringPtr("550"),
							Error:   stringPtr("User unknown"),
							DomainRef: xpv1.Reference{
								Name: "example.com",
							},
						},
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// Setup fake Kubernetes client
			fakeClient := fake.NewClientBuilder().WithScheme(scheme).Build()

			// Setup mock Mailgun client
			mockClient := &MockBounceClient{}

			// Pre-populate mock client for "exists" test
			if name == "BounceExists" {
				mockClient.bounces = map[string]*clients.Bounce{
					"example.com/bounce@example.com": {
						Address:   "bounce@example.com",
						Code:      "550",
						Error:     "User unknown",
						CreatedAt: "2025-01-01T00:00:00Z",
					},
				}
			}

			e := &external{
				service: mockClient,
				kube:    fakeClient,
			}
			got, err := e.Observe(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.o.ResourceExists, got.ResourceExists)
				assert.Equal(t, tc.want.o.ResourceUpToDate, got.ResourceUpToDate)
			}
		})
	}
}

func TestBounceCreate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalCreation
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulCreate": {
			reason: "Should successfully create bounce",
			args: args{
				mg: &v1alpha1.Bounce{
					Spec: v1alpha1.BounceSpec{
						ForProvider: v1alpha1.BounceParameters{
							Address: "new@example.com",
							Code:    stringPtr("550"),
							Error:   stringPtr("User unknown"),
							DomainRef: xpv1.Reference{
								Name: "example.com",
							},
						},
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockBounceClient{}
			e := &external{service: mockClient}

			got, err := e.Create(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.o, got)

				// Verify bounce was created in mock
				key := "example.com/new@example.com"
				bounce, exists := mockClient.bounces[key]
				assert.True(t, exists, "Bounce should be created")
				assert.Equal(t, "new@example.com", bounce.Address)
				assert.Equal(t, "550", bounce.Code)
				assert.Equal(t, "User unknown", bounce.Error)
			}
		})
	}
}

func TestBounceUpdate(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		o   managed.ExternalUpdate
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"UpdateNotSupported": {
			reason: "Update should be no-op for bounces",
			args: args{
				mg: &v1alpha1.Bounce{
					Spec: v1alpha1.BounceSpec{
						ForProvider: v1alpha1.BounceParameters{
							Address: "existing@example.com",
							Code:    stringPtr("550"),
							Error:   stringPtr("Updated error"),
							DomainRef: xpv1.Reference{
								Name: "example.com",
							},
						},
					},
				},
			},
			want: want{
				o: managed.ExternalUpdate{},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockBounceClient{}
			e := &external{service: mockClient}

			got, err := e.Update(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.o, got)
			}
		})
	}
}

func TestBounceDelete(t *testing.T) {
	type args struct {
		mg resource.Managed
	}
	type want struct {
		err error
	}

	cases := map[string]struct {
		reason string
		args   args
		want   want
	}{
		"SuccessfulDelete": {
			reason: "Should successfully delete bounce",
			args: args{
				mg: &v1alpha1.Bounce{
					ObjectMeta: metav1.ObjectMeta{
						Annotations: map[string]string{
							"crossplane.io/external-name": "delete@example.com",
						},
					},
					Spec: v1alpha1.BounceSpec{
						ForProvider: v1alpha1.BounceParameters{
							Address: "delete@example.com",
							Code:    stringPtr("550"),
							Error:   stringPtr("User unknown"),
							DomainRef: xpv1.Reference{
								Name: "example.com",
							},
						},
					},
				},
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockBounceClient{
				bounces: map[string]*clients.Bounce{
					"example.com/delete@example.com": {
						Address:   "delete@example.com",
						Code:      "550",
						Error:     "User unknown",
						CreatedAt: "2025-01-01T00:00:00Z",
					},
				},
			}
			e := &external{service: mockClient}

			_, err := e.Delete(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				// Verify bounce was deleted
				_, exists := mockClient.bounces["example.com/delete@example.com"]
				assert.False(t, exists)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func stringPtrValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func TestResolveDomainName(t *testing.T) {
	cases := map[string]struct {
		domainRefName string
		expected      string
	}{
		"DomainWithDots": {
			domainRefName: "example.com",
			expected:      "example.com",
		},
		"DomainWithoutDots": {
			domainRefName: "example-domain",
			expected:      "example-domain",
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			e := &external{}
			cr := &v1alpha1.Bounce{
				Spec: v1alpha1.BounceSpec{
					ForProvider: v1alpha1.BounceParameters{
						DomainRef: xpv1.Reference{
							Name: tc.domainRefName,
						},
					},
				},
			}

			result, err := e.resolveDomainName(context.Background(), cr)
			require.NoError(t, err)
			assert.Equal(t, tc.expected, result)
		})
	}
}
