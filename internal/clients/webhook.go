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

// CreateWebhook creates a new webhook in Mailgun
func (c *mailgunClient) CreateWebhook(ctx context.Context, domain string, webhook *WebhookSpec) (*Webhook, error) {
	params := map[string]interface{}{
		"url": webhook.URL,
	}

	if webhook.Username != nil {
		params["username"] = *webhook.Username
	}
	if webhook.Password != nil {
		params["password"] = *webhook.Password
	}

	body := strings.NewReader(createFormData(params))
	path := fmt.Sprintf("/domains/%s/webhooks/%s", domain, webhook.EventType)
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create webhook")
	}

	var result struct {
		Webhook *Webhook `json:"webhook"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	// Set the domain and event type
	result.Webhook.Domain = domain
	result.Webhook.EventType = webhook.EventType

	return result.Webhook, nil
}

// GetWebhook retrieves a webhook from Mailgun
func (c *mailgunClient) GetWebhook(ctx context.Context, domain, eventType string) (*Webhook, error) {
	path := fmt.Sprintf("/domains/%s/webhooks/%s", domain, eventType)
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get webhook")
	}

	var result struct {
		Webhook *Webhook `json:"webhook"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	// Set the domain and event type
	result.Webhook.Domain = domain
	result.Webhook.EventType = eventType

	return result.Webhook, nil
}

// UpdateWebhook updates an existing webhook in Mailgun
func (c *mailgunClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *WebhookSpec) (*Webhook, error) {
	params := map[string]interface{}{
		"url": webhook.URL,
	}

	if webhook.Username != nil {
		params["username"] = *webhook.Username
	}
	if webhook.Password != nil {
		params["password"] = *webhook.Password
	}

	body := strings.NewReader(createFormData(params))
	path := fmt.Sprintf("/domains/%s/webhooks/%s", domain, eventType)
	resp, err := c.makeRequest(ctx, "PUT", path, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update webhook")
	}

	var result struct {
		Webhook *Webhook `json:"webhook"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	// Set the domain and event type
	result.Webhook.Domain = domain
	result.Webhook.EventType = eventType

	return result.Webhook, nil
}

// DeleteWebhook deletes a webhook from Mailgun
func (c *mailgunClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	path := fmt.Sprintf("/domains/%s/webhooks/%s", domain, eventType)
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return errors.Wrap(err, "failed to delete webhook")
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return errors.Wrap(err, "failed to handle response")
	}

	return nil
}
