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
	"fmt"
	"io"
	"strings"
)

// CreateSMTPCredential creates a new SMTP credential for a domain
func (c *mailgunClient) CreateSMTPCredential(ctx context.Context, domain string, credential *SMTPCredentialSpec) (*SMTPCredential, error) {
	path := fmt.Sprintf("/domains/%s/credentials", domain)

	params := map[string]interface{}{
		"login": credential.Login,
	}
	// Let Mailgun generate the password if none is provided
	if credential.Password != nil {
		params["password"] = *credential.Password
	}
	// If no password provided, Mailgun will generate one

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "POST", path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create SMTP credential: %w", err)
	}

	// Read the response body once
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Debug: Log the actual response body to understand what Mailgun returns
	fmt.Printf("DEBUG: CreateSMTPCredential response body: %s\n", string(responseBody))

	// Try to parse as credential response first (if Mailgun returns the credential directly)
	var credentialResponse struct {
		Login     string `json:"login"`
		Password  string `json:"password"`
		CreatedAt string `json:"created_at"`
		State     string `json:"state"`
	}

	if err := json.Unmarshal(responseBody, &credentialResponse); err == nil && credentialResponse.Login != "" {
		return &SMTPCredential{
			Login:     credentialResponse.Login,
			Password:  credentialResponse.Password,
			CreatedAt: credentialResponse.CreatedAt,
			State:     credentialResponse.State,
		}, nil
	}

	// Parse as credentials response (new API behavior with generated passwords)
	var createResponse struct {
		Message     string            `json:"message"`
		Note        string            `json:"note"`
		Credentials map[string]string `json:"credentials"`
	}
	if err := json.Unmarshal(responseBody, &createResponse); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if we have credentials with the password
	if createResponse.Credentials != nil {
		if password, exists := createResponse.Credentials[credential.Login]; exists {
			fmt.Printf("DEBUG: Found password in credentials response for %s (length: %d)\n", credential.Login, len(password))
			return &SMTPCredential{
				Login:    credential.Login,
				Password: password,
				State:    "active", // Newly created credentials are active
			}, nil
		}
	}

	// Verify the creation was successful by checking the message
	if !strings.Contains(createResponse.Message, "Created") && !strings.Contains(createResponse.Message, "credentials pair") {
		return nil, fmt.Errorf("unexpected response from Mailgun API: %s", createResponse.Message)
	}

	// If we only got a success message, we need to fetch the credential to get the password
	// This happens when Mailgun generates the password but credentials object is empty
	fmt.Printf("DEBUG: No password in credentials response, falling back to GET\n")
	return c.GetSMTPCredential(ctx, domain, credential.Login)
}

// GetSMTPCredential retrieves an SMTP credential
func (c *mailgunClient) GetSMTPCredential(ctx context.Context, domain, login string) (*SMTPCredential, error) {
	// List all credentials and find the matching one
	path := fmt.Sprintf("/domains/%s/credentials", domain)

	resp, err := c.makeRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMTP credentials: %w", err)
	}

	var result struct {
		Items []SMTPCredential `json:"items"`
	}
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to handle response: %w", err)
	}

	// Find the credential with matching login
	for _, cred := range result.Items {
		if cred.Login == login {
			return &cred, nil
		}
	}

	return nil, fmt.Errorf("credential %s not found (404)", login)
}

// UpdateSMTPCredential updates the password for an SMTP credential
func (c *mailgunClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*SMTPCredential, error) {
	path := fmt.Sprintf("/domains/%s/credentials/%s", domain, login)

	params := map[string]interface{}{
		"password": password,
	}

	body := strings.NewReader(createFormData(params))
	resp, err := c.makeRequest(ctx, "PUT", path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to update SMTP credential: %w", err)
	}

	var result SMTPCredential
	if err := c.handleResponse(resp, &result); err != nil {
		// Some endpoints return empty response on success, so we handle that
		return &SMTPCredential{
			Login: login,
			State: "active",
		}, nil
	}

	return &result, nil
}

// DeleteSMTPCredential deletes an SMTP credential
func (c *mailgunClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	path := fmt.Sprintf("/domains/%s/credentials/%s", domain, login)

	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return fmt.Errorf("failed to delete SMTP credential: %w", err)
	}

	// For DELETE operations, we usually don't need to parse the response
	var result interface{}
	if err := c.handleResponse(resp, &result); err != nil {
		// Many DELETE endpoints return empty responses, so we ignore parse errors
		return nil
	}

	return nil
}
