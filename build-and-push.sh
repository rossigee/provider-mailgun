#!/bin/bash

# Build and push provider-mailgun to multiple registries
set -e

# Default values
VERSION=${VERSION:-dev}
PUSH_EXTERNAL=${PUSH_EXTERNAL:-false}
BUILD_PACKAGE=${BUILD_PACKAGE:-false}
PLATFORMS=${PLATFORMS:-linux/amd64,linux/arm64}
REGISTRY=${REGISTRY:-harbor.golder.lan/library}

# Provider name
PROVIDER_NAME=provider-mailgun

echo "Building ${PROVIDER_NAME} version ${VERSION}"
echo "Platforms: ${PLATFORMS}"
echo "Registry: ${REGISTRY}"
echo "Push to Docker Hub: ${PUSH_EXTERNAL}"
echo "Build Crossplane package: ${BUILD_PACKAGE}"

# Build the Docker image
echo "Building Docker image..."
docker buildx build \
  --platform "${PLATFORMS}" \
  -t "${REGISTRY}/${PROVIDER_NAME}:${VERSION}" \
  -t "${REGISTRY}/${PROVIDER_NAME}:latest" \
  -f cluster/images/provider-mailgun/Dockerfile \
  --push \
  .

# Push to Docker Hub if requested
if [ "${PUSH_EXTERNAL}" = "true" ]; then
  echo "Pushing to Docker Hub..."
  docker buildx build \
    --platform "${PLATFORMS}" \
    -t "docker.io/rossigee/${PROVIDER_NAME}:${VERSION}" \
    -t "docker.io/rossigee/${PROVIDER_NAME}:latest" \
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
