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

	"github.com/pkg/errors"
)

// CreateDomain creates a new domain in Mailgun
func (c *mailgunClient) CreateDomain(ctx context.Context, domain *DomainSpec) (*Domain, error) {
	params := map[string]interface{}{
		"name": domain.Name,
	}

	if domain.Type != nil {
		params["type"] = *domain.Type
	}
	if domain.ForceDKIMAuthority != nil {
		params["force_dkim_authority"] = *domain.ForceDKIMAuthority
	}
	if domain.DKIMKeySize != nil {
		params["dkim_key_size"] = *domain.DKIMKeySize
	}
	if domain.SMTPPassword != nil {
		params["smtp_password"] = *domain.SMTPPassword
	}
	if domain.SpamAction != nil {
		params["spam_action"] = *domain.SpamAction
	}
	if domain.WebScheme != nil {
		params["web_scheme"] = *domain.WebScheme
	}
	if domain.Wildcard != nil {
		params["wildcard"] = *domain.Wildcard
	}
	if len(domain.IPs) > 0 {
		params["ips"] = strings.Join(domain.IPs, ",")
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "POST", "/domains", body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create domain")
	}

	var result struct {
		Domain *Domain `json:"domain"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.Domain, nil
}

// GetDomain retrieves a domain from Mailgun
func (c *mailgunClient) GetDomain(ctx context.Context, name string) (*Domain, error) {
	path := fmt.Sprintf("/domains/%s", name)
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get domain")
	}

	var result struct {
		Domain *Domain `json:"domain"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.Domain, nil
}

// UpdateDomain updates an existing domain in Mailgun
func (c *mailgunClient) UpdateDomain(ctx context.Context, name string, domain *DomainSpec) (*Domain, error) {
	params := map[string]interface{}{}

	if domain.SpamAction != nil {
		params["spam_action"] = *domain.SpamAction
	}
	if domain.WebScheme != nil {
		params["web_scheme"] = *domain.WebScheme
	}
	if domain.Wildcard != nil {
		params["wildcard"] = *domain.Wildcard
	}

	body := strings.NewReader(createFormData(params))
	path := fmt.Sprintf("/domains/%s", name)
	resp, err := c.makeRequest(ctx, "PUT", path, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update domain")
	}

	var result struct {
		Domain *Domain `json:"domain"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.Domain, nil
}

// DeleteDomain deletes a domain from Mailgun
func (c *mailgunClient) DeleteDomain(ctx context.Context, name string) error {
	path := fmt.Sprintf("/domains/%s", name)
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return errors.Wrap(err, "failed to delete domain")
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return errors.Wrap(err, "failed to handle response")
	}

	return nil
}
