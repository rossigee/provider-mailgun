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

// CreateMailingList creates a new mailing list in Mailgun
func (c *mailgunClient) CreateMailingList(ctx context.Context, list *MailingListSpec) (*MailingList, error) {
	params := map[string]interface{}{
		"address": list.Address,
	}

	if list.Name != nil {
		params["name"] = *list.Name
	}
	if list.Description != nil {
		params["description"] = *list.Description
	}
	if list.AccessLevel != nil {
		params["access_level"] = *list.AccessLevel
	}
	if list.ReplyPreference != nil {
		params["reply_preference"] = *list.ReplyPreference
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "POST", "/lists", body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create mailing list")
	}

	var result struct {
		List *MailingList `json:"list"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.List, nil
}

// GetMailingList retrieves a mailing list from Mailgun
func (c *mailgunClient) GetMailingList(ctx context.Context, address string) (*MailingList, error) {
	path := fmt.Sprintf("/lists/%s", address)
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get mailing list")
	}

	var result struct {
		List *MailingList `json:"list"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.List, nil
}

// UpdateMailingList updates an existing mailing list in Mailgun
func (c *mailgunClient) UpdateMailingList(ctx context.Context, address string, list *MailingListSpec) (*MailingList, error) {
	params := map[string]interface{}{}

	if list.Name != nil {
		params["name"] = *list.Name
	}
	if list.Description != nil {
		params["description"] = *list.Description
	}
	if list.AccessLevel != nil {
		params["access_level"] = *list.AccessLevel
	}
	if list.ReplyPreference != nil {
		params["reply_preference"] = *list.ReplyPreference
	}

	body := strings.NewReader(createFormData(params))
	path := fmt.Sprintf("/lists/%s", address)
	resp, err := c.makeRequest(ctx, "PUT", path, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update mailing list")
	}

	var result struct {
		List *MailingList `json:"list"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.List, nil
}

// DeleteMailingList deletes a mailing list from Mailgun
func (c *mailgunClient) DeleteMailingList(ctx context.Context, address string) error {
	path := fmt.Sprintf("/lists/%s", address)
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return errors.Wrap(err, "failed to delete mailing list")
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return errors.Wrap(err, "failed to handle response")
	}

	return nil
}
