# Provider-Mailgun Quick Start Guide

## 🚀 5-Minute Setup

### Prerequisites
- Kubernetes cluster with Crossplane installed
- Mailgun API key (from mailgun.com)
- kubectl configured

### Step 1: Deploy Provider
```bash
kubectl apply -f examples/provider-config.yaml
kubectl wait --for=condition=Healthy provider.pkg.crossplane.io/provider-mailgun --timeout=300s
```

### Step 2: Configure Credentials
```bash
kubectl apply -f examples/provider-config-secret.yaml
kubectl patch secret mailgun-creds -n crossplane-system -p '{"stringData":{"credentials":"{\"api_key\":\"YOUR_API_KEY_HERE\",\"region\":\"us\"}"}}'
```

### Step 3: Test with Domain
```bash
kubectl apply -f - <<EOF
apiVersion: domain.mailgun.crossplane.io/v1alpha1
kind: Domain
metadata:
  name: my-domain
spec:
  forProvider:
    name: mail.yourdomain.com
    smtp_password: "secure-password-123"
  providerConfigRef:
    name: default
EOF
```

### Step 4: Verify
```bash
kubectl get domains
kubectl describe domain my-domain
```

## 📋 Common Resources

### SMTP Credentials
```yaml
apiVersion: smtpcredential.mailgun.crossplane.io/v1alpha1
kind: SMTPCredential
metadata:
  name: app-smtp
spec:
  forProvider:
    login: "app@yourdomain.com"
    passwordSecretRef:
      namespace: default
      name: smtp-secret
      key: password
  providerConfigRef:
    name: default
```

### Webhook
```yaml
apiVersion: webhook.mailgun.crossplane.io/v1alpha1
kind: Webhook
metadata:
  name: delivery-webhook
spec:
  forProvider:
    kind: delivered
    urls:
    - "https://yourdomain.com/webhooks/delivered"
  providerConfigRef:
    name: default
```

### Mailing List
```yaml
apiVersion: mailinglist.mailgun.crossplane.io/v1alpha1
kind: MailingList
metadata:
  name: newsletter
spec:
  forProvider:
    address: newsletter@yourdomain.com
    name: "Company Newsletter"
    access_level: everyone
  providerConfigRef:
    name: default
```

## 🔍 Troubleshooting

### Provider Not Healthy
```bash
kubectl describe provider provider-mailgun
kubectl logs -n crossplane-system -l app.kubernetes.io/name=provider-mailgun
```

### Resource Stuck
```bash
kubectl describe domain <name>  # Check conditions
kubectl get events --field-selector involvedObject.name=<name>
```

### API Key Issues
```bash
kubectl get secret mailgun-creds -n crossplane-system -o yaml
# Verify credentials format: {"api_key":"key-...","region":"us"}
```

## 📊 Monitoring

### Basic Health Check
```bash
kubectl get providers,domains,mailinglists,smtpcredentials
```

### Detailed Status
```bash
kubectl get providers provider-mailgun -o yaml
kubectl logs -n crossplane-system -l app.kubernetes.io/name=provider-mailgun -f
```

### Metrics (if available)
```bash
kubectl port-forward -n crossplane-system svc/provider-mailgun-metrics 8080:8080
curl http://localhost:8080/metrics | grep mailgun
```

## 🔗 Resources

- [Full Deployment Guide](DEPLOYMENT_GUIDE.md)
- [API Reference](https://pkg.go.dev/github.com/rossigee/provider-mailgun)
- [Mailgun API Docs](https://documentation.mailgun.com/en/latest/api_reference.html)
- [Crossplane Docs](https://crossplane.io/docs/)
