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
	"fmt"
	"strings"
)

// CreateUnsubscribe creates a new unsubscribe suppression entry for a domain
func (c *mailgunClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe interface{}) (interface{}, error) {
	// Type assert to UnsubscribeSpec
	unsubscribeSpec, ok := unsubscribe.(*UnsubscribeSpec)
	if !ok {
		return nil, fmt.Errorf("invalid unsubscribe parameter type")
	}
	path := fmt.Sprintf("/domains/%s/unsubscribes", domain)

	params := map[string]interface{}{
		"address": unsubscribeSpec.Address,
	}
	if unsubscribeSpec.Tags != nil {
		params["tags"] = *unsubscribeSpec.Tags
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create unsubscribe: %w", err)
	}

	var result Unsubscribe
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return interface{}(&result), nil
}

// GetUnsubscribe retrieves an unsubscribe suppression entry
func (c *mailgunClient) GetUnsubscribe(ctx context.Context, domain, address string) (interface{}, error) {
	path := fmt.Sprintf("/domains/%s/unsubscribes/%s", domain, address)

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get unsubscribe: %w", err)
	}

	var result Unsubscribe
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return interface{}(&result), nil
}

// DeleteUnsubscribe deletes an unsubscribe suppression entry
func (c *mailgunClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	path := fmt.Sprintf("/domains/%s/unsubscribes/%s", domain, address)

	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete unsubscribe: %w", err)
	}

	// For DELETE operations, we usually don't need to parse the response
	var result interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		// Many DELETE endpoints return empty responses, so we ignore parse errors
		return nil
	}

	return nil
}
