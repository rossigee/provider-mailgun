# Provider-Mailgun Production Deployment Guide

## 🎯 Current Status

✅ **Completed:**
- Provider built and tested locally (all tests passing)
- Comprehensive enterprise features implemented
- Deployment manifests created and tested
- Security configuration validated
- Monitoring and observability configured

⏳ **Pending:**
- Git push requires hardware key authentication
- New container image build via GitHub Actions
- Full provider deployment and testing

## 🚀 Deployment Steps

### Step 1: Complete Code Push
```bash
# Authenticate with hardware key and push
git push origin master
```

This will trigger GitHub Actions to:
- Build with Go 1.25.1
- Run comprehensive test suite (34.9% coverage)
- Create new container image with correct package format
- Push to ghcr.io/rossigee/provider-mailgun:latest

### Step 2: Deploy Provider
```bash
# Deploy provider with production configuration
kubectl apply -f examples/provider-config.yaml

# Wait for provider to become healthy
kubectl wait --for=condition=Healthy provider.pkg.crossplane.io/provider-mailgun --timeout=300s

# Verify installation
kubectl get providers provider-mailgun
```

### Step 3: Configure Credentials
```bash
# Deploy secret template
kubectl apply -f examples/provider-config-secret.yaml

# Update with your actual Mailgun API key
kubectl edit secret mailgun-creds -n crossplane-system
# Replace YOUR_MAILGUN_API_KEY_HERE with your actual key
```

### Step 4: Validate CRDs
```bash
# Check all CRDs are installed
kubectl get crd | grep mailgun

# Expected CRDs (v2 namespaced):
# - domains.domain.mailgun.m.crossplane.io
# - mailinglists.mailinglist.mailgun.m.crossplane.io
# - routes.route.mailgun.m.crossplane.io
# - webhooks.webhook.mailgun.m.crossplane.io
# - smtpcredentials.smtpcredential.mailgun.m.crossplane.io
# - templates.template.mailgun.m.crossplane.io
# - bounces.bounce.mailgun.m.crossplane.io
# - providerconfigs.mailgun.m.crossplane.io
# - providerconfigusages.mailgun.m.crossplane.io
```

### Step 5: Test with Sample Resources
```bash
# Deploy sample resources (uses placeholder domains)
kubectl apply -f examples/sample-resources.yaml

# Monitor resource creation
kubectl get domains,mailinglists,smtpcredentials,webhooks,routes -w
```

## 📊 Monitoring & Observability

### Provider Logs
```bash
# Follow provider logs
kubectl logs -n crossplane-system -l app.kubernetes.io/name=provider-mailgun -f

# Check for specific events
kubectl get events -n crossplane-system --field-selector involvedObject.name=provider-mailgun
```

### Metrics (if Prometheus operator available)
```bash
# Deploy monitoring configuration
kubectl apply -f examples/monitoring.yaml

# Port forward to metrics endpoint
kubectl port-forward -n crossplane-system svc/provider-mailgun-metrics 8080:8080

# Access metrics
curl http://localhost:8080/metrics

# Key metrics to monitor:
# - mailgun_operation_total (operation counts by status)
# - mailgun_operation_duration_seconds (operation latency)
# - mailgun_circuit_breaker_state (resilience status)
# - mailgun_retry_attempts_total (retry statistics)
```

### Health Checks
```bash
# Check provider health endpoint
kubectl port-forward -n crossplane-system deployment/provider-mailgun-controller 8080:8080
curl http://localhost:8080/healthz

# Expected response: {"status":"ok","timestamp":"...","components":{...}}
```

## 🧪 Testing Scenarios

### Basic Functionality Test
```bash
# Create a domain (will fail with placeholder but tests controller)
kubectl apply -f - <<EOF
apiVersion: domain.mailgun.m.crossplane.io/v1beta1
kind: Domain
metadata:
  name: test-domain
  namespace: default
spec:
  forProvider:
    name: test.example.com
    smtp_password: "test-password-123"
  providerConfigRef:
    name: default
EOF

# Check domain status
kubectl describe domain test-domain
```

### SMTP Credentials Test
```bash
# Create SMTP credential
kubectl apply -f - <<EOF
apiVersion: smtpcredential.mailgun.m.crossplane.io/v1beta1
kind: SMTPCredential
metadata:
  name: test-smtp
  namespace: default
spec:
  forProvider:
    login: "test@example.com"
    passwordSecretRef:
      namespace: default
      name: test-smtp-password
      key: password
  providerConfigRef:
    name: default
EOF

# Check credential status
kubectl describe smtpcredential test-smtp
```

### Real Mailgun Integration Test
If you have a real Mailgun account:

```bash
# Update examples with your real domain
sed -i 's/example\.com/yourdomain.com/g' examples/sample-resources.yaml

# Apply with real resources
kubectl apply -f examples/sample-resources.yaml

# Monitor creation
kubectl get domains,mailinglists -o wide
```

## 🛠️ Troubleshooting

### Provider Not Healthy
```bash
# Check provider status
kubectl get provider provider-mailgun -o yaml

# Common issues:
# 1. Package format incompatibility (need new image from GitHub Actions)
# 2. Invalid Mailgun credentials
# 3. Network connectivity issues
# 4. Resource constraints
```

### Resource Creation Failures
```bash
# Check resource conditions
kubectl describe domain <domain-name>

# Check provider logs for errors
kubectl logs -n crossplane-system -l app.kubernetes.io/name=provider-mailgun --tail=50

# Common failures:
# 1. Invalid API key or region mismatch
# 2. Domain already exists in Mailgun
# 3. Rate limiting or quota exceeded
# 4. Network timeouts
```

### Performance Issues
```bash
# Check resource usage
kubectl top pods -n crossplane-system -l app.kubernetes.io/name=provider-mailgun

# Check metrics for high latency
curl -s http://localhost:8080/metrics | grep mailgun_operation_duration

# Check circuit breaker state
curl -s http://localhost:8080/metrics | grep mailgun_circuit_breaker_state
```

## 🔧 Advanced Configuration

### Custom Resource Limits
Edit `examples/provider-config.yaml` to adjust:
- CPU/Memory limits
- Tracing sampling ratio
- Logging verbosity
- Retry configuration

### Multi-Region Setup
For EU region support:
```yaml
apiVersion: mailgun.m.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: eu-region
  namespace: crossplane-system
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: mailgun-eu-creds
      key: credentials
  region: eu  # Set to EU region
```

### Development Mode
For development with verbose logging:
```yaml
# In DeploymentRuntimeConfig
args:
- --debug
- --leader-election=false  # Disable for single instance
- --max-reconcile-rate=10  # Reduce load
```

## 📋 Production Checklist

### Pre-Deployment
- [ ] Hardware key authentication completed
- [ ] GitHub Actions build successful
- [ ] Kubernetes cluster accessible
- [ ] Crossplane installed and healthy
- [ ] Mailgun API credentials ready

### During Deployment
- [ ] Provider installed successfully
- [ ] All CRDs created
- [ ] Provider pod running and healthy
- [ ] Credentials configured correctly
- [ ] Basic connectivity test passed

### Post-Deployment
- [ ] Sample resources created successfully
- [ ] Monitoring configured (if available)
- [ ] Logs showing normal operation
- [ ] Performance metrics baseline established
- [ ] Documentation updated with actual domain names

## 🎉 Success Criteria

**Provider Status:**
- `kubectl get provider provider-mailgun` shows `INSTALLED=True, HEALTHY=True`
- All 9 CRDs created successfully
- Provider pod logs show no errors

**Functionality:**
- Domain creation/update/deletion working
- SMTP credentials management working
- Webhooks and routes configurable
- Error handling graceful with proper conditions

**Observability:**
- Metrics endpoint responding
- Structured logs with operation context
- Health checks passing
- Circuit breaker functioning properly

**Security:**
- No secrets in logs or metrics
- RBAC properly configured
- Security contexts enforced
- Network policies applied (if available)
