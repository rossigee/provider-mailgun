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
- ✅ **v1beta1 APIs**: All 9 resource types using namespaced v1beta1 APIs
- ✅ **Breaking Change Migration**: Removed all v1alpha1 cluster-scoped APIs in v0.11.0
- ✅ Multi-tenancy support through namespace isolation
- ✅ Project structure and build configuration
- ✅ API definitions for all resource types (Domain, MailingList, Route, Webhook, Template, SMTPCredential, Bounce, Complaint, Unsubscribe)
- ✅ Mailgun client interface and HTTP client implementation
- ✅ Provider configuration and main entry point
- ✅ Example manifests for all resources (updated to v1beta1 namespaced)
- ✅ DeepCopy code generation for all API types
- ✅ Crossplane managed resource methods generation
- ✅ All 9 controllers implementation (functional with dedicated unit tests for all controllers)
- ✅ Comprehensive test suite (173 test functions across 29 test files, all passing)
- ✅ Complete integration test coverage for multi-resource workflows
- ✅ Error handling and network failure test coverage
- ✅ HTTP client reliability improvements with retry logic and proper body handling
- ✅ Test performance optimizations (sub-second execution)
- ✅ Docker build infrastructure and CI/CD workflows
- ✅ Docker image build process (Go 1.27.1 compatible)
- ✅ Health probe endpoints (/healthz and /readyz on port 8080)
- ✅ Improved logging configuration for production deployments
- ✅ ManagementPolicies support (Observe-only enforcement)
- ✅ State metrics (`MRStateMetrics`) reporting Ready/Synced counts
- ✅ Lint-compliant codebase (0 issues)

**✅ Production Deployment**:
- Docker image: `ghcr.io/rossigee/provider-mailgun:v0.23.2` (current - Crossplane v2 with crossplane-runtime v2.5.0 and ModernManaged)
- All controllers operational with comprehensive test coverage
- **BREAKING CHANGE**: v0.11.0 removed all v1alpha1 cluster-scoped APIs
- **Test Coverage**: 41.7% overall (177 test functions across 31 test files)
  - HTTP Client: 56.8% coverage (core networking and API communication)
  - Controllers: 47.7-68.6% coverage (domain 68.6%, smtpcredential 62.9%, template 57.0%, bounce 54.1%, route 51.1%, mailinglist 47.8%, webhook 47.7%, complaint 59.3%, unsubscribe 59.3%, pcusage 12.5%)
  - Utility modules: 92.7-100% coverage (metrics 100%, errors 98.6%, health 97.1%, tracing 73.0%, features 45.8%)

## Build and Deployment Process

### ⚠️ Critical Build Requirements
- **Go Version**: Go 1.27.1+ required (specified in go.mod)
- **Dockerfile**: Builds the provider binary via `make build` and copies `bin/${TARGETOS}_${TARGETARCH}/provider` into `gcr.io/distroless/static:nonroot`
- **golangci-lint**: Use v2.13.2 for Go 1.27.1 compatibility

### Standard Build Commands
```bash
# Build provider binary directly
 go build -o provider cmd/provider/main.go

# Run comprehensive test suite
make test

# Generate code (DeepCopy, managed resources, CRDs)
make generate

# Build the release package
make build.all build.artifacts VERSION=v0.23.2 PLATFORMS="linux_amd64 linux_arm64"
for platform in linux_amd64 linux_arm64; do
  make xpkg.build VERSION=v0.23.2 PLATFORMS="linux_amd64 linux_arm64" PLATFORM="$platform"
done
```

### Release Process

Publication is performed by the tag-only GitHub Actions workflow after the release PR is merged and `master` is green. Do not push a plain image to the xpkg repository.

## Recent Improvements (2026-09-19)

### isDomainUpToDate Drift Detection, Example Fixes, Coverage (v0.23.1)
- **isDomainUpToDate real comparison**: previously always returned `true` (no drift detection). Mailgun's v4 GET `/domains/{name}` returns mutable fields `web_scheme`, `wildcard`, `spam_action`, `tracking`. These are now mapped via `responseToObservation` and compared in `isDomainUpToDate`, so PUT is issued when user changes any of them.
- **Example fixes**: `examples/domain/domain.yaml` — removed duplicate `namespace:` keys; `examples/sample-resources.yaml` — removed duplicate `namespace:` in Secret, converted snake_case to camelCase for all forProvider fields, removed invalid fields `pool_id` and `web_prefix`.
- **Coverage improvements**: Added `internal/tracing/tracing_test.go` (73.0%), `internal/controller/pcusage/tracker_test.go`, `TestIsDomainUpToDate` with 14 cases, and features tests (ValidateIPAllowlist, IsIPAllowed, LoginValidator, GetCredentialMetrics). Overall: 38.9% → 41.7%.
- **Pre-existing bugs fixed**: complaint and unsubscribe mock clients returned plain `errors.New` instead of `*clients.APIError{StatusCode: http.StatusNotFound}` for 404 responses, breaking `IsNotFound()` checks in Observe and Delete paths.
- **Nil tracer guard**: `StartOperation` now checks `tracer == nil` before calling `tracer.Start` to prevent panic when tracing is disabled.

### Type-Aware DNS Verification and Throttled Re-Verification (v0.22.0)
- **Bug A - type-aware DNSVerified**: `responseToObservation` previously computed `DNSVerified` across the union of receiving + sending records. A `sending` domain has no MX records (they are not required for sending), but Mailgun's v4 API still lists them with `valid="unknown"`, so every sending domain was permanently reported unverified. `DNSVerified` and `status.atProvider.requiredDnsRecords` are now computed from only the record set required for the domain's type (`sendingDnsRecords` by default, `receivingDnsRecords` for `type=receiving`), using the type Mailgun echoes in the v4 response.
- **Bug B - throttled re-verification**: the controller called `PUT /v4/domains/{name}/verify` on every reconcile while DNS was unverified. Each call triggers Mailgun's asynchronous DNS re-check and can send the account owner a "domain is now verified" notification email, causing API and email churn every 5-10 minutes. Explicit re-verification is now throttled to once per `dnsReverifyCooldown` (30 minutes) per Domain, tracked by the new `mailgun.crossplane.io/last-reverify` annotation (`v1beta1.AnnotationDNSLastReverify`).
- **Tests**: added client fixtures for sending-type (unknown MX + valid TXT/CNAME => verified) and receiving-type (unknown MX => unverified) semantics, a verify-path regression fixture, and controller tests for first revertify / within-cooldown skip / post-cooldown revertify plus the RFC3339 timestamp write.

### ManagementPolicies, State Metrics, Required Region (v0.19.0 - v0.21.2)
- **ManagementPolicies support (Observe)**: controllers honour Crossplane `managementPolicies`, including Observe-only enforcement.
- **State metrics**: `MRStateMetrics` reports Ready/Synced counts per managed resource kind.
- **Required region**: the `ProviderConfig` `region` field is required (US or EU) with no implicit default; explicit `apiBaseURL` still derives the v4 base URL by replacing the `/v3` suffix.
- **Toolchain**: Go 1.27.1, golangci-lint 2.13.2, crossplane-runtime v2.5.0, GitHub Actions brought up to current major versions.

## Recent Improvements (2026-07-30)

### DNS Automation Paths - ConfigMap Output, external-dns Annotation, DNS Probe (v0.18.0)
- **MX records name fix**: Mailgun's v4 Domains API omits the `name` field on MX records (they belong to the parent domain). `convertDNSRecords` now substitutes the parent domain name for MX records with empty `name`, so `kubectl describe domain` and the new ConfigMap output show complete records instead of `(MX)`.
- **DNSProvider spec (cloudflare stub)**: New `spec.forProvider.dnsProvider` block on `Domain` — currently a placeholder for a future Cloudflare integration (kept additive so existing Domains are unaffected). The wire shape is stable: `dnsProvider.cloudflare.{apiTokenSecretRef, zoneID, keepOnDelete}`.
- **DNS-records ConfigMap output (opt-in)**: Annotation `mailgun.crossplane.io/dns-configmap: "true"` causes the controller to create/update a ConfigMap named `<domain>-dns-records` in the Domain's namespace. Three keys: `records.yaml` (generic YAML), `terraform.tf` (hashicorp/dns HCL), `bind-zone.txt` (BIND zone fragment). Owner-referenced to the Domain for automatic cleanup. RBAC shipped as `examples/provider/configmap-rbac.yaml`.
- **external-dns annotation (default-on)**: Controller writes `external-dns.alpha.kubernetes.io/hostname` onto every Domain that does not have `mailgun.crossplane.io/disable-external-dns: "true"`. Drop-in compatible with the standard `external-dns` Deployment + Mailgun webhook provider.
- **DNSRecordsRequired Normal event**: Emitted on every reconcile while DNS is unverified. Message lists every record as `<type>-<hash>=<expected value>` so users can copy/paste the entire list into their DNS provider without leaving `kubectl describe`. Per-event annotations are baked into the message because Crossplane runtime v2's Recorder does not persist the `Event.Annotations` field to k8s events.
- **DNS propagation probe (opt-in)**: Annotation `mailgun.crossplane.io/dns-probe: "true"` enables a DNS prober (miekg/dns v1.1.72) that queries 8.8.8.8, 1.1.1.1, and 9.9.9.9 and emits `DNSNotPropagated` / `DNSValueMismatch` / `DNSRecordMatches` / `SPFNeedsMerge` / `DNSProbeError`. SPF-merge detection flags TXT records at the apex whose value starts with `v=spf1` and contains `include:mailgun.org` but is not the expected Mailgun directive.
- **Tests**: Added `TestDNSRecordsConfigMapName`, `TestBuildDNSRecordsConfigMap_HasAllThreeFormats`, `TestRenderRecordsYAML_IncludesAllRecords`, `TestRenderTerraform_RoundTripsAllRecordTypes` (with double-dot-regression guard), `TestRenderBindZone_HandlesMXAndTXT`, `TestDNSRecordAnnotationKey_*`, `TestFormatAnnotationBlock_*`, `TestProbeOne_*` (8 cases), `TestMiekDNSProber_*`, `TestRenderAnswers_*`, `TestMatchesExpected_*`, `TestSetExternalDNSHostname_*`, `TestEnsureDNSRecordsConfigMap_*`, `TestPublishDNSRecordsRequired_*`. 99+ test functions, all passing, lint clean.

## Recent Improvements (2026-07-30)

### DNS Re-Verification Loop, Rate-Limit Handling, RecordValidity Enum, Stable ServiceAccount (v0.17.3)
- **DNS Re-Verification Loop**: When DNS records are not yet verified, the Domain controller now calls `VerifyDomain` to trigger a Mailgun DNS re-check and sets `crossplane.io/poll-interval=5m` so the next reconcile happens on a Mailgun-friendly cadence rather than the controller-runtime default. This eliminates API-quota-burning tight loops on `dnsVerified=False`. A `DNSReverifyRequested` event is emitted so the user has a visible signal.
- **Rate-Limit Handling**: The HTTP client now detects HTTP 429 responses, parses `Retry-After` (seconds or HTTP-date) and `X-RateLimit-Reset` (epoch seconds), and honours the server-supplied back-off between retries. A typed `*RateLimitError` is returned after the configured retry budget is exhausted.
- **Typed `RecordValidity` Enum**: `DNSRecord.Valid` is now `*RecordValidity` (a string-typed enum) instead of `*string`. `// +kubebuilder:validation:Enum=valid;unknown` is applied to the CRD field so typos are rejected at admission time. The `IsVerified()` helper encapsulates the nil / unknown check.
- **State Printcolumn**: `kubectl get domain` now displays a `STATE` column (`active`, `unverified`, etc.) sourced from `.status.atProvider.state`.
- **Stable ServiceAccount**: `examples/provider/stable-sa-deployment-runtime-config.yaml` shows how to pin ProviderRevisions to a stable ServiceAccount (`provider-mailgun`) via `DeploymentRuntimeConfig` and bind RBAC to it precisely. `examples/provider/events-rbac.yaml` ships the matching `ClusterRole` granting `events.events.k8s.io/v1` to the provider.
- **Tests**: Added `TestParseRateLimit`, `TestRateLimitError_Error`, `TestMakeRequestRateLimitRetry`, `TestMakeRequestRateLimitExhaustion` for 429 handling, plus `TestDomainObserveDNSReverify` covering the new reverify / poll-interval logic. All 133+ tests pass, lint clean.

### Mailgun v4 Domains API Migration (v0.17.2)
- **Bug Fix**: `status.atProvider.receivingDnsRecords`, `sendingDnsRecords`, `requiredDnsRecords` and `dnsVerified` were empty for every Domain reconcile under v0.17.0 / v0.17.1. Root cause: the v3 endpoints used previously do not return DNS records in their responses.
- **Fix**: Migrated `CreateDomain`, `GetDomain`, `UpdateDomain` and `VerifyDomain` to the Mailgun v4 Domains API (`/v4/domains`, `/v4/domains/{name}`, `/v4/domains/{name}/verify`). The v4 responses include `receiving_dns_records` and `sending_dns_records` at the top level alongside `domain`, plus a per-record `valid` field (`"valid"` or `"unknown"`).
- **API Type Changes**: `DNSRecord.Valid` changed from `*bool` to `*string` to match Mailgun's v4 enum exactly; `DNSRecord.Priority` changed from `*int` to `*string` to match the wire format (e.g. `"10"`).
- **URL Handling**: Added `Config.V4BaseURL` field and `makeRequestAt(baseURL, ...)` helper. `UseProviderConfig` derives `V4BaseURL` from a user-supplied `/v3` `apiBaseURL` by swapping the suffix, so existing ProviderConfigs continue to work without modification.
- **Controller Simplification**: Removed the Observe fallback that called `VerifyDomain` when records were empty — `GetDomain` now returns the records on every reconcile.
- **Verify Endpoint**: `VerifyDomain` no longer follows up with a second `GetDomain`; the verify response already carries the current DNS record set with validity.
- **Tests**: Rewrote wire-shape fixtures for v4; added `TestRecordIsValid` covering nil / "valid" / "unknown" / empty / arbitrary inputs. All 133+ tests pass, lint clean.

## Previous Improvements (2025-10-01)

### Go 1.26.3 and golangci-lint 2.5.0 Upgrade (v0.14.3)
- **Go Version Upgrade**: Updated from Go 1.24.5 to Go 1.26.3 throughout entire codebase
- **golangci-lint Upgrade**: Upgraded to golangci-lint 2.5.0 for modern Go support and compatibility
- **Code Quality Cleanup**: Removed 8 unused functions causing lint warnings across controllers
- **Lint Compliance**: Achieved 0 lint issues with make lint passing cleanly
- **Documentation Updates**: Updated all version references in README, CLAUDE.md, GitHub workflows
- **Build System**: Updated Makefile and build/makelib/golang.mk for consistent tooling
- **Test Stability**: All 133+ tests continue passing after cleanup with no functional changes
- **Crossplane v2 Native**: Confirmed clean v2 provider with no backward compatibility baggage

### Crossplane Runtime Update (v0.12.0) (2025-09-15)

### Crossplane Runtime Update (v0.12.0)
- **Updated crossplane-runtime**: Upgraded from v1.20.0 to v1.21.0-rc.0 to address ProviderConfigUsage namespace creation issues
- **Fixed Package Building**: Resolved "not exactly one package meta type" error by removing duplicate package.yaml file
- **Generated Code Completion**: Added missing v1beta1 managed resource code for all resource types
- **Deployment Status**: Successfully deployed to production with updated runtime
- **Build System**: Fixed Crossplane package (.xpkg) building process with embedded Docker images
- **All Tests Passing**: 133+ test functions continue to pass with 36.3% overall coverage

### HTTP Service Consolidation (v0.10.1) (2025-08-14)

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

The provider supports multiple Mailgun regions via a registry table in
`internal/clients/mailgun.go` (`Region` struct + `regions` slice). Adding
a new region is one entry plus a CRD enum update — no conditional code
edits required.

Currently registered:
- **US** (default): API `https://api.mailgun.net/v3`, v4 Domains `https://api.mailgun.net/v4`, SMTP `smtp.mailgun.org`
- **EU**: API `https://api.eu.mailgun.net/v3`, v4 Domains `https://api.eu.mailgun.net/v4`, SMTP `smtp.eu.mailgun.org`

Most resources (mailing lists, routes, webhooks, SMTP credentials, templates, bounces, complaints, unsubscribes) hit v3 endpoints. Domain management (create, get, update, verify) uses the Mailgun v4 Domains API so that `receiving_dns_records` and `sending_dns_records` (plus their per-record `valid` status) are returned to the user.

Configure via `region` field in ProviderConfig (required: US or EU) or explicit `apiBaseURL`. Region resolution priority:

1. Explicit `spec.region` if it matches a registered code.
2. URL marker on `apiBaseURL` (longest match wins, so `eu.mailgun.net` resolves to EU rather than the more general US `mailgun.net` marker).
3. First registered region (US) - fallback for backwards compatibility.

When `apiBaseURL` is supplied with a `/v3` suffix, the v4 base URL is derived by replacing `/v3` with `/v4`. The SMTP host is always taken from the resolved region so downstream consumers (Keycloak, Odoo, …) authenticate against the relay that matches where the credential was issued.
