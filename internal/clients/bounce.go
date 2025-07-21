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

// CreateBounce creates a new bounce suppression entry for a domain
func (c *mailgunClient) CreateBounce(ctx context.Context, domain string, bounce *BounceSpec) (*Bounce, error) {
	path := fmt.Sprintf("/domains/%s/bounces", domain)

	params := map[string]interface{}{
		"address": bounce.Address,
	}
	if bounce.Code != nil {
		params["code"] = *bounce.Code
	}
	if bounce.Error != nil {
		params["error"] = *bounce.Error
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create bounce: %w", err)
	}

	var result Bounce
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return &result, nil
}

// GetBounce retrieves a bounce suppression entry
func (c *mailgunClient) GetBounce(ctx context.Context, domain, address string) (*Bounce, error) {
	path := fmt.Sprintf("/domains/%s/bounces/%s", domain, address)

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get bounce: %w", err)
	}

	var result Bounce
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return &result, nil
}

// DeleteBounce deletes a bounce suppression entry
func (c *mailgunClient) DeleteBounce(ctx context.Context, domain, address string) error {
	path := fmt.Sprintf("/domains/%s/bounces/%s", domain, address)

	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete bounce: %w", err)
	}

	// For DELETE operations, we usually don't need to parse the response
	var result interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		// Many DELETE endpoints return empty responses, so we ignore parse errors
		return nil
	}

	return nil
}
