# Crossplane Provider Mailgun Helm Chart

This Helm chart installs the Crossplane Provider for Mailgun in a Kubernetes cluster.

## Prerequisites

- Kubernetes 1.20+
- Helm 3.0+
- Crossplane installed in the cluster

## Installation

### Add the repository (when published)

```bash
helm repo add provider-mailgun https://crossplane-contrib.github.io/provider-mailgun
helm repo update
```

### Install from local chart

```bash
# Clone the repository
git clone https://github.com/rossigee/provider-mailgun.git
cd provider-mailgun

# Install the chart
helm install provider-mailgun ./charts/provider-mailgun \
  --namespace crossplane-system \
  --create-namespace
```

### Install with custom values

```bash
helm install provider-mailgun ./charts/provider-mailgun \
  --namespace crossplane-system \
  --create-namespace \
  --values my-values.yaml
```

## Configuration

The following table lists the configurable parameters and their default values:

| Parameter | Description | Default |
|-----------|-------------|---------|
| `image.repository` | Provider image repository | `crossplane/provider-mailgun` |
| `image.tag` | Provider image tag | `v0.1.0` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `resources.limits.cpu` | CPU limit | `500m` |
| `resources.limits.memory` | Memory limit | `512Mi` |
| `resources.requests.cpu` | CPU request | `100m` |
| `resources.requests.memory` | Memory request | `128Mi` |
| `healthChecks.enabled` | Enable health checks | `true` |
| `metrics.enabled` | Enable metrics | `true` |
| `metrics.serviceMonitor.enabled` | Enable ServiceMonitor for Prometheus | `false` |
| `rbac.create` | Create RBAC resources | `true` |
| `serviceAccount.create` | Create service account | `true` |
| `podDisruptionBudget.enabled` | Enable PodDisruptionBudget | `false` |
| `networkPolicy.enabled` | Enable NetworkPolicy | `false` |

## Usage

After installation, you can create a ProviderConfig to configure the Mailgun API access:

```yaml
apiVersion: pkg.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: mailgun-config
spec:
  credentials:
    source: Secret
    secretRef:
      namespace: crossplane-system
      name: mailgun-credentials
      key: apikey
```

Create the secret with your Mailgun API key:

```bash
kubectl create secret generic mailgun-credentials \
  --from-literal=apikey=your-mailgun-api-key \
  --namespace crossplane-system
```

Then you can create Mailgun resources:

```yaml
apiVersion: domain.mailgun.m.crossplane.io/v1beta1
kind: Domain
metadata:
  name: example-domain
  namespace: default
spec:
  forProvider:
    name: example.com
    smtpPassword: auto-generate
  providerConfigRef:
    name: mailgun-config
```

## Monitoring

The chart supports Prometheus monitoring through:

1. **Metrics endpoint**: Available at `:8080/metrics`
2. **ServiceMonitor**: Enable with `metrics.serviceMonitor.enabled=true`
3. **Health checks**: Liveness and readiness probes at `/healthz` and `/readyz`

## Security

The chart includes several security features:

- **Security Context**: Runs as non-root user with read-only filesystem
- **RBAC**: Minimal required permissions
- **Network Policy**: Optional network isolation
- **Pod Security Standards**: Compatible with restricted pod security standards

## Troubleshooting

### Check provider status

```bash
kubectl get providers
kubectl describe provider provider-mailgun
```

### Check controller logs

```bash
kubectl logs -n crossplane-system -l app.kubernetes.io/name=provider-mailgun
```

### Verify ProviderConfig

```bash
kubectl get providerconfigs
kubectl describe providerconfig mailgun-config
```

## Contributing

Contributions are welcome! Please read the [contributing guide](../../CONTRIBUTING.md) for details.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](../../LICENSE) file for details.
