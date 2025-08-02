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

package clients

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMailingList(t *testing.T) {
	tests := []struct {
		name         string
		listSpec     *MailingListSpec
		expectedList *MailingList
		expectedError bool
	}{
		{
			name: "successful creation with minimal params",
			listSpec: &MailingListSpec{
				Address: "test@example.com",
			},
			expectedList: &MailingList{
				Address:         "test@example.com",
				Name:            "Test List",
				AccessLevel:     "readonly",
				ReplyPreference: "list",
				CreatedAt:       "2025-01-01T00:00:00Z",
				MembersCount:    0,
			},
			expectedError: false,
		},
		{
			name: "successful creation with all params",
			listSpec: &MailingListSpec{
				Address:         "full@example.com",
				Name:            stringPtr("Full List"),
				Description:     stringPtr("Complete test list"),
				AccessLevel:     stringPtr("members"),
				ReplyPreference: stringPtr("sender"),
			},
			expectedList: &MailingList{
				Address:         "full@example.com",
				Name:            "Full List",
				Description:     "Complete test list",
				AccessLevel:     "members",
				ReplyPreference: "sender",
				CreatedAt:       "2025-01-01T00:00:00Z",
				MembersCount:    0,
			},
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				assert.Equal(t, "/v3/lists", r.URL.Path)

				// Verify request body
				_ = r.ParseForm()
				assert.Equal(t, tt.listSpec.Address, r.FormValue("address"))
				if tt.listSpec.Name != nil {
					assert.Equal(t, *tt.listSpec.Name, r.FormValue("name"))
				}
				if tt.listSpec.AccessLevel != nil {
					assert.Equal(t, *tt.listSpec.AccessLevel, r.FormValue("access_level"))
				}

				w.WriteHeader(http.StatusOK)
				response := map[string]interface{}{
					"list": tt.expectedList,
				}
				_ = json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL + "/v3",
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			result, err := client.CreateMailingList(context.Background(), tt.listSpec)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedList, result)
			}
		})
	}
}

func TestGetMailingList(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		expectedList  *MailingList
		expectedError bool
		statusCode    int
	}{
		{
			name:    "successful get",
			address: "get@example.com",
			expectedList: &MailingList{
				Address:         "get@example.com",
				Name:            "Get List",
				Description:     "List for testing get",
				AccessLevel:     "readonly",
				ReplyPreference: "list",
				CreatedAt:       "2025-01-01T00:00:00Z",
				MembersCount:    5,
			},
			expectedError: false,
			statusCode:    200,
		},
		{
			name:          "list not found",
			address:       "notfound@example.com",
			expectedList:  nil,
			expectedError: true,
			statusCode:    404,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, "/v3/lists/"+tt.address, r.URL.Path)

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					response := map[string]interface{}{
						"list": tt.expectedList,
					}
					_ = json.NewEncoder(w).Encode(response)
				} else {
					_, _ = w.Write([]byte("List not found"))
				}
			}))
			defer server.Close()

			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL + "/v3",
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			result, err := client.GetMailingList(context.Background(), tt.address)

			if tt.expectedError {
				require.Error(t, err)
				assert.Nil(t, result)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedList, result)
			}
		})
	}
}

func TestUpdateMailingList(t *testing.T) {
	listSpec := &MailingListSpec{
		Name:            stringPtr("Updated List"),
		Description:     stringPtr("Updated description"),
		AccessLevel:     stringPtr("members"),
		ReplyPreference: stringPtr("sender"),
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Equal(t, "/v3/lists/update@example.com", r.URL.Path)

		// Verify request body
		_ = r.ParseForm()
		assert.Equal(t, "Updated List", r.FormValue("name"))
		assert.Equal(t, "Updated description", r.FormValue("description"))
		assert.Equal(t, "members", r.FormValue("access_level"))
		assert.Equal(t, "sender", r.FormValue("reply_preference"))

		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"list": map[string]interface{}{
				"address":          "update@example.com",
				"name":             "Updated List",
				"description":      "Updated description",
				"access_level":     "members",
				"reply_preference": "sender",
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	config := &Config{
		APIKey:     "test-key",
		BaseURL:    server.URL + "/v3",
		HTTPClient: &http.Client{},
	}
	client := NewClient(config)

	result, err := client.UpdateMailingList(context.Background(), "update@example.com", listSpec)

	require.NoError(t, err)
	assert.Equal(t, "update@example.com", result.Address)
	assert.Equal(t, "Updated List", result.Name)
	assert.Equal(t, "members", result.AccessLevel)
}

func TestDeleteMailingList(t *testing.T) {
	tests := []struct {
		name          string
		address       string
		statusCode    int
		expectedError bool
	}{
		{
			name:          "successful delete",
			address:       "delete@example.com",
			statusCode:    200,
			expectedError: false,
		},
		{
			name:          "list not found",
			address:       "notfound@example.com",
			statusCode:    404,
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "DELETE", r.Method)
				assert.Equal(t, "/v3/lists/"+tt.address, r.URL.Path)

				w.WriteHeader(tt.statusCode)
				if tt.statusCode == 200 {
					response := map[string]interface{}{
						"message": "List has been deleted",
					}
					_ = json.NewEncoder(w).Encode(response)
				} else {
					_, _ = w.Write([]byte("List not found"))
				}
			}))
			defer server.Close()

			config := &Config{
				APIKey:     "test-key",
				BaseURL:    server.URL + "/v3",
				HTTPClient: &http.Client{},
			}
			client := NewClient(config)

			err := client.DeleteMailingList(context.Background(), tt.address)

			if tt.expectedError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
