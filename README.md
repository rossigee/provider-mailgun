# Provider Mailgun

[![CI](https://img.shields.io/github/actions/workflow/status/rossigee/provider-mailgun/ci.yml?branch=master)][build]
[![Version](https://img.shields.io/github/v/release/rossigee/provider-mailgun)][releases]
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[build]: https://github.com/rossigee/provider-mailgun/actions/workflows/ci.yml
[releases]: https://github.com/rossigee/provider-mailgun/releases

A Crossplane v2 provider for managing Mailgun resources with complete namespace isolation for multi-tenancy.

## Container Registry

- **Primary**: `ghcr.io/rossigee/provider-mailgun:v0.15.0`

## Overview

A Crossplane v2 provider for managing Mailgun resources including domains, mailing lists, routes, webhooks, templates, and SMTP credentials.

## Features

- **Crossplane v2 Architecture**: Complete namespace-scoped resource management
- **Multi-tenancy**: All resources isolated by namespace for team separation
- **Comprehensive Mailgun API Coverage**: Domains, routing, templates, credentials, webhooks, mailing lists, and suppression lists (bounce, complaint, unsubscribe)
- **Credential Rotation Strategy**: Handles write-only SMTP credentials with automatic rotation
- **Unified Regional Support**: Single API key works across US and EU regions
- **Health Monitoring**: Built-in health probes for Kubernetes liveness and readiness checks
- **Secure by Default**: URL-encoded path parameters, removed sensitive data logging

## Getting Started

### Prerequisites

- Kubernetes cluster with Crossplane installed
- Mailgun account with API access
- Mailgun API key (unified key works for both US and EU regions)

### Installation

```bash
kubectl crossplane install provider ghcr.io/rossigee/provider-mailgun:v0.15.0
```

### Configuration

Create a secret with your Mailgun API key:

```bash
kubectl create secret generic mailgun-credentials \
  --from-literal=credentials=your-unified-api-key-here \
  -n crossplane-system
```

Create the ProviderConfig:

```yaml
apiVersion: mailgun.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
  namespace: crossplane-system
spec:
  region: US
  credentials:
    source: Secret
    secretRef:
      name: mailgun-credentials
      namespace: crossplane-system
      key: credentials
```

## Usage

### Create a Domain

```yaml
apiVersion: domain.mailgun.m.crossplane.io/v1beta1
kind: Domain
metadata:
  name: example-com
  namespace: production
spec:
  forProvider:
    name: example.com
    spamAction: tag
  providerConfigRef:
    name: default
```

### Create SMTP Credentials

```yaml
apiVersion: smtpcredential.mailgun.m.crossplane.io/v1beta1
kind: SMTPCredential
metadata:
  name: mailer
  namespace: production
spec:
  forProvider:
    parentDomainRef:
      name: example-com
    passwordSecretRef:
      key: password
      secretName: smtp-password
      namespace: production
  writeConnectionSecretToRef:
    name: mailer-credentials
    namespace: production
  providerConfigRef:
    name: default
```

## Resource Types

| Resource | API Version | Description |
|----------|-------------|-------------|
| Domain | `domain.mailgun.m.crossplane.io/v1beta1` | Sending/receiving domains |
| MailingList | `mailinglist.mailgun.m.crossplane.io/v1beta1` | Subscriber lists |
| Route | `route.mailgun.m.crossplane.io/v1beta1` | Email routing rules |
| Webhook | `webhook.mailgun.m.crossplane.io/v1beta1` | Event notifications |
| Template | `template.mailgun.m.crossplane.io/v1beta1` | Email templates |
| SMTPCredential | `smtpcredential.mailgun.m.crossplane.io/v1beta1` | SMTP credentials |
| Bounce | `bounce.mailgun.m.crossplane.io/v1beta1` | Bounce suppressions |
| Complaint | `complaint.mailgun.m.crossplane.io/v1beta1` | Complaint suppressions |
| Unsubscribe | `unsubscribe.mailgun.m.crossplane.io/v1beta1` | Unsubscribe suppressions |

## Unsupported Mailgun APIs

The following Mailgun APIs are not yet supported by this provider:

- IP address management and warmup
- Click and open tracking configuration
- Email validation and recipient verification
- Message metadata and acceptance inspection
- List validation services

## Development

```bash
# Build the provider
make build

# Run tests
make test

# Lint code
make lint

# Generate CRDs
make generate
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

provider-mailgun is under the Apache 2.0 license.
