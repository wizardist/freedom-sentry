#!/usr/bin/env bash
set -euo pipefail

CHART_NAME="freedom-sentry-helm"
CHART_VERSION="${1}"

if [ -z "$CHART_VERSION" ]; then
    echo "Usage: $0 <chart-version>"
    echo "Example: $0 0.1.0"
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

CHART_DIR="./chart"
REGISTRY="oci://${REGISTRY_URL}/${REGISTRY_PROJECT}"

mkdir -p ./dist

echo "=== Updating Chart.yaml version to ${CHART_VERSION} ==="
sed -i "s/^version: .*/version: ${CHART_VERSION}/" "${CHART_DIR}/Chart.yaml"
sed -i "s/^appVersion: .*/appVersion: \"${CHART_VERSION}\"/" "${CHART_DIR}/Chart.yaml"

echo "=== Packaging Helm chart ==="
helm package "${CHART_DIR}" -d ./dist/

CHART_PACKAGE="./dist/${CHART_NAME}-${CHART_VERSION}.tgz"

echo "=== Pushing chart to ${REGISTRY} ==="
helm push "$CHART_PACKAGE" "$REGISTRY"

echo "=== Release ${CHART_VERSION} complete! ==="
echo "Chart: ${REGISTRY}/${CHART_NAME}:${CHART_VERSION}"
