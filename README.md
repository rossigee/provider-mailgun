# Provider Mailgun

A Crossplane provider for managing Mailgun resources with namespace scoping support for multi-tenancy.

## Features

- **Namespace-scoped Resources**: SMTPCredentials, Templates, MailingLists, Webhooks
- **Cluster-scoped Resources**: Domains, Routes, Suppressions
- **Multi-tenancy Support**: Teams can manage resources in isolated namespaces
- **Complete Mailgun API Coverage**: Domains, routing, templates, credentials, suppressions
- **Regional Support**: Both US and EU Mailgun regions

## Supported Resources

### Namespace-scoped (Multi-tenant)
- **SMTPCredential** - Team-isolated SMTP credentials
- **Template** - Team-specific email templates
- **MailingList** - Team-managed subscriber lists
- **Webhook** - Team-specific event notifications

### Cluster-scoped (Platform-managed)
- **Domain** - Platform-controlled sending/receiving domains
- **Route** - Global email routing rules
- **Bounce/Complaint/Unsubscribe** - Domain-wide suppression management

## Quick Start

1. Install the provider:
```bash
kubectl apply -f examples/provider/config.yaml
```

2. Create a secret with your Mailgun API key:
```bash
kubectl create secret generic mailgun-secret \
  --from-literal=password=your-api-key-here \
  -n crossplane-system
```

3. Configure the provider:
```bash
kubectl apply -f examples/provider/providerconfig.yaml
```

4. Create your first domain:
```bash
kubectl apply -f examples/domain/domain.yaml
```

5. Create SMTP credentials for sending emails:
```bash
kubectl apply -f examples/smtpcredential.yaml
```

6. Create email templates for consistent messaging:
```bash
kubectl apply -f examples/template.yaml
```

## Configuration

The provider supports both US and EU regions:

- **US Region**: `https://api.mailgun.net/v3` (default)
- **EU Region**: `https://api.eu.mailgun.net/v3`

Set the `region` field in your ProviderConfig to specify the region.

## Examples

See the `examples/` directory for complete usage examples of all supported resources.

## Development

This provider is built using the standard Crossplane provider framework.

### Build Requirements
- Go 1.24.5+ (specified in go.mod)
- Docker with buildx support
- Recommended Docker context: `ulta-docker-engine-1`

### Quick Build
```bash
# Build provider binary
go build -o provider cmd/provider/main.go

# Run tests
make test

# Build Docker image
docker build -t provider-mailgun:latest -f cluster/images/provider-mailgun/Dockerfile .

# Build and push to registries
VERSION=v0.1.0 ./build-and-push.sh
```

### Development Setup
```bash
# Clone and setup
git clone <repository>
cd provider-mailgun

# Install dependencies
go mod download

# Generate code
make generate

# Run out-of-cluster for development
make run
```

See `CLAUDE.md` for comprehensive development guidance.

## License

This project is licensed under the Apache License 2.0. See LICENSE for details.
