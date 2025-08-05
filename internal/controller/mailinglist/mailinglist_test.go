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
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-mailgun/apis/mailinglist/v1alpha1"
	"github.com/rossigee/provider-mailgun/internal/clients"
)

// MockMailingListClient for testing
type MockMailingListClient struct {
	mailingLists map[string]*clients.MailingList
	err          error
}

func (m *MockMailingListClient) CreateMailingList(ctx context.Context, list *clients.MailingListSpec) (*clients.MailingList, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := &clients.MailingList{
		Address:         list.Address,
		Name:            "Test List",
		Description:     "Test mailing list",
		AccessLevel:     "readonly",
		ReplyPreference: "list",
		CreatedAt:       "2025-01-01T00:00:00Z",
		MembersCount:    0,
	}

	if list.Name != nil {
		result.Name = *list.Name
	}
	if list.Description != nil {
		result.Description = *list.Description
	}
	if list.AccessLevel != nil {
		result.AccessLevel = *list.AccessLevel
	}
	if list.ReplyPreference != nil {
		result.ReplyPreference = *list.ReplyPreference
	}

	if m.mailingLists == nil {
		m.mailingLists = make(map[string]*clients.MailingList)
	}
	m.mailingLists[list.Address] = result

	return result, nil
}

func (m *MockMailingListClient) GetMailingList(ctx context.Context, address string) (*clients.MailingList, error) {
	if m.err != nil {
		return nil, m.err
	}

	if list, exists := m.mailingLists[address]; exists {
		return list, nil
	}

	return nil, errors.New("mailing list not found (404)")
}

func (m *MockMailingListClient) UpdateMailingList(ctx context.Context, address string, list *clients.MailingListSpec) (*clients.MailingList, error) {
	if m.err != nil {
		return nil, m.err
	}

	if existing, exists := m.mailingLists[address]; exists {
		// Update modifiable fields
		if list.Name != nil {
			existing.Name = *list.Name
		}
		if list.Description != nil {
			existing.Description = *list.Description
		}
		if list.AccessLevel != nil {
			existing.AccessLevel = *list.AccessLevel
		}
		if list.ReplyPreference != nil {
			existing.ReplyPreference = *list.ReplyPreference
		}
		return existing, nil
	}

	return nil, errors.New("mailing list not found (404)")
}

func (m *MockMailingListClient) DeleteMailingList(ctx context.Context, address string) error {
	if m.err != nil {
		return m.err
	}

	delete(m.mailingLists, address)
	return nil
}

// Implement other required client methods as no-ops
func (m *MockMailingListClient) CreateDomain(ctx context.Context, domain *clients.DomainSpec) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) GetDomain(ctx context.Context, name string) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) UpdateDomain(ctx context.Context, name string, domain *clients.DomainSpec) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) DeleteDomain(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *MockMailingListClient) CreateRoute(ctx context.Context, route *clients.RouteSpec) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) GetRoute(ctx context.Context, id string) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) UpdateRoute(ctx context.Context, id string, route *clients.RouteSpec) (*clients.Route, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) DeleteRoute(ctx context.Context, id string) error {
	return errors.New("not implemented")
}

func (m *MockMailingListClient) CreateWebhook(ctx context.Context, domain string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) GetWebhook(ctx context.Context, domain, eventType string) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return errors.New("not implemented")
}

func (m *MockMailingListClient) CreateSMTPCredential(ctx context.Context, domain string, credential *clients.SMTPCredentialSpec) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) GetSMTPCredential(ctx context.Context, domain, login string) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return errors.New("not implemented")
}

func (m *MockMailingListClient) CreateTemplate(ctx context.Context, domain string, template *clients.TemplateSpec) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) GetTemplate(ctx context.Context, domain, name string) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) UpdateTemplate(ctx context.Context, domain, name string, template *clients.TemplateSpec) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return errors.New("not implemented")
}

// Bounce suppression operations
func (m *MockMailingListClient) CreateBounce(ctx context.Context, domain string, bounce *clients.BounceSpec) (*clients.Bounce, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) GetBounce(ctx context.Context, domain, address string) (*clients.Bounce, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Complaint suppression operations
func (m *MockMailingListClient) CreateComplaint(ctx context.Context, domain string, complaint *clients.ComplaintSpec) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) GetComplaint(ctx context.Context, domain, address string) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Unsubscribe suppression operations
func (m *MockMailingListClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe *clients.UnsubscribeSpec) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) GetUnsubscribe(ctx context.Context, domain, address string) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockMailingListClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func TestMailingListObserve(t *testing.T) {
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
		"MailingListExists": {
			reason: "Should return ResourceExists when mailing list exists",
			args: args{
				mg: &v1alpha1.MailingList{
					Spec: v1alpha1.MailingListSpec{
						ForProvider: v1alpha1.MailingListParameters{
							Address: "test@example.com",
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
		"MailingListNotFound": {
			reason: "Should return ResourceExists false when mailing list not found",
			args: args{
				mg: &v1alpha1.MailingList{
					Spec: v1alpha1.MailingListSpec{
						ForProvider: v1alpha1.MailingListParameters{
							Address: "notfound@example.com",
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
			// Setup mock client
			mockClient := &MockMailingListClient{}

			// Pre-populate with test mailing list for "exists" test
			if name == "MailingListExists" {
				mockClient.mailingLists = map[string]*clients.MailingList{
					"test@example.com": {
						Address:         "test@example.com",
						Name:            "Test List",
						Description:     "Test mailing list",
						AccessLevel:     "readonly",
						ReplyPreference: "list",
						CreatedAt:       "2025-01-01T00:00:00Z",
						MembersCount:    0,
					},
				}
			}

			e := &external{service: mockClient}
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

func TestMailingListCreate(t *testing.T) {
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
			reason: "Should successfully create mailing list",
			args: args{
				mg: &v1alpha1.MailingList{
					Spec: v1alpha1.MailingListSpec{
						ForProvider: v1alpha1.MailingListParameters{
							Address:     "new@example.com",
							Name:        stringPtr("New List"),
							AccessLevel: stringPtr("readonly"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalCreation{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockMailingListClient{}
			e := &external{service: mockClient}

			got, err := e.Create(context.Background(), tc.args.mg)

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

func TestMailingListUpdate(t *testing.T) {
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
			reason: "Should successfully update mailing list",
			args: args{
				mg: &v1alpha1.MailingList{
					Spec: v1alpha1.MailingListSpec{
						ForProvider: v1alpha1.MailingListParameters{
							Address:     "existing@example.com",
							Name:        stringPtr("Updated List"),
							AccessLevel: stringPtr("members"),
						},
					},
				},
			},
			want: want{
				o: managed.ExternalUpdate{
					ConnectionDetails: managed.ConnectionDetails{},
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockMailingListClient{
				mailingLists: map[string]*clients.MailingList{
					"existing@example.com": {
						Address:         "existing@example.com",
						Name:            "Original List",
						Description:     "Original description",
						AccessLevel:     "readonly",
						ReplyPreference: "list",
						CreatedAt:       "2025-01-01T00:00:00Z",
						MembersCount:    0,
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

func TestMailingListDelete(t *testing.T) {
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
			reason: "Should successfully delete mailing list",
			args: args{
				mg: &v1alpha1.MailingList{
					Spec: v1alpha1.MailingListSpec{
						ForProvider: v1alpha1.MailingListParameters{
							Address: "delete@example.com",
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
			mockClient := &MockMailingListClient{
				mailingLists: map[string]*clients.MailingList{
					"delete@example.com": {
						Address:         "delete@example.com",
						Name:            "Delete List",
						AccessLevel:     "readonly",
						ReplyPreference: "list",
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
				// Verify mailing list was deleted
				_, exists := mockClient.mailingLists["delete@example.com"]
				assert.False(t, exists)
			}
		})
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
