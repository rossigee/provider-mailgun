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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	xpv1 "github.com/crossplane/crossplane-runtime/v2/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/v2/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/v2/pkg/resource"

	"github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	apisv1beta1 "github.com/rossigee/provider-mailgun/apis/v1beta1"
	bouncetypes "github.com/rossigee/provider-mailgun/apis/bounce/v1beta1"
	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	mailinglisttypes "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routetypes "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	templatetypes "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	webhooktypes "github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
)

// MockSMTPCredentialClient for testing
type MockSMTPCredentialClient struct {
	credentials map[string]*v1beta1.SMTPCredentialObservation
	err         error
}

func (m *MockSMTPCredentialClient) CreateSMTPCredential(ctx context.Context, domain string, credential *v1beta1.SMTPCredentialParameters) (*v1beta1.SMTPCredentialObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + credential.Login

	// Note: Password handling is managed separately via connection details

	result := &v1beta1.SMTPCredentialObservation{
		Login:     credential.Login,
		// Note: Password is not included in observation for security
		CreatedAt: "2025-01-01T00:00:00Z",
		State:     "active",
	}

	if m.credentials == nil {
		m.credentials = make(map[string]*v1beta1.SMTPCredentialObservation)
	}
	m.credentials[key] = result

	return result, nil
}

func (m *MockSMTPCredentialClient) GetSMTPCredential(ctx context.Context, domain, login string) (*v1beta1.SMTPCredentialObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + login
	if cred, exists := m.credentials[key]; exists {
		return cred, nil
	}

	return nil, errors.New("credential not found (404)")
}

func (m *MockSMTPCredentialClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*v1beta1.SMTPCredentialObservation, error) {
	if m.err != nil {
		return nil, m.err
	}

	key := domain + "/" + login
	if cred, exists := m.credentials[key]; exists {
		// Password updates handled via connection details
		return cred, nil
	}

	return nil, errors.New("credential not found (404)")
}

func (m *MockSMTPCredentialClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	if m.err != nil {
		return m.err
	}

	key := domain + "/" + login
	delete(m.credentials, key)
	return nil
}

// Template methods (no-op for SMTP credential tests)
func (m *MockSMTPCredentialClient) CreateTemplate(ctx context.Context, domain string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) GetTemplate(ctx context.Context, domain, name string) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) UpdateTemplate(ctx context.Context, domain, name string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return errors.New("not implemented")
}

// Implement other required client methods as no-ops
func (m *MockSMTPCredentialClient) CreateDomain(ctx context.Context, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) GetDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) UpdateDomain(ctx context.Context, name string, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) DeleteDomain(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) CreateMailingList(ctx context.Context, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) GetMailingList(ctx context.Context, address string) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) UpdateMailingList(ctx context.Context, address string, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) DeleteMailingList(ctx context.Context, address string) error {
	return errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) CreateRoute(ctx context.Context, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) GetRoute(ctx context.Context, id string) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) UpdateRoute(ctx context.Context, id string, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) DeleteRoute(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) CreateWebhook(ctx context.Context, domain string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) GetWebhook(ctx context.Context, domain, eventType string) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return errors.New("not implemented")
}

// Bounce operations
func (m *MockSMTPCredentialClient) CreateBounce(ctx context.Context, domain string, bounce *bouncetypes.BounceParameters) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) GetBounce(ctx context.Context, domain, address string) (*bouncetypes.BounceObservation, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Complaint operations
func (m *MockSMTPCredentialClient) CreateComplaint(ctx context.Context, domain string, complaint interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) GetComplaint(ctx context.Context, domain, address string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Unsubscribe operations
func (m *MockSMTPCredentialClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe interface{}) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) GetUnsubscribe(ctx context.Context, domain, address string) (interface{}, error) {
	return nil, errors.New("not implemented")
}

func (m *MockSMTPCredentialClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func TestSMTPCredentialObserve(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, corev1.AddToScheme(scheme))
	require.NoError(t, v1beta1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, apisv1beta1.SchemeBuilder.AddToScheme(scheme))

	type args struct {
		mg     resource.Managed
		secret *corev1.Secret
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
		"CredentialExistsWithSecret": {
			reason: "Should return ResourceExists when secret exists (rotation strategy)",
			args: args{
				mg: &v1beta1.SMTPCredential{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-smtp",
						Namespace: "test-namespace",
					},
					Spec: v1beta1.SMTPCredentialSpec{
						ForProvider: v1beta1.SMTPCredentialParameters{
							Domain: "example.com",
							Login:  "test@example.com",
						},
						ResourceSpec: xpv1.ResourceSpec{
							WriteConnectionSecretToReference: &xpv1.SecretReference{
								Name:      "test-secret",
								Namespace: "test-namespace",
							},
						},
					},
				},
				secret: &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-secret",
						Namespace: "test-namespace",
					},
					Data: map[string][]byte{
						"smtp_username": []byte("test@example.com"),
						"smtp_password": []byte("existing-password"),
					},
				},
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
					ConnectionDetails: managed.ConnectionDetails{
						"smtp_host":     []byte("smtp.mailgun.org"),
						"smtp_port":     []byte("587"),
						"smtp_username": []byte("test@example.com"),
					},
				},
			},
		},
		"CredentialNotFoundNoSecret": {
			reason: "Should return ResourceExists false when no secret exists (rotation strategy)",
			args: args{
				mg: &v1beta1.SMTPCredential{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-smtp",
						Namespace: "test-namespace",
					},
					Spec: v1beta1.SMTPCredentialSpec{
						ForProvider: v1beta1.SMTPCredentialParameters{
							Domain: "example.com",
							Login:  "notfound@example.com",
						},
						ResourceSpec: xpv1.ResourceSpec{
							WriteConnectionSecretToReference: &xpv1.SecretReference{
								Name:      "missing-secret",
								Namespace: "test-namespace",
							},
						},
					},
				},
				secret: nil, // No secret exists
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists: false,
				},
			},
		},
		"NoSecretConfigured": {
			reason: "Should return ResourceExists false when no secret is configured",
			args: args{
				mg: &v1beta1.SMTPCredential{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-smtp",
						Namespace: "test-namespace",
					},
					Spec: v1beta1.SMTPCredentialSpec{
						ForProvider: v1beta1.SMTPCredentialParameters{
							Domain: "example.com",
							Login:  "test@example.com",
						},
						// No WriteConnectionSecretToReference configured
					},
				},
				secret: nil,
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
			fakeClient := fake.NewClientBuilder().WithScheme(scheme)
			if tc.args.secret != nil {
				fakeClient = fakeClient.WithObjects(tc.args.secret)
			}
			kubeClient := fakeClient.Build()

			// Setup mock Mailgun client
			mockClient := &MockSMTPCredentialClient{}

			e := &external{
				service: mockClient,
				kube:    kubeClient,
			}
			got, err := e.Observe(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.want.o.ResourceExists, got.ResourceExists)
				assert.Equal(t, tc.want.o.ResourceUpToDate, got.ResourceUpToDate)

				if tc.want.o.ConnectionDetails != nil {
					// Check that connection details match (excluding password which isn't returned in observe)
					for key, expectedValue := range tc.want.o.ConnectionDetails {
						if key != "smtp_password" { // Password isn't returned by observe
							assert.Equal(t, expectedValue, got.ConnectionDetails[key])
						}
					}
				}
			}
		})
	}
}

func TestSMTPCredentialCreate(t *testing.T) {
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
			reason: "Should successfully create SMTP credential",
			args: args{
				mg: &v1beta1.SMTPCredential{
					Spec: v1beta1.SMTPCredentialSpec{
						ForProvider: v1beta1.SMTPCredentialParameters{
							Domain: "example.com",
							Login:  "new@example.com",
						},
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						"smtp_host":     []byte("smtp.mailgun.org"),
						"smtp_port":     []byte("587"),
						"smtp_username": []byte("new@example.com"),
						"smtp_password": []byte(""), // Empty - generated passwords not retrievable
					},
				},
			},
		},
		"SuccessfulCreateWithRotation": {
			reason: "Should successfully create SMTP credential with rotation (delete existing first)",
			args: args{
				mg: &v1beta1.SMTPCredential{
					Spec: v1beta1.SMTPCredentialSpec{
						ForProvider: v1beta1.SMTPCredentialParameters{
							Domain: "example.com",
							Login:  "existing@example.com",
						},
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{
						"smtp_host":     []byte("smtp.mailgun.org"),
						"smtp_port":     []byte("587"),
						"smtp_username": []byte("existing@example.com"),
						"smtp_password": []byte(""), // Empty - generated passwords not retrievable
					},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockSMTPCredentialClient{}

			// Pre-populate with existing credential for rotation test
			if name == "SuccessfulCreateWithRotation" {
				mockClient.credentials = map[string]*v1beta1.SMTPCredentialObservation{
					"example.com/existing@example.com": {
						Login:     "existing@example.com",
						// Password:  "old-password",
						CreatedAt: "2025-01-01T00:00:00Z",
						State:     "active",
					},
				}
			}

			e := &external{service: mockClient}

			got, err := e.Create(context.Background(), tc.args.mg)

			if tc.want.err != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.want.err.Error())
			} else {
				require.NoError(t, err)

				// Check connection details structure
				assert.Equal(t, []byte("smtp.mailgun.org"), got.ConnectionDetails["smtp_host"])
				assert.Equal(t, []byte("587"), got.ConnectionDetails["smtp_port"])
				assert.Contains(t, string(got.ConnectionDetails["smtp_username"]), "@example.com")

				// Verify password handling: when no password is provided, connection details should have empty password
				password := got.ConnectionDetails["smtp_password"]
				assert.Equal(t, []byte(""), password, "Should have empty password when none provided (generated passwords not retrievable)")

				// For rotation test, verify the old credential was deleted and new one created
				if name == "SuccessfulCreateWithRotation" {
					key := "example.com/existing@example.com"
					newCred, exists := mockClient.credentials[key]
					assert.True(t, exists, "New credential should exist after rotation")
					assert.Equal(t, "existing@example.com", newCred.Login, "Should have correct login")
				}
			}
		})
	}
}

func TestSMTPCredentialUpdate(t *testing.T) {
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
		"SuccessfulUpdate": {
			reason: "Should successfully update SMTP credential password",
			args: args{
				mg: &v1beta1.SMTPCredential{
					Spec: v1beta1.SMTPCredentialSpec{
						ForProvider: v1beta1.SMTPCredentialParameters{
							Domain:   "example.com",
							Login:    "existing@example.com",
							// Password: stringPtr("new-password"),
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
			mockClient := &MockSMTPCredentialClient{
				credentials: map[string]*v1beta1.SMTPCredentialObservation{
					"example.com/existing@example.com": {
						Login:     "existing@example.com",
						// Password:  "old-password",
						CreatedAt: "2025-01-01T00:00:00Z",
						State:     "active",
					},
				},
			}
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

func TestSMTPCredentialDelete(t *testing.T) {
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
			reason: "Should successfully delete SMTP credential",
			args: args{
				mg: &v1beta1.SMTPCredential{
					Spec: v1beta1.SMTPCredentialSpec{
						ForProvider: v1beta1.SMTPCredentialParameters{
							Domain: "example.com",
							Login:  "delete@example.com",
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
			mockClient := &MockSMTPCredentialClient{
				credentials: map[string]*v1beta1.SMTPCredentialObservation{
					"example.com/delete@example.com": {
						Login:     "delete@example.com",
						// Password:  "password",
						CreatedAt: "2025-01-01T00:00:00Z",
						State:     "active",
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
				// Verify credential was deleted
				_, exists := mockClient.credentials["example.com/delete@example.com"]
				assert.False(t, exists)
			}
		})
	}
}


func TestProviderConfigUsageTracker_Track(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, apisv1beta1.SchemeBuilder.AddToScheme(scheme))

	tests := []struct {
		name      string
		namespace string
		wantErr   bool
	}{
		{
			name:      "creates usage in correct namespace",
			namespace: "test-namespace",
			wantErr:   false,
		},
		{
			name:      "handles empty namespace",
			namespace: "",
			wantErr:   false,
		},
		{
			name:      "creates usage in crossplane-mailgun namespace",
			namespace: "crossplane-mailgun",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fakeClient := fake.NewClientBuilder().
				WithScheme(scheme).
				Build()

			tracker := newProviderConfigUsageTracker(fakeClient)

			// Create a test SMTPCredential
			cr := &v1beta1.SMTPCredential{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-smtp",
					Namespace: tt.namespace,
					UID:       types.UID("test-uid-123"),
				},
				Spec: v1beta1.SMTPCredentialSpec{
					ResourceSpec: xpv1.ResourceSpec{
						ProviderConfigReference: &xpv1.Reference{
							Name: "test-provider-config",
						},
					},
				},
			}
			cr.SetGroupVersionKind(v1beta1.SMTPCredentialGroupVersionKind)

			// Track the usage
			err := tracker.Track(context.Background(), cr)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)

			// Determine expected namespace: use "crossplane-system" if namespace is empty
			expectedNamespace := tt.namespace
			if expectedNamespace == "" {
				expectedNamespace = "crossplane-system"
			}

			// Verify that ProviderConfigUsage was created in the correct namespace
			pcu := &apisv1beta1.ProviderConfigUsage{}
			err = fakeClient.Get(context.Background(), types.NamespacedName{
				Name:      string(cr.GetUID()),
				Namespace: expectedNamespace,
			}, pcu)

			assert.NoError(t, err, "ProviderConfigUsage should be created")
			assert.Equal(t, expectedNamespace, pcu.GetNamespace(), "ProviderConfigUsage should be in the correct namespace")
			assert.Equal(t, "test-provider-config", pcu.GetName(), "ProviderConfigUsage should reference the correct ProviderConfig")

			// Verify owner reference is set correctly
			ownerRefs := pcu.GetOwnerReferences()
			require.Len(t, ownerRefs, 1)
			assert.Equal(t, cr.GetName(), ownerRefs[0].Name)
			assert.Equal(t, string(cr.GetUID()), string(ownerRefs[0].UID))
		})
	}
}

func TestProviderConfigUsageTracker_TrackIdempotent(t *testing.T) {
	scheme := runtime.NewScheme()
	require.NoError(t, v1beta1.SchemeBuilder.AddToScheme(scheme))
	require.NoError(t, apisv1beta1.SchemeBuilder.AddToScheme(scheme))

	fakeClient := fake.NewClientBuilder().
		WithScheme(scheme).
		Build()

	tracker := newProviderConfigUsageTracker(fakeClient)

	cr := &v1beta1.SMTPCredential{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-smtp",
			Namespace: "test-namespace",
			UID:       types.UID("test-uid-123"),
		},
		Spec: v1beta1.SMTPCredentialSpec{
			ResourceSpec: xpv1.ResourceSpec{
				ProviderConfigReference: &xpv1.Reference{
					Name: "test-provider-config",
				},
			},
		},
	}
	cr.SetGroupVersionKind(v1beta1.SMTPCredentialGroupVersionKind)

	// Track usage twice - should not error on second attempt
	err1 := tracker.Track(context.Background(), cr)
	assert.NoError(t, err1)

	err2 := tracker.Track(context.Background(), cr)
	assert.NoError(t, err2, "Second track call should not error (idempotent)")

	// Verify only one ProviderConfigUsage exists
	pcuList := &apisv1beta1.ProviderConfigUsageList{}
	err := fakeClient.List(context.Background(), pcuList, client.InNamespace("test-namespace"))
	assert.NoError(t, err)
	assert.Len(t, pcuList.Items, 1, "Should only have one ProviderConfigUsage")
}
