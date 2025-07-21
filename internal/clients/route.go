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

// CreateRoute creates a new route in Mailgun
func (c *mailgunClient) CreateRoute(ctx context.Context, route *RouteSpec) (*Route, error) {
	params := map[string]interface{}{
		"expression": route.Expression,
	}

	if route.Priority != nil {
		params["priority"] = *route.Priority
	}
	if route.Description != nil {
		params["description"] = *route.Description
	}

	// Convert actions to string array format
	if len(route.Actions) > 0 {
		actionStrs := make([]string, len(route.Actions))
		for i, action := range route.Actions {
			if action.Destination != nil {
				actionStrs[i] = fmt.Sprintf("%s(\"%s\")", action.Type, *action.Destination)
			} else {
				actionStrs[i] = action.Type
			}
		}
		params["action"] = strings.Join(actionStrs, ",")
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "POST", "/routes", body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create route")
	}

	var result struct {
		Route *Route `json:"route"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.Route, nil
}

// GetRoute retrieves a route from Mailgun
func (c *mailgunClient) GetRoute(ctx context.Context, id string) (*Route, error) {
	path := fmt.Sprintf("/routes/%s", id)
	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get route")
	}

	var result struct {
		Route *Route `json:"route"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.Route, nil
}

// UpdateRoute updates an existing route in Mailgun
func (c *mailgunClient) UpdateRoute(ctx context.Context, id string, route *RouteSpec) (*Route, error) {
	params := map[string]interface{}{
		"expression": route.Expression,
	}

	if route.Priority != nil {
		params["priority"] = *route.Priority
	}
	if route.Description != nil {
		params["description"] = *route.Description
	}

	// Convert actions to string array format
	if len(route.Actions) > 0 {
		actionStrs := make([]string, len(route.Actions))
		for i, action := range route.Actions {
			if action.Destination != nil {
				actionStrs[i] = fmt.Sprintf("%s(\"%s\")", action.Type, *action.Destination)
			} else {
				actionStrs[i] = action.Type
			}
		}
		params["action"] = strings.Join(actionStrs, ",")
	}

	body := strings.NewReader(createFormData(params))
	path := fmt.Sprintf("/routes/%s", id)
	resp, err := c.makeRequest(ctx, "PUT", path, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update route")
	}

	var result struct {
		Route *Route `json:"route"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return result.Route, nil
}

// DeleteRoute deletes a route from Mailgun
func (c *mailgunClient) DeleteRoute(ctx context.Context, id string) error {
	path := fmt.Sprintf("/routes/%s", id)
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return errors.Wrap(err, "failed to delete route")
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return errors.Wrap(err, "failed to handle response")
	}

	return nil
}
