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
	"net/url"
	"strings"

	"github.com/pkg/errors"
	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
)

// convertDNSRecords copies the client DNSRecord slice into the API DNSRecord
// slice. Both structs share the same field types now, so this is a straight
// field copy.
func convertDNSRecords(clientRecords []DNSRecord) []domaintypes.DNSRecord {
	if clientRecords == nil {
		return nil
	}

	apiRecords := make([]domaintypes.DNSRecord, len(clientRecords))
	for i, r := range clientRecords {
		apiRecords[i] = domaintypes.DNSRecord{
			Name:     r.Name,
			Type:     r.Type,
			Value:    r.Value,
			Priority: r.Priority,
			Valid:    r.Valid,
		}
	}
	return apiRecords
}

// computeDNSVerified returns true if every record in the provided slice has
// a Valid pointer set to "valid", false if at least one record is not valid,
// and nil when there are no records to evaluate.
func computeDNSVerified(records []domaintypes.DNSRecord) *bool {
	if len(records) == 0 {
		return nil
	}

	allValid := true
	for _, record := range records {
		if record.Valid == nil || *record.Valid != "valid" {
			allValid = false
			break
		}
	}
	return &allValid
}

// responseToObservation converts a Mailgun v4 DomainResponse into the API
// DomainObservation, computing DNSVerified across receiving + sending records.
func responseToObservation(r *DomainResponse) *domaintypes.DomainObservation {
	obs := &domaintypes.DomainObservation{}

	if r.Domain != nil {
		obs.ID = r.Domain.Name
		obs.State = r.Domain.State
		obs.CreatedAt = r.Domain.CreatedAt
		obs.SMTPLogin = r.Domain.SMTPLogin
		obs.SMTPPassword = r.Domain.SMTPPassword
	}

	obs.ReceivingDNSRecords = convertDNSRecords(r.ReceivingDNSRecords)
	obs.SendingDNSRecords = convertDNSRecords(r.SendingDNSRecords)

	// DNSVerified is true iff every required receiving + sending record is valid.
	combined := make([]domaintypes.DNSRecord, 0, len(obs.ReceivingDNSRecords)+len(obs.SendingDNSRecords))
	combined = append(combined, obs.ReceivingDNSRecords...)
	combined = append(combined, obs.SendingDNSRecords...)
	obs.DNSVerified = computeDNSVerified(combined)

	return obs
}

// CreateDomain creates a new domain in Mailgun via the v4 Domains API.
func (c *mailgunClient) CreateDomain(ctx context.Context, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
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
	resp, err := c.makeRequestAt(ctx, c.config.V4BaseURL, "POST", "/domains", body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create domain")
	}

	var result DomainResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return responseToObservation(&result), nil
}

// GetDomain retrieves a domain from Mailgun via the v4 Domains API. The
// response includes the receiving and sending DNS records at the top level
// along with their current validity status.
func (c *mailgunClient) GetDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	path := fmt.Sprintf("/domains/%s", url.PathEscape(name))
	resp, err := c.makeRequestAt(ctx, c.config.V4BaseURL, "GET", path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get domain")
	}

	var result DomainResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return responseToObservation(&result), nil
}

// UpdateDomain updates an existing domain in Mailgun via the v4 Domains API.
func (c *mailgunClient) UpdateDomain(ctx context.Context, name string, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
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
	path := fmt.Sprintf("/domains/%s", url.PathEscape(name))
	resp, err := c.makeRequestAt(ctx, c.config.V4BaseURL, "PUT", path, body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to update domain")
	}

	var result DomainResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return responseToObservation(&result), nil
}

// DeleteDomain deletes a domain from Mailgun. The v4 Domains API does not
// expose DELETE, so we use the legacy /v3 endpoint.
func (c *mailgunClient) DeleteDomain(ctx context.Context, name string) error {
	path := fmt.Sprintf("/v3/domains/%s", name)
	resp, err := c.makeRequest(ctx, "DELETE", path, nil)
	if err != nil {
		return errors.Wrap(err, "failed to delete domain")
	}

	if err := c.handleResponse(resp, nil); err != nil {
		return errors.Wrap(err, "failed to handle response")
	}

	return nil
}

// VerifyDomain triggers DNS verification on a domain via the v4 Domains API
// and returns the verification result including the DNS records and their
// current validity status. This is the authoritative source for DNS
// verification information in Mailgun.
func (c *mailgunClient) VerifyDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	path := fmt.Sprintf("/domains/%s/verify", url.PathEscape(name))
	resp, err := c.makeRequestAt(ctx, c.config.V4BaseURL, "PUT", path, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to verify domain")
	}

	var result DomainResponse
	if err := c.handleResponse(resp, &result); err != nil {
		return nil, errors.Wrap(err, "failed to handle response")
	}

	return responseToObservation(&result), nil
}
