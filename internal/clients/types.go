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

// Domain represents a Mailgun domain
type Domain struct {
	Name               string      `json:"name"`
	Type               string      `json:"type,omitempty"`
	State              string      `json:"state,omitempty"`
	CreatedAt          string      `json:"created_at,omitempty"`
	SMTPLogin          string      `json:"smtp_login,omitempty"`
	SMTPPassword       string      `json:"smtp_password,omitempty"`
	RequiredDNSRecords []DNSRecord `json:"required_dns_records,omitempty"`
	ReceivingDNSRecords []DNSRecord `json:"receiving_dns_records,omitempty"`
	SendingDNSRecords   []DNSRecord `json:"sending_dns_records,omitempty"`
}

// DomainSpec represents the parameters for creating/updating a domain
type DomainSpec struct {
	Name               string  `json:"name"`
	Type               *string `json:"type,omitempty"`
	ForceDKIMAuthority *bool   `json:"force_dkim_authority,omitempty"`
	DKIMKeySize        *int    `json:"dkim_key_size,omitempty"`
	IPs                []string `json:"ips,omitempty"`
	SMTPPassword       *string `json:"smtp_password,omitempty"`
	SpamAction         *string `json:"spam_action,omitempty"`
	WebScheme          *string `json:"web_scheme,omitempty"`
	Wildcard           *bool   `json:"wildcard,omitempty"`
}

// DNSRecord represents a DNS record
type DNSRecord struct {
	Name     string `json:"name,omitempty"`
	Type     string `json:"record_type,omitempty"`
	Value    string `json:"value,omitempty"`
	Priority *int   `json:"priority,omitempty"`
	Valid    *bool  `json:"valid,omitempty"`
}

// MailingList represents a Mailgun mailing list
type MailingList struct {
	Address         string `json:"address"`
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	AccessLevel     string `json:"access_level,omitempty"`
	ReplyPreference string `json:"reply_preference,omitempty"`
	CreatedAt       string `json:"created_at,omitempty"`
	MembersCount    int    `json:"members_count,omitempty"`
}

// MailingListSpec represents the parameters for creating/updating a mailing list
type MailingListSpec struct {
	Address         string               `json:"address"`
	Name            *string              `json:"name,omitempty"`
	Description     *string              `json:"description,omitempty"`
	AccessLevel     *string              `json:"access_level,omitempty"`
	ReplyPreference *string              `json:"reply_preference,omitempty"`
	Members         []MailingListMember  `json:"members,omitempty"`
}

// MailingListMember represents a member of a mailing list
type MailingListMember struct {
	Address    string            `json:"address"`
	Name       *string           `json:"name,omitempty"`
	Vars       map[string]string `json:"vars,omitempty"`
	Subscribed *bool             `json:"subscribed,omitempty"`
}

// Route represents a Mailgun route
type Route struct {
	ID          string        `json:"id,omitempty"`
	Priority    int           `json:"priority,omitempty"`
	Description string        `json:"description,omitempty"`
	Expression  string        `json:"expression,omitempty"`
	Actions     []RouteAction `json:"actions,omitempty"`
	CreatedAt   string        `json:"created_at,omitempty"`
}

// RouteSpec represents the parameters for creating/updating a route
type RouteSpec struct {
	Priority    *int          `json:"priority,omitempty"`
	Description *string       `json:"description,omitempty"`
	Expression  string        `json:"expression"`
	Actions     []RouteAction `json:"actions"`
}

// RouteAction represents an action for a route
type RouteAction struct {
	Type        string  `json:"action"`
	Destination *string `json:"destination,omitempty"`
}

// Webhook represents a Mailgun webhook
type Webhook struct {
	ID        string `json:"id,omitempty"`
	EventType string `json:"event_type,omitempty"`
	URL       string `json:"url,omitempty"`
	Username  string `json:"username,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Domain    string `json:"domain,omitempty"`
}

// WebhookSpec represents the parameters for creating/updating a webhook
type WebhookSpec struct {
	EventType string  `json:"event_type"`
	URL       string  `json:"url"`
	Username  *string `json:"username,omitempty"`
	Password  *string `json:"password,omitempty"`
}

// SMTPCredential represents a Mailgun SMTP credential
type SMTPCredential struct {
	Login     string `json:"login"`
	Password  string `json:"password,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	State     string `json:"state,omitempty"`
}

// SMTPCredentialSpec represents the parameters for creating an SMTP credential
type SMTPCredentialSpec struct {
	Login    string  `json:"login"`
	Password *string `json:"password,omitempty"`
}

// Template represents a Mailgun email template
type Template struct {
	Name         string            `json:"name"`
	Description  string            `json:"description,omitempty"`
	CreatedAt    string            `json:"created_at,omitempty"`
	CreatedBy    string            `json:"created_by,omitempty"`
	Version      *TemplateVersion  `json:"version,omitempty"`
	Versions     []TemplateVersion `json:"versions,omitempty"`
}

// TemplateSpec represents the parameters for creating/updating a template
type TemplateSpec struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Template    *string `json:"template,omitempty"`
	Engine      *string `json:"engine,omitempty"`
	Comment     *string `json:"comment,omitempty"`
	Tag         *string `json:"tag,omitempty"`
}

// TemplateVersion represents a version of a template
type TemplateVersion struct {
	Tag       string `json:"tag,omitempty"`
	Engine    string `json:"engine,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	Comment   string `json:"comment,omitempty"`
	Active    bool   `json:"active,omitempty"`
	Template  string `json:"template,omitempty"`
}

// Bounce represents a bounce suppression entry
type Bounce struct {
	Address   string `json:"address"`
	Code      string `json:"code,omitempty"`
	Error     string `json:"error,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// BounceSpec represents the parameters for creating a bounce entry
type BounceSpec struct {
	Address string  `json:"address"`
	Code    *string `json:"code,omitempty"`
	Error   *string `json:"error,omitempty"`
}

// Complaint represents a complaint suppression entry
type Complaint struct {
	Address   string `json:"address"`
	CreatedAt string `json:"created_at,omitempty"`
}

// ComplaintSpec represents the parameters for creating a complaint entry
type ComplaintSpec struct {
	Address string `json:"address"`
}

// Unsubscribe represents an unsubscribe suppression entry
type Unsubscribe struct {
	Address   string `json:"address"`
	Tags      string `json:"tags,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// UnsubscribeSpec represents the parameters for creating an unsubscribe entry
type UnsubscribeSpec struct {
	Address string  `json:"address"`
	Tags    *string `json:"tags,omitempty"`
}
