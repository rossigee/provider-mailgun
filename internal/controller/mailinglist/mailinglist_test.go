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
	"fmt"
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

// Additional comprehensive test coverage for MailingList controller

// Test error handling scenarios
func TestMailingListObserveErrors(t *testing.T) {
	cases := map[string]struct {
		reason     string
		mockErr    error
		setupMock  func(*MockMailingListClient)
		expectErr  bool
		expectExists bool
	}{
		"ClientError": {
			reason:    "Should handle client errors gracefully",
			mockErr:   errors.New("mailgun api error"),
			expectErr: true,
		},
		"MailingListNotFoundButUpToDate": {
			reason: "Should handle mailing list not found correctly",
			setupMock: func(m *MockMailingListClient) {
				// Mock will return "mailing list not found" error for nonexistent lists
			},
			expectErr:    false,
			expectExists: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockMailingListClient{err: tc.mockErr}
			if tc.setupMock != nil {
				tc.setupMock(mockClient)
			}

			e := &external{service: mockClient}
			mg := &v1alpha1.MailingList{
				Spec: v1alpha1.MailingListSpec{
					ForProvider: v1alpha1.MailingListParameters{
						Address: "test@example.com",
					},
				},
			}

			obs, err := e.Observe(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expectExists, obs.ResourceExists)
			}
		})
	}
}

func TestMailingListCreateErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"CreateError": {
			reason:    "Should handle create errors",
			mockErr:   errors.New("failed to create mailing list"),
			expectErr: true,
		},
		"DuplicateAddress": {
			reason:    "Should handle duplicate address errors",
			mockErr:   errors.New("address already exists"),
			expectErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockMailingListClient{err: tc.mockErr}
			e := &external{service: mockClient}

			mg := &v1alpha1.MailingList{
				Spec: v1alpha1.MailingListSpec{
					ForProvider: v1alpha1.MailingListParameters{
						Address:     "error@example.com",
						Name:        stringPtr("Error List"),
						Description: stringPtr("Error test list"),
						AccessLevel: stringPtr("readonly"),
					},
				},
			}

			_, err := e.Create(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMailingListUpdateErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"UpdateError": {
			reason:    "Should handle update errors",
			mockErr:   errors.New("failed to update mailing list"),
			expectErr: true,
		},
		"MailingListNotFound": {
			reason:    "Should handle mailing list not found during update",
			mockErr:   errors.New("mailing list not found (404)"),
			expectErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockMailingListClient{err: tc.mockErr}
			e := &external{service: mockClient}

			mg := &v1alpha1.MailingList{
				Spec: v1alpha1.MailingListSpec{
					ForProvider: v1alpha1.MailingListParameters{
						Address:     "error@example.com",
						Name:        stringPtr("Updated List"),
						Description: stringPtr("Updated description"),
					},
				},
			}

			_, err := e.Update(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMailingListDeleteErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"DeleteError": {
			reason:    "Should handle delete errors",
			mockErr:   errors.New("failed to delete mailing list"),
			expectErr: true,
		},
		"MailingListNotFound": {
			reason:    "Should handle mailing list not found during delete gracefully",
			mockErr:   errors.New("mailing list not found (404)"),
			expectErr: false, // Should handle 404 gracefully on delete
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockMailingListClient{err: tc.mockErr}
			e := &external{service: mockClient}

			mg := &v1alpha1.MailingList{
				Spec: v1alpha1.MailingListSpec{
					ForProvider: v1alpha1.MailingListParameters{
						Address: "error@example.com",
					},
				},
			}

			_, err := e.Delete(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test edge cases and boundary conditions
func TestMailingListEdgeCases(t *testing.T) {
	t.Run("MinimalMailingList", func(t *testing.T) {
		mockClient := &MockMailingListClient{}
		e := &external{service: mockClient}

		// Mailing list with only required fields
		mg := &v1alpha1.MailingList{
			Spec: v1alpha1.MailingListSpec{
				ForProvider: v1alpha1.MailingListParameters{
					Address: "minimal@example.com",
					// Name, Description, AccessLevel, ReplyPreference are nil/empty
				},
			},
		}

		_, err := e.Create(context.Background(), mg)
		require.NoError(t, err)

		// Verify mailing list was created with minimal fields
		list, exists := mockClient.mailingLists["minimal@example.com"]
		assert.True(t, exists)
		assert.Equal(t, "minimal@example.com", list.Address)
		assert.Equal(t, "Test List", list.Name) // Default from mock
	})

	t.Run("CompleteMailingListConfiguration", func(t *testing.T) {
		mockClient := &MockMailingListClient{}
		e := &external{service: mockClient}

		// Mailing list with all fields specified
		mg := &v1alpha1.MailingList{
			Spec: v1alpha1.MailingListSpec{
				ForProvider: v1alpha1.MailingListParameters{
					Address:         "complete@example.com",
					Name:            stringPtr("Complete Test List"),
					Description:     stringPtr("A comprehensive test mailing list with all configuration options"),
					AccessLevel:     stringPtr("members"),
					ReplyPreference: stringPtr("sender"),
				},
			},
		}

		_, err := e.Create(context.Background(), mg)
		require.NoError(t, err)

		// Verify mailing list was created with all specified fields
		list, exists := mockClient.mailingLists["complete@example.com"]
		assert.True(t, exists)
		assert.Equal(t, "complete@example.com", list.Address)
		assert.Equal(t, "Complete Test List", list.Name)
		assert.Equal(t, "A comprehensive test mailing list with all configuration options", list.Description)
		assert.Equal(t, "members", list.AccessLevel)
		assert.Equal(t, "sender", list.ReplyPreference)
	})

	t.Run("MailingListStatusUpdate", func(t *testing.T) {
		mockClient := &MockMailingListClient{
			mailingLists: map[string]*clients.MailingList{
				"status@example.com": {
					Address:         "status@example.com",
					Name:            "Status Test List",
					Description:     "List for status testing",
					AccessLevel:     "readonly",
					ReplyPreference: "list",
					CreatedAt:       "2025-01-01T00:00:00Z",
					MembersCount:    5,
				},
			},
		}
		e := &external{service: mockClient}

		mg := &v1alpha1.MailingList{
			Spec: v1alpha1.MailingListSpec{
				ForProvider: v1alpha1.MailingListParameters{
					Address: "status@example.com",
				},
			},
		}

		obs, err := e.Observe(context.Background(), mg)
		require.NoError(t, err)
		assert.True(t, obs.ResourceExists)
		assert.True(t, obs.ResourceUpToDate)

		// Verify the managed resource status is updated
		assert.Equal(t, "status@example.com", mg.Status.AtProvider.Address)
		assert.Equal(t, "Status Test List", mg.Status.AtProvider.Name)
		assert.Equal(t, "List for status testing", mg.Status.AtProvider.Description)
		assert.Equal(t, "readonly", mg.Status.AtProvider.AccessLevel)
		assert.Equal(t, "list", mg.Status.AtProvider.ReplyPreference)
		assert.Equal(t, "2025-01-01T00:00:00Z", mg.Status.AtProvider.CreatedAt)
		assert.Equal(t, 5, mg.Status.AtProvider.MembersCount)
	})

	t.Run("PartialFieldUpdate", func(t *testing.T) {
		mockClient := &MockMailingListClient{
			mailingLists: map[string]*clients.MailingList{
				"partial@example.com": {
					Address:         "partial@example.com",
					Name:            "Original Name",
					Description:     "Original Description",
					AccessLevel:     "readonly",
					ReplyPreference: "list",
					CreatedAt:       "2025-01-01T00:00:00Z",
					MembersCount:    0,
				},
			},
		}
		e := &external{service: mockClient}

		// Update only the name field
		mg := &v1alpha1.MailingList{
			Spec: v1alpha1.MailingListSpec{
				ForProvider: v1alpha1.MailingListParameters{
					Address: "partial@example.com",
					Name:    stringPtr("Updated Name Only"),
					// Description, AccessLevel, ReplyPreference not specified
				},
			},
		}

		_, err := e.Update(context.Background(), mg)
		require.NoError(t, err)

		// Verify only the name was updated, other fields remain unchanged
		list := mockClient.mailingLists["partial@example.com"]
		assert.Equal(t, "Updated Name Only", list.Name)
		assert.Equal(t, "Original Description", list.Description) // Should remain unchanged
		assert.Equal(t, "readonly", list.AccessLevel)             // Should remain unchanged
		assert.Equal(t, "list", list.ReplyPreference)             // Should remain unchanged
	})
}

// Test different access levels and reply preferences
func TestMailingListAccessLevelsAndReplyPreferences(t *testing.T) {
	accessLevels := []string{"readonly", "members", "everyone"}
	replyPreferences := []string{"list", "sender"}

	for _, accessLevel := range accessLevels {
		for _, replyPref := range replyPreferences {
			t.Run(fmt.Sprintf("AccessLevel_%s_ReplyPref_%s", accessLevel, replyPref), func(t *testing.T) {
				mockClient := &MockMailingListClient{}
				e := &external{service: mockClient}

				address := fmt.Sprintf("%s-%s@example.com", accessLevel, replyPref)
				mg := &v1alpha1.MailingList{
					Spec: v1alpha1.MailingListSpec{
						ForProvider: v1alpha1.MailingListParameters{
							Address:         address,
							Name:            stringPtr(fmt.Sprintf("Test %s %s", accessLevel, replyPref)),
							AccessLevel:     &accessLevel,
							ReplyPreference: &replyPref,
						},
					},
				}

				_, err := e.Create(context.Background(), mg)
				require.NoError(t, err)

				// Verify the configuration was set correctly
				list, exists := mockClient.mailingLists[address]
				assert.True(t, exists)
				assert.Equal(t, accessLevel, list.AccessLevel)
				assert.Equal(t, replyPref, list.ReplyPreference)
			})
		}
	}
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
