#!/bin/bash
set -e

echo "🚀 Testing provider-mailgun deployment and functionality"
echo "======================================================="

# Check if we're in a Kubernetes cluster
if ! kubectl cluster-info &>/dev/null; then
    echo "❌ No Kubernetes cluster available. Please ensure kubectl is configured."
    exit 1
fi

# Check if Crossplane is installed
if ! kubectl get deployment crossplane -n crossplane-system &>/dev/null; then
    echo "❌ Crossplane not found. Please install Crossplane first:"
    echo "   helm repo add crossplane-stable https://charts.crossplane.io/stable"
    echo "   helm repo update"
    echo "   helm install crossplane crossplane-stable/crossplane --namespace crossplane-system --create-namespace"
    exit 1
fi

echo "✅ Kubernetes cluster and Crossplane found"

# Install the provider
echo "📦 Installing provider-mailgun..."
kubectl apply -f examples/provider-config.yaml

# Wait for provider to be installed
echo "⏳ Waiting for provider to be installed..."
kubectl wait --for=condition=Installed provider.pkg.crossplane.io/provider-mailgun --timeout=300s

# Check provider status
echo "📊 Provider status:"
kubectl get provider.pkg.crossplane.io/provider-mailgun -o yaml

# Install monitoring (if Prometheus operator is available)
if kubectl get crd servicemonitors.monitoring.coreos.com &>/dev/null; then
    echo "📈 Installing monitoring..."
    kubectl apply -f examples/monitoring.yaml
    echo "✅ Monitoring configured"
else
    echo "⚠️  Prometheus operator not found, skipping monitoring setup"
fi

# Validate CRDs are installed
echo "🔍 Checking CRDs..."
CRDS=(
    "domains.domain.mailgun.crossplane.io"
    "mailinglists.mailinglist.mailgun.crossplane.io"
    "routes.route.mailgun.crossplane.io"
    "webhooks.webhook.mailgun.crossplane.io"
    "smtpcredentials.smtpcredential.mailgun.crossplane.io"
    "templates.template.mailgun.crossplane.io"
    "bounces.bounce.mailgun.crossplane.io"
    "providerconfigs.mailgun.crossplane.io"
    "providerconfigusages.mailgun.crossplane.io"
)

for crd in "${CRDS[@]}"; do
    if kubectl get crd "$crd" &>/dev/null; then
        echo "✅ $crd"
    else
        echo "❌ $crd"
    fi
done

# Check provider pod logs
echo "📋 Provider pod logs (last 20 lines):"
POD_NAME=$(kubectl get pods -n crossplane-system -l app=provider-mailgun -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")
if [ -n "$POD_NAME" ]; then
    kubectl logs -n crossplane-system "$POD_NAME" --tail=20
else
    echo "⚠️  Provider pod not found yet"
fi

echo ""
echo "🎉 Provider installation test completed!"
echo ""
echo "📝 Next steps:"
echo "   1. Create Mailgun API key secret: kubectl apply -f examples/provider-config-secret.yaml"
echo "   2. Update the secret with your actual Mailgun API key"
echo "   3. Test with sample resources: kubectl apply -f examples/sample-resources.yaml"
echo ""
echo "📊 Monitor with:"
echo "   kubectl get providers"
echo "   kubectl get domains"
echo "   kubectl get mailinglists"
echo "   kubectl logs -n crossplane-system -l app=provider-mailgun -f"
