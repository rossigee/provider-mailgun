# Provider Mailgun

A Crossplane provider for managing Mailgun resources with namespace scoping support for multi-tenancy and intelligent credential rotation. Updated with modern CI/CD pipeline and GHCR publishing.

## Features

- **Namespace-scoped Resources**: SMTPCredentials, Templates, MailingLists, Webhooks
- **Cluster-scoped Resources**: Domains, Routes, Suppressions
- **Multi-tenancy Support**: Teams can manage resources in isolated namespaces
- **Complete Mailgun API Coverage**: Domains, routing, templates, credentials, suppressions
- **Credential Rotation Strategy**: Handles write-only SMTP credentials with automatic rotation
- **Unified Regional Support**: Single API key works across US and EU regions

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

2. Create a secret with your unified Mailgun API key:
```bash
kubectl create secret generic mailgun-credentials \
  --from-literal=credentials=your-unified-api-key-here \
  -n crossplane-system
```

3. Configure the provider (single ProviderConfig for all regions):
```yaml
apiVersion: mailgun.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: mailgun
  namespace: crossplane-system
spec:
  region: US  # Works for both US and EU with unified API key
  apiBaseURL: https://api.mailgun.net/v3
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: mailgun-credentials
      key: credentials
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

### Simplified Regional Configuration

The provider uses a unified API key that works across both US and EU regions:

- **Unified ProviderConfig**: Single configuration handles all domains regardless of region
- **API Key**: Works interchangeably between US (`https://api.mailgun.net/v3`) and EU (`https://api.eu.mailgun.net/v3`) endpoints
- **Automatic Routing**: Provider determines appropriate endpoint based on domain configuration

### SMTP Credential Rotation Strategy

Due to Mailgun's write-only SMTP credentials API, the provider implements an intelligent rotation strategy:

- **Initial Creation**: Creates new SMTP credentials and stores in Kubernetes Secret
- **Subsequent Operations**: Checks for existing secrets to determine credential status
- **Rotation**: Automatically deletes and recreates credentials when needed
- **Secret Management**: Maintains connection details in writeConnectionSecretToRef

## Examples

See the `examples/` directory for complete usage examples of all supported resources.

## Development

This provider is built using the standard Crossplane provider framework with enhanced SMTP credential management.

### Build Requirements
- Go 1.24.5+ (specified in go.mod)
- Docker with buildx support
- Recommended Docker context: `ulta-docker-engine-1`

### Quick Build
```bash
# Build provider binary
go build -o provider cmd/provider/main.go

# Run tests with rotation strategy coverage
make test

# Build Crossplane package (.xpkg)
crossplane xpkg build -f package/ --embed-runtime-image=ghcr.io/rossigee/provider-mailgun:v0.8.3

# Build and push to registries
VERSION=v0.8.3 ./build-and-push.sh
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

### Testing

The provider includes comprehensive test coverage for:
- **SMTP Credential Rotation**: Tests write-only credential handling
- **Secret Management**: Kubernetes secret integration testing
- **ProviderConfig Usage**: Namespace-scoped usage tracking
- **Mock Client**: Complete Mailgun API simulation
- **Integration Scenarios**: Multi-resource workflow testing
- **Error Handling**: Network failures, malformed responses, context cancellation
- **Controller Coverage**: Enhanced test coverage across all 6 controllers

Current test coverage: **36.3%** (133 test functions) with focus on critical paths and HTTP client reliability (55.7%).

See `CLAUDE.md` for comprehensive development guidance.

## License

This project is licensed under the Apache License 2.0. See LICENSE for details.
