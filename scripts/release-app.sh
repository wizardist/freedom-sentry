#!/usr/bin/env bash
set -euo pipefail

VERSION="${1}"

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 1.2.3"
    exit 1
fi

# Load registry configuration from .env
if [ -f .env ]; then
    set -a
    source .env
    set +a
else
    echo "Error: .env file not found"
    echo "Copy .env.example to .env and configure registry URLs"
    exit 1
fi

# Verify required variables
if [ -z "${REGISTRY_URL:-}" ] || [ -z "${REGISTRY_PROJECT:-}" ]; then
    echo "Error: REGISTRY_URL and REGISTRY_PROJECT must be set in .env"
    exit 1
fi

IMAGE_NAME="freedom-sentry"
REGISTRY="${REGISTRY_URL}/${REGISTRY_PROJECT}"

# Ensure clean git state for production releases
if [[ "$VERSION" != *"-dev" ]] && [[ "$VERSION" != *"-"* ]]; then
    if [ -n "$(git status --porcelain)" ]; then
        echo "Error: Working directory not clean. Commit changes first."
        exit 1
    fi
fi

echo "=== Building ${REGISTRY}/${IMAGE_NAME}:${VERSION} ==="
docker build \
    --tag "${REGISTRY}/${IMAGE_NAME}:${VERSION}" \
    --tag "${REGISTRY}/${IMAGE_NAME}:latest" \
    .

echo "=== Pushing ${REGISTRY}/${IMAGE_NAME}:${VERSION} ==="
docker push "${REGISTRY}/${IMAGE_NAME}:${VERSION}"
docker push "${REGISTRY}/${IMAGE_NAME}:latest"

echo "=== Release ${VERSION} complete! ==="
echo "Image: ${REGISTRY}/${IMAGE_NAME}:${VERSION}"
