#!/bin/bash

# Build and push provider-mailgun to multiple registries
set -e

# Default values - Standardized registry configuration
VERSION=${VERSION:-dev}
BUILD_PACKAGE=${BUILD_PACKAGE:-false}
PLATFORMS=${PLATFORMS:-linux/amd64,linux/arm64}

# Primary registry: GitHub Container Registry
PRIMARY_REGISTRY=${PRIMARY_REGISTRY:-ghcr.io/rossigee}

# Optional registries
ENABLE_HARBOR_PUBLISH=${ENABLE_HARBOR_PUBLISH:-false}
ENABLE_DOCKERHUB_PUBLISH=${ENABLE_DOCKERHUB_PUBLISH:-false}
HARBOR_REGISTRY=${HARBOR_REGISTRY:-} # Harbor removed - no default registry
DOCKERHUB_REGISTRY=${DOCKERHUB_REGISTRY:-docker.io/rossigee}

# Provider name
PROVIDER_NAME=provider-mailgun

echo "Building ${PROVIDER_NAME} version ${VERSION}"
echo "Platforms: ${PLATFORMS}"
echo "Primary Registry: ${PRIMARY_REGISTRY}"
echo "Harbor Registry: ${ENABLE_HARBOR_PUBLISH}"
echo "Docker Hub Registry: ${ENABLE_DOCKERHUB_PUBLISH}"
echo "Build Crossplane package: ${BUILD_PACKAGE}"

# Build and push to primary registry (GitHub Container Registry)
echo "Building and pushing to primary registry (${PRIMARY_REGISTRY})..."
docker buildx build \
  --platform "${PLATFORMS}" \
  -t "${PRIMARY_REGISTRY}/${PROVIDER_NAME}:${VERSION}" \
  -t "${PRIMARY_REGISTRY}/${PROVIDER_NAME}:latest" \
  -f cluster/images/provider-mailgun/Dockerfile \
  --push \
  .

# Push to Harbor registry if enabled
if [ "${ENABLE_HARBOR_PUBLISH}" = "true" ]; then
  echo "Pushing to Harbor registry (${HARBOR_REGISTRY})..."
  docker buildx build \
    --platform "${PLATFORMS}" \
    -t "${HARBOR_REGISTRY}/${PROVIDER_NAME}:${VERSION}" \
    -t "${HARBOR_REGISTRY}/${PROVIDER_NAME}:latest" \
    -f cluster/images/provider-mailgun/Dockerfile \
    --push \
    .
fi

# Push to Docker Hub if enabled
if [ "${ENABLE_DOCKERHUB_PUBLISH}" = "true" ]; then
  echo "Pushing to Docker Hub (${DOCKERHUB_REGISTRY})..."
  docker buildx build \
    --platform "${PLATFORMS}" \
    -t "${DOCKERHUB_REGISTRY}/${PROVIDER_NAME}:${VERSION}" \
    -t "${DOCKERHUB_REGISTRY}/${PROVIDER_NAME}:latest" \
    -f cluster/images/provider-mailgun/Dockerfile \
    --push \
    .
fi

# Build Crossplane package if requested
if [ "${BUILD_PACKAGE}" = "true" ]; then
  echo "Building Crossplane package..."
  make xpkg.build
fi

echo "Build complete!"
