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

// CreateTemplate creates a new email template for a domain
func (c *mailgunClient) CreateTemplate(ctx context.Context, domain string, template *TemplateSpec) (*Template, error) {
	path := fmt.Sprintf("/domains/%s/templates", domain)

	params := map[string]interface{}{
		"name": template.Name,
	}

	if template.Description != nil {
		params["description"] = *template.Description
	}
	if template.Template != nil {
		params["template"] = *template.Template
	}
	if template.Engine != nil {
		params["engine"] = *template.Engine
	}
	if template.Comment != nil {
		params["comment"] = *template.Comment
	}
	if template.Tag != nil {
		params["tag"] = *template.Tag
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	var result struct {
		Template *Template `json:"template"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return result.Template, nil
}

// GetTemplate retrieves a template by name
func (c *mailgunClient) GetTemplate(ctx context.Context, domain, name string) (*Template, error) {
	path := fmt.Sprintf("/domains/%s/templates/%s", domain, name)

	// Request with active flag to get the active version content
	resp, err := c.makeRequest(ctx, "GET", path+"?active=yes", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	var result struct {
		Template *Template `json:"template"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	return result.Template, nil
}

// UpdateTemplate updates a template's description
func (c *mailgunClient) UpdateTemplate(ctx context.Context, domain, name string, template *TemplateSpec) (*Template, error) {
	path := fmt.Sprintf("/domains/%s/templates/%s", domain, name)

	params := map[string]interface{}{}

	// Only description can be updated via PUT
	if template.Description != nil {
		params["description"] = *template.Description
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "PUT", path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	var result struct {
		Template *Template `json:"template"`
		Message  string    `json:"message"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	// If result.Template is nil, fetch the updated template
	if result.Template == nil {
		return c.GetTemplate(ctx, domain, name)
	}

	return result.Template, nil
}

// DeleteTemplate deletes a template and all its versions
func (c *mailgunClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	path := fmt.Sprintf("/domains/%s/templates/%s", domain, name)

	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}

	var result interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		// Many DELETE endpoints return empty responses, so we ignore parse errors
		return nil
	}

	return nil
}
