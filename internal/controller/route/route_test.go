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

package route

import (
	"context"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"

	"github.com/rossigee/provider-mailgun/apis/route/v1alpha1"
	"github.com/rossigee/provider-mailgun/internal/clients"
)

// MockRouteClient for testing
type MockRouteClient struct {
	routes map[string]*clients.Route
	err    error
}

func (m *MockRouteClient) CreateRoute(ctx context.Context, route *clients.RouteSpec) (*clients.Route, error) {
	if m.err != nil {
		return nil, m.err
	}

	result := &clients.Route{
		ID:          "route_123",
		Priority:    0,
		Description: "Test route",
		Expression:  route.Expression,
		CreatedAt:   "2025-01-01T00:00:00Z",
	}

	if route.Priority != nil {
		result.Priority = *route.Priority
	}
	if route.Description != nil {
		result.Description = *route.Description
	}
	if len(route.Actions) > 0 {
		result.Actions = make([]clients.RouteAction, len(route.Actions))
		for i, action := range route.Actions {
			result.Actions[i] = clients.RouteAction{
				Type:        action.Type,
				Destination: action.Destination,
			}
		}
	}

	if m.routes == nil {
		m.routes = make(map[string]*clients.Route)
	}
	m.routes[result.ID] = result

	return result, nil
}

func (m *MockRouteClient) GetRoute(ctx context.Context, id string) (*clients.Route, error) {
	if m.err != nil {
		return nil, m.err
	}

	if route, exists := m.routes[id]; exists {
		return route, nil
	}

	return nil, errors.New("route not found (404)")
}

func (m *MockRouteClient) UpdateRoute(ctx context.Context, id string, route *clients.RouteSpec) (*clients.Route, error) {
	if m.err != nil {
		return nil, m.err
	}

	if existing, exists := m.routes[id]; exists {
		// Update modifiable fields
		if route.Priority != nil {
			existing.Priority = *route.Priority
		}
		if route.Description != nil {
			existing.Description = *route.Description
		}
		existing.Expression = route.Expression
		if len(route.Actions) > 0 {
			existing.Actions = make([]clients.RouteAction, len(route.Actions))
			for i, action := range route.Actions {
				existing.Actions[i] = clients.RouteAction{
					Type:        action.Type,
					Destination: action.Destination,
				}
			}
		}
		return existing, nil
	}

	return nil, errors.New("route not found (404)")
}

func (m *MockRouteClient) DeleteRoute(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}

	delete(m.routes, id)
	return nil
}

// Implement other required client methods as no-ops
func (m *MockRouteClient) CreateDomain(ctx context.Context, domain *clients.DomainSpec) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) GetDomain(ctx context.Context, name string) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) UpdateDomain(ctx context.Context, name string, domain *clients.DomainSpec) (*clients.Domain, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) DeleteDomain(ctx context.Context, name string) error {
	return errors.New("not implemented")
}

func (m *MockRouteClient) CreateMailingList(ctx context.Context, list *clients.MailingListSpec) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) GetMailingList(ctx context.Context, address string) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) UpdateMailingList(ctx context.Context, address string, list *clients.MailingListSpec) (*clients.MailingList, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) DeleteMailingList(ctx context.Context, address string) error {
	return errors.New("not implemented")
}

func (m *MockRouteClient) CreateWebhook(ctx context.Context, domain string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) GetWebhook(ctx context.Context, domain, eventType string) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *clients.WebhookSpec) (*clients.Webhook, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return errors.New("not implemented")
}

func (m *MockRouteClient) CreateSMTPCredential(ctx context.Context, domain string, credential *clients.SMTPCredentialSpec) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) GetSMTPCredential(ctx context.Context, domain, login string) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*clients.SMTPCredential, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return errors.New("not implemented")
}

func (m *MockRouteClient) CreateTemplate(ctx context.Context, domain string, template *clients.TemplateSpec) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) GetTemplate(ctx context.Context, domain, name string) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) UpdateTemplate(ctx context.Context, domain, name string, template *clients.TemplateSpec) (*clients.Template, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return errors.New("not implemented")
}

// Bounce suppression operations
func (m *MockRouteClient) CreateBounce(ctx context.Context, domain string, bounce *clients.BounceSpec) (*clients.Bounce, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) GetBounce(ctx context.Context, domain, address string) (*clients.Bounce, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Complaint suppression operations
func (m *MockRouteClient) CreateComplaint(ctx context.Context, domain string, complaint *clients.ComplaintSpec) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) GetComplaint(ctx context.Context, domain, address string) (*clients.Complaint, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

// Unsubscribe suppression operations
func (m *MockRouteClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe *clients.UnsubscribeSpec) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) GetUnsubscribe(ctx context.Context, domain, address string) (*clients.Unsubscribe, error) {
	return nil, errors.New("not implemented")
}

func (m *MockRouteClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return errors.New("not implemented")
}

func TestRouteObserve(t *testing.T) {
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
		"RouteExists": {
			reason: "Should return ResourceExists when route exists",
			args: args{
				mg: func() *v1alpha1.Route {
					r := &v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							ForProvider: v1alpha1.RouteParameters{
								Expression: "match_recipient(\".*@example.com\")",
								Actions: []v1alpha1.RouteAction{
									{
										Type:        "forward",
										Destination: stringPtr("user@destination.com"),
									},
								},
							},
						},
					}
					r.SetName("test-route")
					r.SetAnnotations(map[string]string{
						"crossplane.io/external-name": "route_123",
					})
					return r
				}(),
			},
			want: want{
				o: managed.ExternalObservation{
					ResourceExists:   true,
					ResourceUpToDate: true,
				},
			},
		},
		"RouteNotFound": {
			reason: "Should return ResourceExists false when route not found",
			args: args{
				mg: func() *v1alpha1.Route {
					r := &v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							ForProvider: v1alpha1.RouteParameters{
								Expression: "match_recipient(\".*@notfound.com\")",
								Actions: []v1alpha1.RouteAction{
									{Type: "stop"},
								},
							},
						},
					}
					r.SetName("notfound-route")
					r.SetAnnotations(map[string]string{
						"crossplane.io/external-name": "route_notfound",
					})
					return r
				}(),
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
			mockClient := &MockRouteClient{}

			// Pre-populate with test route for "exists" test
			if name == "RouteExists" {
				mockClient.routes = map[string]*clients.Route{
					"route_123": {
						ID:          "route_123",
						Priority:    0,
						Description: "Test route",
						Expression:  "match_recipient(\".*@example.com\")",
						Actions: []clients.RouteAction{
							{
								Type:        "forward",
								Destination: stringPtr("user@destination.com"),
							},
						},
						CreatedAt: "2025-01-01T00:00:00Z",
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

func TestRouteCreate(t *testing.T) {
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
			reason: "Should successfully create route",
			args: args{
				mg: &v1alpha1.Route{
					Spec: v1alpha1.RouteSpec{
						ForProvider: v1alpha1.RouteParameters{
							Priority:    intPtr(10),
							Description: stringPtr("New route"),
							Expression:  "match_recipient(\".*@new.com\")",
							Actions: []v1alpha1.RouteAction{
								{
									Type:        "forward",
									Destination: stringPtr("admin@new.com"),
								},
							},
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
			mockClient := &MockRouteClient{}
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

func TestRouteDelete(t *testing.T) {
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
			reason: "Should successfully delete route",
			args: args{
				mg: func() *v1alpha1.Route {
					r := &v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							ForProvider: v1alpha1.RouteParameters{
								Expression: "match_recipient(\".*@delete.com\")",
								Actions: []v1alpha1.RouteAction{
									{Type: "stop"},
								},
							},
						},
					}
					r.SetAnnotations(map[string]string{
						"crossplane.io/external-name": "route_delete",
					})
					return r
				}(),
			},
			want: want{
				err: nil,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockRouteClient{
				routes: map[string]*clients.Route{
					"route_delete": {
						ID:         "route_delete",
						Priority:   0,
						Expression: "match_recipient(\".*@delete.com\")",
						Actions: []clients.RouteAction{
							{Type: "stop"},
						},
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
				// Verify route was deleted
				_, exists := mockClient.routes["route_delete"]
				assert.False(t, exists)
			}
		})
	}
}

func TestRouteUpdate(t *testing.T) {
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
			reason: "Should successfully update route",
			args: args{
				mg: func() *v1alpha1.Route {
					r := &v1alpha1.Route{
						Spec: v1alpha1.RouteSpec{
							ForProvider: v1alpha1.RouteParameters{
								Priority:    intPtr(20),
								Description: stringPtr("Updated route"),
								Expression:  "match_recipient(\".*@updated.com\")",
								Actions: []v1alpha1.RouteAction{
									{
										Type:        "forward",
										Destination: stringPtr("updated@destination.com"),
									},
								},
							},
						},
					}
					r.SetAnnotations(map[string]string{
						"crossplane.io/external-name": "route_update",
					})
					return r
				}(),
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
			mockClient := &MockRouteClient{
				routes: map[string]*clients.Route{
					"route_update": {
						ID:          "route_update",
						Priority:    10,
						Description: "Original route",
						Expression:  "match_recipient(\".*@original.com\")",
						Actions: []clients.RouteAction{
							{
								Type:        "forward",
								Destination: stringPtr("original@destination.com"),
							},
						},
						CreatedAt: "2025-01-01T00:00:00Z",
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

				// Verify the route was updated
				updatedRoute := mockClient.routes["route_update"]
				assert.Equal(t, 20, updatedRoute.Priority)
				assert.Equal(t, "Updated route", updatedRoute.Description)
				assert.Equal(t, "match_recipient(\".*@updated.com\")", updatedRoute.Expression)
			}
		})
	}
}

// Additional comprehensive test coverage for Route controller

// Test error handling scenarios
func TestRouteObserveErrors(t *testing.T) {
	cases := map[string]struct {
		reason     string
		mockErr    error
		setupMock  func(*MockRouteClient)
		expectErr  bool
		expectExists bool
	}{
		"ClientError": {
			reason:    "Should handle client errors gracefully",
			mockErr:   errors.New("mailgun api error"),
			expectErr: true,
		},
		"RouteNotFoundButUpToDate": {
			reason: "Should handle route not found correctly",
			setupMock: func(m *MockRouteClient) {
				// Mock will return "route not found" error for nonexistent routes
			},
			expectErr:    false,
			expectExists: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockRouteClient{err: tc.mockErr}
			if tc.setupMock != nil {
				tc.setupMock(mockClient)
			}

			e := &external{service: mockClient}
			mg := func() *v1alpha1.Route {
				r := &v1alpha1.Route{
					Spec: v1alpha1.RouteSpec{
						ForProvider: v1alpha1.RouteParameters{
							Expression: "match_recipient(\".*@test.com\")",
							Actions:    []v1alpha1.RouteAction{{Type: "stop"}},
						},
					},
				}
				r.SetAnnotations(map[string]string{
					"crossplane.io/external-name": "route_test",
				})
				return r
			}()

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

func TestRouteCreateErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"CreateError": {
			reason:    "Should handle create errors",
			mockErr:   errors.New("failed to create route"),
			expectErr: true,
		},
		"InvalidExpression": {
			reason:    "Should handle invalid expression errors",
			mockErr:   errors.New("invalid expression syntax"),
			expectErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockRouteClient{err: tc.mockErr}
			e := &external{service: mockClient}

			mg := &v1alpha1.Route{
				Spec: v1alpha1.RouteSpec{
					ForProvider: v1alpha1.RouteParameters{
						Priority:    intPtr(10),
						Description: stringPtr("Error test route"),
						Expression:  "invalid_expression",
						Actions: []v1alpha1.RouteAction{
							{Type: "stop"},
						},
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

func TestRouteUpdateErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"UpdateError": {
			reason:    "Should handle update errors",
			mockErr:   errors.New("failed to update route"),
			expectErr: true,
		},
		"RouteNotFound": {
			reason:    "Should handle route not found during update",
			mockErr:   errors.New("route not found (404)"),
			expectErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockRouteClient{err: tc.mockErr}
			e := &external{service: mockClient}

			mg := func() *v1alpha1.Route {
				r := &v1alpha1.Route{
					Spec: v1alpha1.RouteSpec{
						ForProvider: v1alpha1.RouteParameters{
							Expression: "match_recipient(\".*@error.com\")",
							Actions:    []v1alpha1.RouteAction{{Type: "stop"}},
						},
					},
				}
				r.SetAnnotations(map[string]string{
					"crossplane.io/external-name": "route_error",
				})
				return r
			}()

			_, err := e.Update(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRouteDeleteErrors(t *testing.T) {
	cases := map[string]struct {
		reason    string
		mockErr   error
		expectErr bool
	}{
		"DeleteError": {
			reason:    "Should handle delete errors",
			mockErr:   errors.New("failed to delete route"),
			expectErr: true,
		},
		"RouteNotFound": {
			reason:    "Should handle route not found during delete gracefully",
			mockErr:   errors.New("route not found (404)"),
			expectErr: false, // Should handle 404 gracefully on delete
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			mockClient := &MockRouteClient{err: tc.mockErr}
			e := &external{service: mockClient}

			mg := func() *v1alpha1.Route {
				r := &v1alpha1.Route{
					Spec: v1alpha1.RouteSpec{
						ForProvider: v1alpha1.RouteParameters{
							Expression: "match_recipient(\".*@error.com\")",
							Actions:    []v1alpha1.RouteAction{{Type: "stop"}},
						},
					},
				}
				r.SetAnnotations(map[string]string{
					"crossplane.io/external-name": "route_error",
				})
				return r
			}()

			_, err := e.Delete(context.Background(), mg)

			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// Test edge cases and complex routing scenarios
func TestRouteEdgeCases(t *testing.T) {
	t.Run("MinimalRoute", func(t *testing.T) {
		mockClient := &MockRouteClient{}
		e := &external{service: mockClient}

		// Route with only required fields
		mg := &v1alpha1.Route{
			Spec: v1alpha1.RouteSpec{
				ForProvider: v1alpha1.RouteParameters{
					Expression: "match_recipient(\".*@minimal.com\")",
					Actions: []v1alpha1.RouteAction{
						{Type: "stop"},
					},
					// Priority and Description are nil/empty
				},
			},
		}

		_, err := e.Create(context.Background(), mg)
		require.NoError(t, err)

		// Verify route was created with minimal fields
		route := mockClient.routes["route_123"] // Default ID from mock
		assert.Equal(t, "match_recipient(\".*@minimal.com\")", route.Expression)
		assert.Equal(t, 0, route.Priority) // Default from mock
		assert.Len(t, route.Actions, 1)
		assert.Equal(t, "stop", route.Actions[0].Type)
	})

	t.Run("ComplexRouteWithMultipleActions", func(t *testing.T) {
		mockClient := &MockRouteClient{}
		e := &external{service: mockClient}

		// Route with multiple actions and complex expression
		mg := &v1alpha1.Route{
			Spec: v1alpha1.RouteSpec{
				ForProvider: v1alpha1.RouteParameters{
					Priority:    intPtr(50),
					Description: stringPtr("Complex multi-action route for sales team"),
					Expression:  "match_recipient(\"sales-.*@company.com\") AND match_header(\"X-Priority\", \"high\")",
					Actions: []v1alpha1.RouteAction{
						{
							Type:        "forward",
							Destination: stringPtr("sales-manager@company.com"),
						},
						{
							Type:        "store",
							Destination: stringPtr("high-priority-sales-archive"),
						},
					},
				},
			},
		}

		_, err := e.Create(context.Background(), mg)
		require.NoError(t, err)

		// Verify route was created with all specified fields
		route := mockClient.routes["route_123"]
		assert.Equal(t, 50, route.Priority)
		assert.Equal(t, "Complex multi-action route for sales team", route.Description)
		assert.Equal(t, "match_recipient(\"sales-.*@company.com\") AND match_header(\"X-Priority\", \"high\")", route.Expression)
		assert.Len(t, route.Actions, 2)
		assert.Equal(t, "forward", route.Actions[0].Type)
		assert.Equal(t, "sales-manager@company.com", *route.Actions[0].Destination)
		assert.Equal(t, "store", route.Actions[1].Type)
		assert.Equal(t, "high-priority-sales-archive", *route.Actions[1].Destination)
	})

	t.Run("RouteStatusUpdate", func(t *testing.T) {
		mockClient := &MockRouteClient{
			routes: map[string]*clients.Route{
				"route_status": {
					ID:          "route_status",
					Priority:    25,
					Description: "Route for status testing",
					Expression:  "match_recipient(\".*@status.com\")",
					Actions: []clients.RouteAction{
						{
							Type:        "forward",
							Destination: stringPtr("admin@status.com"),
						},
					},
					CreatedAt: "2025-01-01T00:00:00Z",
				},
			},
		}
		e := &external{service: mockClient}

		mg := func() *v1alpha1.Route {
			r := &v1alpha1.Route{
				Spec: v1alpha1.RouteSpec{
					ForProvider: v1alpha1.RouteParameters{
						Expression: "match_recipient(\".*@status.com\")",
						Actions:    []v1alpha1.RouteAction{{Type: "forward"}},
					},
				},
			}
			r.SetAnnotations(map[string]string{
				"crossplane.io/external-name": "route_status",
			})
			return r
		}()

		obs, err := e.Observe(context.Background(), mg)
		require.NoError(t, err)
		assert.True(t, obs.ResourceExists)
		// Note: ResourceUpToDate depends on comparison logic in the actual controller

		// Verify the managed resource status is updated
		assert.Equal(t, "route_status", mg.Status.AtProvider.ID)
		assert.Equal(t, 25, mg.Status.AtProvider.Priority)
		assert.Equal(t, "Route for status testing", mg.Status.AtProvider.Description)
		assert.Equal(t, "match_recipient(\".*@status.com\")", mg.Status.AtProvider.Expression)
		assert.Equal(t, "2025-01-01T00:00:00Z", mg.Status.AtProvider.CreatedAt)
		assert.Len(t, mg.Status.AtProvider.Actions, 1)
	})
}

// Test different route action types and patterns
func TestRouteActionTypes(t *testing.T) {
	actionTypes := []struct {
		type_ string
		dest  *string
	}{
		{"forward", stringPtr("user@example.com")},
		{"store", stringPtr("mailing-list-archive")},
		{"stop", nil},
	}

	for _, action := range actionTypes {
		t.Run(fmt.Sprintf("ActionType_%s", action.type_), func(t *testing.T) {
			mockClient := &MockRouteClient{}
			e := &external{service: mockClient}

			mg := &v1alpha1.Route{
				Spec: v1alpha1.RouteSpec{
					ForProvider: v1alpha1.RouteParameters{
						Expression: fmt.Sprintf("match_recipient(\".*@%s.com\")", action.type_),
						Actions: []v1alpha1.RouteAction{
							{
								Type:        action.type_,
								Destination: action.dest,
							},
						},
					},
				},
			}

			_, err := e.Create(context.Background(), mg)
			require.NoError(t, err)

			// Verify the action was configured correctly
			route := mockClient.routes["route_123"]
			assert.Len(t, route.Actions, 1)
			assert.Equal(t, action.type_, route.Actions[0].Type)
			if action.dest != nil {
				assert.Equal(t, action.dest, route.Actions[0].Destination)
			} else {
				assert.Nil(t, route.Actions[0].Destination)
			}
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}
