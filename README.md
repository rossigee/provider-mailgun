# Provider Mailgun

[![CI](https://img.shields.io/github/actions/workflow/status/rossigee/provider-mailgun/ci.yml?branch=master)][build]
[![Version](https://img.shields.io/github/v/release/rossigee/provider-mailgun)][releases]
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

[build]: https://github.com/rossigee/provider-mailgun/actions/workflows/ci.yml
[releases]: https://github.com/rossigee/provider-mailgun/releases

A Crossplane v2 provider for managing Mailgun resources with complete namespace isolation for multi-tenancy.

## Container Registry

- **Primary**: `ghcr.io/rossigee/provider-mailgun:v0.18.0`

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
kubectl crossplane install provider ghcr.io/rossigee/provider-mailgun:v0.18.0
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
  region: US  # Required: US or EU
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

#### Test Email

To verify SMTP credentials are working, set an annotation on the SMTPCredential:

```bash
kubectl annotate smtpcredential mailer mailgun.crossplane.io/test-email-to=you@example.com --overwrite
```

On the next reconcile, the provider will send a test email to that address using the credential's login as the sender. On success, the annotation is automatically cleared. On failure, a warning event is emitted and the annotation is retained for retry.

## DNS Verification Flow

When you create a `Domain` resource, Mailgun responds with a list of
DNS records you must configure for the domain to be `active` and able to
send/receive mail. The provider surfaces these records in three ways so
you can choose the workflow that fits your team:

### 1. Diagnostic events and status fields (always on)

Every `Domain` has `.status.atProvider.requiredDnsRecords` populated with
the exact list Mailgun is waiting on, plus a `STATE` column in
`kubectl get domain` and a `DNS-VERIFIED` column. While records are
unverified, the controller emits a Normal `DNSRecordsRequired` event on
each reconcile with the full per-record manifest embedded in the message
so `kubectl describe domain <name>` is enough to copy/paste into your DNS
provider:

```bash
$ kubectl describe domain bankrut-info -n mailgun-resources
Events:
  Type    Reason              Age   From             Message
  ----    ------              ----  ----             -------
  Normal  DNSRecordsRequired  2m    provider-mailgun Mailgun requires 4 DNS record(s) to verify this domain.
                                              Each record below appears as `<type>-<hash>=<expected value>`:
                                                mx-7b2e9f0a=mxa.mailgun.org
                                                mx-7b2e9f0a=mxb.mailgun.org
                                                txt-1a2b3c4d=v=spf1 include:mailgun.org ~all
                                                txt-5d6e7f80=k=rsa; p=MIGfMA0GCSqGSIb3DQEBA...
```

The controller also calls Mailgun's `/v4/domains/{name}/verify`
endpoint on the slow path (DNS unverified) and sets
`crossplane.io/poll-interval=5m` so the next reconcile runs on a
Mailgun-friendly cadence rather than controller-runtime's default, which
would burn API quota.

### 2. Opt-in: DNS-records ConfigMap output

Set the annotation `mailgun.crossplane.io/dns-configmap: "true"` on the
Domain. The controller will create/update a ConfigMap named
`<domain>-dns-records` in the Domain's namespace, containing three keys
you can consume from any external automation:

| Key              | Format                          | Use case                          |
|------------------|---------------------------------|-----------------------------------|
| `records.yaml`   | YAML list of records            | `kubectl apply`, generic scripts  |
| `terraform.tf`   | HCL for `hashicorp/dns` provider| Terraform-based DNS automation    |
| `bind-zone.txt`  | BIND master-file fragment       | `nsupdate`, manual zone includes  |

The ConfigMap is owned by the Domain (OwnerReference) so it is
garbage-collected when the Domain is deleted. To grant the provider
permission to write ConfigMaps, apply `examples/provider/configmap-rbac.yaml`.
The RBAC is opt-in: without the annotation, the controller never calls
the apiserver for ConfigMaps.

### 3. Opt-in: external-dns integration

The controller writes the
`external-dns.alpha.kubernetes.io/hostname` annotation onto every
Domain that does not have `mailgun.crossplane.io/disable-external-dns:
"true"`. With `external-dns` installed in the cluster and a Mailgun
provider configured, the records will be pushed to Mailgun's DNS
automatically. See `examples/domain/with-external-dns.yaml` for a
complete Deployment + RBAC + Secret setup.

### 4. Opt-in: live DNS propagation probe

Set `mailgun.crossplane.io/dns-probe: "true"` on the Domain. The
controller will query 8.8.8.8, 1.1.1.1, and 9.9.9.9 for each record
and emit one of these events per record:

| Event                  | Meaning                                              |
|------------------------|------------------------------------------------------|
| `DNSRecordMatches`     | Record is propagated and value matches Mailgun       |
| `DNSValueMismatch`     | Record is propagated but value differs               |
| `DNSNotPropagated`     | Record is not yet visible in DNS                     |
| `SPFNeedsMerge`        | Existing SPF detected, manual merge required        |
| `DNSProbeError`        | Resolver timeout or unexpected RCODE                 |

This requires outbound UDP/TCP 53 from the provider Pod to the public
internet. Disable egress controls accordingly.

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
