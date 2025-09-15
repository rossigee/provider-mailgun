# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

# Mailgun Crossplane Provider

## Architecture Overview

This is a **Crossplane v2 managed resource provider** for Mailgun integration with namespaced resources:

- **Crossplane v2 Architecture**: Namespace-scoped resources only with `.m.` API group naming (e.g., `mailgun.m.crossplane.io`)
- **Core Resources**: Domain, MailingList, Route, Webhook, Template, SMTPCredential, and Bounce management with full CRUD operations
- **External Client Pattern**: Mailgun API abstraction with interface-based design
- **Cross-Resource References**: Webhooks reference Domains using Kubernetes-native `spec.domainRef`
- **Provider Configuration**: Authentication via ProviderConfig with Kubernetes secret references
- **Multi-tenancy**: Namespace isolation for secure multi-tenant deployments

**Key Directory Structure**:
- `apis/` - CRD definitions (Domain, MailingList, Route, Webhook, etc.)
- `internal/clients/` - Mailgun API client implementation
- `internal/controller/` - Crossplane managed resource controllers
- `examples/` - Complete usage examples and production setups
- `package/` - Crossplane packaging and metadata

## Development Commands

### Essential Build Commands
```bash
# Code generation (ALWAYS run after API changes)
make generate

# Build and test
make build
make test

# Local development
make run              # Run provider out-of-cluster
make install-crds     # Install CRDs into cluster

# Packaging
make docker-build
make xpkg-build       # Build Crossplane package
```

## Critical Implementation Patterns

### Standard Crossplane Resource Controller
All controllers follow this pattern:
```go
func (c *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error)
func (c *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error)
func (c *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error)
func (c *external) Delete(ctx context.Context, mg resource.Managed) error
```

### Cross-Resource References
Webhooks reference Domains using Crossplane's standard pattern:
```yaml
spec:
  domainRef:
    name: my-domain-resource
```

### Status Conditions
Use Crossplane's standard conditions:
- `xpv1.Available()` - Resource ready
- `xpv1.Creating()` - Resource being created
- `xpv1.Deleting()` - Resource being deleted

### Error Handling
- Always wrap errors with context using `errors.Wrap()`
- Detect 404s to determine resource existence
- Handle Mailgun API rate limits and failures gracefully

## Mailgun Client Usage

The Mailgun client (`internal/clients/mailgun.go`) provides abstracted access to Mailgun API:
- **Domains**: CRUD operations, DNS record management, tracking settings
- **MailingLists**: Creation, member management, access control
- **Routes**: Email routing rule configuration
- **Webhooks**: Event notification endpoint setup
- **Authentication**: API key-based authentication via ProviderConfig

## API Design Conventions

### Field Validation
Use kubebuilder validation tags extensively:
```go
// +kubebuilder:validation:Required
// +kubebuilder:validation:Enum=US;EU
// +kubebuilder:validation:Pattern="^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$"
```

### Optional Fields
Use pointer types for optional fields:
```go
Region       *string `json:"region,omitempty"`
Description  *string `json:"description,omitempty"`
```

### Status Reporting
Include observed state in status:
```go
type DomainObservation struct {
    ID              string      `json:"id,omitempty"`
    State           string      `json:"state,omitempty"`
    RequiredDNSRecords []DNSRecord `json:"requiredDnsRecords,omitempty"`
}
```

## Current Implementation Status

**✅ Complete - Crossplane v2 Provider**:
- ✅ **Crossplane v2 Architecture**: Namespaced resources only with .m. API group naming
- ✅ **v1beta1 APIs**: All 7 resource types using namespaced v1beta1 APIs
- ✅ **Breaking Change Migration**: Removed all v1alpha1 cluster-scoped APIs in v0.11.0
- ✅ Multi-tenancy support through namespace isolation
- ✅ Project structure and build configuration
- ✅ API definitions for all resource types (Domain, MailingList, Route, Webhook, Template, SMTPCredential, Bounce)
- ✅ Mailgun client interface and HTTP client implementation
- ✅ Provider configuration and main entry point
- ✅ Example manifests for all resources (updated to v1beta1 namespaced)
- ✅ DeepCopy code generation for all API types
- ✅ Crossplane managed resource methods generation
- ✅ All 7 controllers implementation (functional)
- ✅ Comprehensive test suite (133+ tests, all passing)
- ✅ Complete integration test coverage for multi-resource workflows
- ✅ Error handling and network failure test coverage
- ✅ HTTP client reliability improvements with retry logic and proper body handling
- ✅ Test performance optimizations (sub-second execution)
- ✅ Docker build infrastructure and CI/CD workflows
- ✅ Docker image build process (Go 1.24.5 compatible)
- ✅ Kubernetes deployment manifests for golder-secops cluster
- ✅ Health probe endpoints (/healthz and /readyz on port 8080)
- ✅ Improved logging configuration for production deployments
- ✅ Lint-compliant codebase (0 issues)

**✅ Production Deployment**:
- Successfully deployed to golder-secops cluster
- Docker image: `ghcr.io/rossigee/provider-mailgun:v0.12.0` (current - Crossplane v2 with runtime v1.21.0)
- All controllers operational with comprehensive test coverage
- **BREAKING CHANGE**: v0.11.0 removed all v1alpha1 cluster-scoped APIs
- **Test Coverage**: 36.3% overall (133 test functions across 22 test files)
  - HTTP Client: 55.7% coverage (core networking and API communication)
  - Controllers: 47.5-58.7% coverage across all 6 controllers
  - Utility modules: 92.7-100% coverage (metrics, tracing, errors, health)
  - Comprehensive integration scenarios and error handling coverage

## Build and Deployment Process

### ⚠️ Critical Build Requirements
- **Go Version**: Go 1.24.5+ required (specified in go.mod)
- **Docker Context**: Use `ulta-docker-engine-1` for optimal build performance
- **Dockerfile**: Updated to use `golang:1.24.5` base image for latest bugfixes

### Standard Build Commands
```bash
# Build provider binary directly (fastest method)
go build -o provider cmd/provider/main.go

# Run comprehensive test suite
make test

# Generate code (DeepCopy, managed resources, CRDs)
make generate

# Docker build (requires Go 1.23 compatible Dockerfile)
docker build -t provider-mailgun:latest -f cluster/images/provider-mailgun/Dockerfile .

# Build and push to multiple registries
./build-and-push.sh
```

### Docker Build Process
```bash
# Switch to optimal Docker context
docker context use ulta-docker-engine-1

# Build image locally
docker build -t provider-mailgun:test -f cluster/images/provider-mailgun/Dockerfile .

# Build and push to Harbor (internal registry)
VERSION=v0.10.0 ./build-and-push.sh

# Build and push to both Harbor and GHCR
VERSION=v0.10.0 PUSH_EXTERNAL=true ./build-and-push.sh

# Build with Crossplane package
VERSION=v0.10.0 BUILD_PACKAGE=true ./build-and-push.sh
```

### Environment Variables for Registry Override
- **`VERSION`** - Image version tag (default: `dev`)
- **`PUSH_EXTERNAL`** - Push to GHCR (GitHub Container Registry) (`true`/`false`)
- **`BUILD_PACKAGE`** - Build Crossplane .xpkg package (`true`/`false`)
- **`PLATFORMS`** - Build platforms (default: `linux/amd64,linux/arm64`)
- **`XPKG_REG_ORGS`** - Override crossplane package registry (default: `xpkg.upbound.io/crossplane-contrib`)
- **`REGISTRY`** - Registry location (now using ghcr.io/rossigee)

### Deployment to golder-secops Cluster
The provider is deployed via Flux GitOps:
- **Manifest**: `/home/rossg/clients/golder/infrastructure/flux-golder/clusters/golder-secops/crossplane-providers/provider-mailgun.yaml`
- **Registry**: `ghcr.io/rossigee/provider-mailgun:v0.10.0`
- **Runtime**: Uses shared `provider-runtime` DeploymentRuntimeConfig
- **Health Probes**: Kubernetes liveness and readiness probes configured for port 8080
- **Secrets**: Uses `harbor-credentials` for image pull authentication

## Recent Improvements (2025-09-15)

### Crossplane Runtime Update (v0.12.0)
- **Updated crossplane-runtime**: Upgraded from v1.20.0 to v1.21.0-rc.0 to address ProviderConfigUsage namespace creation issues
- **Fixed Package Building**: Resolved "not exactly one package meta type" error by removing duplicate package.yaml file
- **Generated Code Completion**: Added missing v1beta1 managed resource code for all resource types
- **Deployment Status**: Successfully deployed to production with updated runtime
- **Build System**: Fixed Crossplane package (.xpkg) building process with embedded Docker images
- **All Tests Passing**: 133+ test functions continue to pass with 36.3% overall coverage

## Previous Improvements (2025-08-14)

### HTTP Service Consolidation (v0.10.1)
- **Single HTTP Server**: Consolidated metrics (previously port 9090) and health checks onto single port 8080
- **Controller-Runtime Integration**: Migrated to controller-runtime's built-in health check system using `AddHealthzCheck`/`AddReadyzCheck`
- **Simplified Architecture**: Removed custom HTTP server implementation (~200 lines of code)
- **Standard Endpoints**: Health checks available at `/healthz` and `/readyz`, metrics at `/metrics` - all on port 8080
- **Test Updates**: Updated health check tests to use new controller-runtime interface

### HTTP Client Reliability Fixes
- **Fixed Request Body Handling**: Resolved race condition in HTTP retry logic that caused "ContentLength=X with Body length 0" errors
- **Optimized Test Performance**: Reduced test retry delays from 2s to 50ms for testing, improving test execution time by 95%
- **Enhanced Error Handling**: Added proper error checking for response body closing operations
- **Test Suite Stability**: Fixed SMTP credential test response format mismatch

### Code Quality Improvements
- **Eliminated Lint Errors**: Fixed all errcheck and ineffassign lint warnings
- **Improved Request Reliability**: Restructured body reading logic to prevent multiple consumption issues
- **Testing Optimizations**: Made retry behavior context-aware (shorter delays during testing)

### Test Coverage Status
- **Overall Coverage**: 34.8% (up from 27.4%)
- **HTTP Client Coverage**: 51.4% (critical path well-tested)
- **Performance**: Tests now complete in <1 second vs previous 28+ seconds

## Resource Relationships

- **Domains** are independent resources
- **MailingLists** are associated with domains via email address
- **Routes** apply to all domains or can be domain-specific via expression
- **Webhooks** reference specific domains via `domainRef`

## Regional Support

The provider supports both Mailgun regions:
- **US**: `https://api.mailgun.net/v3` (default)
- **EU**: `https://api.eu.mailgun.net/v3`

Configure via `region` field in ProviderConfig or explicit `apiBaseURL`.
