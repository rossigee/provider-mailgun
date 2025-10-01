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

package resilience

import (
	"context"

	"github.com/rossigee/provider-mailgun/internal/clients"
	bouncetypes "github.com/rossigee/provider-mailgun/apis/bounce/v1beta1"
	domaintypes "github.com/rossigee/provider-mailgun/apis/domain/v1beta1"
	mailinglisttypes "github.com/rossigee/provider-mailgun/apis/mailinglist/v1beta1"
	routetypes "github.com/rossigee/provider-mailgun/apis/route/v1beta1"
	smtpcredentialtypes "github.com/rossigee/provider-mailgun/apis/smtpcredential/v1beta1"
	templatetypes "github.com/rossigee/provider-mailgun/apis/template/v1beta1"
	webhooktypes "github.com/rossigee/provider-mailgun/apis/webhook/v1beta1"
)

// ResilientClient wraps a Mailgun client with retry and circuit breaker logic
type ResilientClient struct {
	client         clients.Client
	retryConfig    *RetryConfig
	circuitBreaker *CircuitBreaker
}

// NewResilientClient creates a new resilient client wrapper
func NewResilientClient(client clients.Client, retryConfig *RetryConfig) *ResilientClient {
	if retryConfig == nil {
		retryConfig = APIRetryConfig()
	}

	circuitBreaker := NewCircuitBreaker("mailgun-api", 5, retryConfig.MaxBackoff)

	return &ResilientClient{
		client:         client,
		retryConfig:    retryConfig,
		circuitBreaker: circuitBreaker,
	}
}

// SMTP Credential operations with resilience

func (r *ResilientClient) CreateSMTPCredential(ctx context.Context, domain string, credential *smtpcredentialtypes.SMTPCredentialParameters) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	var result *smtpcredentialtypes.SMTPCredentialObservation
	var err error

	retryErr := WithRetry(ctx, "create_smtp_credential", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateSMTPCredential(ctx, domain, credential)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetSMTPCredential(ctx context.Context, domain, login string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	var result *smtpcredentialtypes.SMTPCredentialObservation
	var err error

	retryErr := WithRetry(ctx, "get_smtp_credential", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetSMTPCredential(ctx, domain, login)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) UpdateSMTPCredential(ctx context.Context, domain, login string, password string) (*smtpcredentialtypes.SMTPCredentialObservation, error) {
	var result *smtpcredentialtypes.SMTPCredentialObservation
	var err error

	retryErr := WithRetry(ctx, "update_smtp_credential", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.UpdateSMTPCredential(ctx, domain, login, password)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteSMTPCredential(ctx context.Context, domain, login string) error {
	return WithRetry(ctx, "delete_smtp_credential", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteSMTPCredential(ctx, domain, login)
		})
	})
}

// Template operations with resilience

func (r *ResilientClient) CreateTemplate(ctx context.Context, domain string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	var result *templatetypes.TemplateObservation
	var err error

	retryErr := WithRetry(ctx, "create_template", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateTemplate(ctx, domain, template)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetTemplate(ctx context.Context, domain, name string) (*templatetypes.TemplateObservation, error) {
	var result *templatetypes.TemplateObservation
	var err error

	retryErr := WithRetry(ctx, "get_template", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetTemplate(ctx, domain, name)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) UpdateTemplate(ctx context.Context, domain, name string, template *templatetypes.TemplateParameters) (*templatetypes.TemplateObservation, error) {
	var result *templatetypes.TemplateObservation
	var err error

	retryErr := WithRetry(ctx, "update_template", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.UpdateTemplate(ctx, domain, name, template)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteTemplate(ctx context.Context, domain, name string) error {
	return WithRetry(ctx, "delete_template", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteTemplate(ctx, domain, name)
		})
	})
}

// Domain operations with resilience

func (r *ResilientClient) CreateDomain(ctx context.Context, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	var result *domaintypes.DomainObservation
	var err error

	retryErr := WithRetry(ctx, "create_domain", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateDomain(ctx, domain)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetDomain(ctx context.Context, name string) (*domaintypes.DomainObservation, error) {
	var result *domaintypes.DomainObservation
	var err error

	retryErr := WithRetry(ctx, "get_domain", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetDomain(ctx, name)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) UpdateDomain(ctx context.Context, name string, domain *domaintypes.DomainParameters) (*domaintypes.DomainObservation, error) {
	var result *domaintypes.DomainObservation
	var err error

	retryErr := WithRetry(ctx, "update_domain", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.UpdateDomain(ctx, name, domain)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteDomain(ctx context.Context, name string) error {
	return WithRetry(ctx, "delete_domain", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteDomain(ctx, name)
		})
	})
}

// Mailing List operations with resilience

func (r *ResilientClient) CreateMailingList(ctx context.Context, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	var result *mailinglisttypes.MailingListObservation
	var err error

	retryErr := WithRetry(ctx, "create_mailing_list", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateMailingList(ctx, list)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetMailingList(ctx context.Context, address string) (*mailinglisttypes.MailingListObservation, error) {
	var result *mailinglisttypes.MailingListObservation
	var err error

	retryErr := WithRetry(ctx, "get_mailing_list", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetMailingList(ctx, address)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) UpdateMailingList(ctx context.Context, address string, list *mailinglisttypes.MailingListParameters) (*mailinglisttypes.MailingListObservation, error) {
	var result *mailinglisttypes.MailingListObservation
	var err error

	retryErr := WithRetry(ctx, "update_mailing_list", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.UpdateMailingList(ctx, address, list)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteMailingList(ctx context.Context, address string) error {
	return WithRetry(ctx, "delete_mailing_list", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteMailingList(ctx, address)
		})
	})
}

// Route operations with resilience

func (r *ResilientClient) CreateRoute(ctx context.Context, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	var result *routetypes.RouteObservation
	var err error

	retryErr := WithRetry(ctx, "create_route", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateRoute(ctx, route)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetRoute(ctx context.Context, id string) (*routetypes.RouteObservation, error) {
	var result *routetypes.RouteObservation
	var err error

	retryErr := WithRetry(ctx, "get_route", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetRoute(ctx, id)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) UpdateRoute(ctx context.Context, id string, route *routetypes.RouteParameters) (*routetypes.RouteObservation, error) {
	var result *routetypes.RouteObservation
	var err error

	retryErr := WithRetry(ctx, "update_route", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.UpdateRoute(ctx, id, route)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteRoute(ctx context.Context, id string) error {
	return WithRetry(ctx, "delete_route", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteRoute(ctx, id)
		})
	})
}

// Webhook operations with resilience

func (r *ResilientClient) CreateWebhook(ctx context.Context, domain string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	var result *webhooktypes.WebhookObservation
	var err error

	retryErr := WithRetry(ctx, "create_webhook", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateWebhook(ctx, domain, webhook)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetWebhook(ctx context.Context, domain, eventType string) (*webhooktypes.WebhookObservation, error) {
	var result *webhooktypes.WebhookObservation
	var err error

	retryErr := WithRetry(ctx, "get_webhook", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetWebhook(ctx, domain, eventType)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) UpdateWebhook(ctx context.Context, domain, eventType string, webhook *webhooktypes.WebhookParameters) (*webhooktypes.WebhookObservation, error) {
	var result *webhooktypes.WebhookObservation
	var err error

	retryErr := WithRetry(ctx, "update_webhook", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.UpdateWebhook(ctx, domain, eventType, webhook)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteWebhook(ctx context.Context, domain, eventType string) error {
	return WithRetry(ctx, "delete_webhook", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteWebhook(ctx, domain, eventType)
		})
	})
}

// Bounce operations with resilience

func (r *ResilientClient) CreateBounce(ctx context.Context, domain string, bounce *bouncetypes.BounceParameters) (*bouncetypes.BounceObservation, error) {
	var result *bouncetypes.BounceObservation
	var err error

	retryErr := WithRetry(ctx, "create_bounce", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateBounce(ctx, domain, bounce)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetBounce(ctx context.Context, domain, address string) (*bouncetypes.BounceObservation, error) {
	var result *bouncetypes.BounceObservation
	var err error

	retryErr := WithRetry(ctx, "get_bounce", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetBounce(ctx, domain, address)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteBounce(ctx context.Context, domain, address string) error {
	return WithRetry(ctx, "delete_bounce", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteBounce(ctx, domain, address)
		})
	})
}

// Complaint operations with resilience

func (r *ResilientClient) CreateComplaint(ctx context.Context, domain string, complaint interface{}) (interface{}, error) {
	var result interface{}
	var err error

	retryErr := WithRetry(ctx, "create_complaint", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateComplaint(ctx, domain, complaint)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetComplaint(ctx context.Context, domain, address string) (interface{}, error) {
	var result interface{}
	var err error

	retryErr := WithRetry(ctx, "get_complaint", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetComplaint(ctx, domain, address)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteComplaint(ctx context.Context, domain, address string) error {
	return WithRetry(ctx, "delete_complaint", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteComplaint(ctx, domain, address)
		})
	})
}

// Unsubscribe operations with resilience

func (r *ResilientClient) CreateUnsubscribe(ctx context.Context, domain string, unsubscribe interface{}) (interface{}, error) {
	var result interface{}
	var err error

	retryErr := WithRetry(ctx, "create_unsubscribe", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.CreateUnsubscribe(ctx, domain, unsubscribe)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) GetUnsubscribe(ctx context.Context, domain, address string) (interface{}, error) {
	var result interface{}
	var err error

	retryErr := WithRetry(ctx, "get_unsubscribe", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			result, err = r.client.GetUnsubscribe(ctx, domain, address)
			return err
		})
	})

	if retryErr != nil {
		return nil, retryErr
	}
	return result, nil
}

func (r *ResilientClient) DeleteUnsubscribe(ctx context.Context, domain, address string) error {
	return WithRetry(ctx, "delete_unsubscribe", r.retryConfig, func() error {
		return r.circuitBreaker.Execute(ctx, func() error {
			return r.client.DeleteUnsubscribe(ctx, domain, address)
		})
	})
}
