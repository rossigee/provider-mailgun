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

// CreateComplaint creates a new complaint suppression entry for a domain
func (c *mailgunClient) CreateComplaint(ctx context.Context, domain string, complaint interface{}) (interface{}, error) {
	// Type assert to ComplaintSpec
	complaintSpec, ok := complaint.(*ComplaintSpec)
	if !ok {
		return nil, fmt.Errorf("invalid complaint parameter type")
	}
	path := fmt.Sprintf("/domains/%s/complaints", domain)

	params := map[string]interface{}{
		"address": complaintSpec.Address,
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create complaint: %w", err)
	}

	var result Complaint
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return interface{}(&result), nil
}

// GetComplaint retrieves a complaint suppression entry
func (c *mailgunClient) GetComplaint(ctx context.Context, domain, address string) (interface{}, error) {
	path := fmt.Sprintf("/domains/%s/complaints/%s", domain, address)

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get complaint: %w", err)
	}

	var result Complaint
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return interface{}(&result), nil
}

// DeleteComplaint deletes a complaint suppression entry
func (c *mailgunClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	path := fmt.Sprintf("/domains/%s/complaints/%s", domain, address)

	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete complaint: %w", err)
	}

	// For DELETE operations, we usually don't need to parse the response
	var result interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		// Many DELETE endpoints return empty responses, so we ignore parse errors
		return nil
	}

	return nil
}
