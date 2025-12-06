#!/usr/bin/env bash
set -euo pipefail

# Unified release script for Freedom Sentry
# Automates version bumping, Docker image building, git tagging, and ops repo updates

# Configuration
APP_NAME="freedom-sentry"

# Load configuration from .env
if [ -f .env ]; then
    set -a
    source .env
    set +a
else
    echo "Error: .env file not found"
    echo "Copy .env.example to .env and configure required variables"
    exit 1
fi

# Verify required variables
if [ -z "${REGISTRY_URL:-}" ] || [ -z "${REGISTRY_PROJECT:-}" ]; then
    echo "Error: REGISTRY_URL and REGISTRY_PROJECT must be set in .env"
    exit 1
fi

if [ -z "${OPS_REPO_PATH:-}" ] || [ -z "${OPS_VALUES_FILE:-}" ]; then
    echo "Error: OPS_REPO_PATH and OPS_VALUES_FILE must be set in .env"
    exit 1
fi

REGISTRY="${REGISTRY_URL}/${REGISTRY_PROJECT}"

# Global flags
DRY_RUN=false
NO_PUSH=false

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print functions
info() {
    echo -e "${BLUE}[INFO]${NC} $*"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*" >&2
}

# Help text
show_help() {
    cat <<EOF
Usage: $0 [BUMP_TYPE|VERSION] [FLAGS]

Automated release workflow for Freedom Sentry.
Handles version bumping, Docker builds, git tagging, and ops repo updates.

Arguments:
  BUMP_TYPE       One of: patch, minor, major (default: patch)
  VERSION         Explicit version (e.g., 1.5.0)

Flags:
  --no-push       Create tag and commits locally without pushing to remotes
  --dry-run       Show what would be done without making any changes
  --help          Show this help message

Examples:
  $0                    # Patch bump: 1.3.2 → 1.3.3
  $0 patch              # Same as above
  $0 minor              # Minor bump: 1.3.2 → 1.4.0
  $0 major              # Major bump: 1.3.2 → 2.0.0
  $0 1.5.0              # Explicit version
  $0 patch --no-push    # Bump locally without pushing
  $0 --dry-run minor    # Preview changes without executing

EOF
}

# Get the latest version from git tags
get_latest_version() {
    local latest_tag
    latest_tag=$(git tag -l 'v*' | sort -V | tail -1)

    if [[ -z "$latest_tag" ]]; then
        echo "0.0.0"
    else
        echo "${latest_tag#v}"  # Remove 'v' prefix
    fi
}

# Validate semantic version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        error "Invalid version format: $version (expected: X.Y.Z)"
        return 1
    fi
    return 0
}

# Bump version based on bump type
bump_version() {
    local current_version=$1
    local bump_type=$2

    IFS='.' read -r major minor patch <<< "$current_version"

    case $bump_type in
        major)
            echo "$((major + 1)).0.0"
            ;;
        minor)
            echo "${major}.$((minor + 1)).0"
            ;;
        patch)
            echo "${major}.${minor}.$((patch + 1))"
            ;;
        *)
            error "Invalid bump type: $bump_type (expected: major, minor, or patch)"
            return 1
            ;;
    esac
}

# Check if git working directory is clean
check_git_clean() {
    local repo_path=${1:-.}
    local repo_name=${2:-repository}

    if [[ -n $(git -C "$repo_path" status --porcelain) ]]; then
        error "$repo_name has uncommitted changes. Please commit or stash them first."
        git -C "$repo_path" status --short
        return 1
    fi
    return 0
}

# Build and push Docker image
build_and_push_image() {
    local version=$1

    info "Building and pushing Docker image for version $version..."

    if [[ "$DRY_RUN" == "true" ]]; then
        info "[DRY RUN] Would execute: ./scripts/release-app.sh \"$version\""
        return 0
    fi

    ./scripts/release-app.sh "$version"
}

# Create git tag in app repo
create_git_tag() {
    local version=$1
    local tag="v$version"

    info "Creating git tag: $tag"

    if [[ "$DRY_RUN" == "true" ]]; then
        info "[DRY RUN] Would create tag: $tag"
        return 0
    fi

    git tag -a "$tag" -m "release: bump version to $version"

    if [[ "$NO_PUSH" == "false" ]]; then
        info "Pushing tag to origin..."
        git push origin "$tag"
        success "Tag $tag pushed successfully"
    else
        warn "Tag created locally (--no-push flag set)"
    fi
}

# Update ops repo values.yaml with new image tag
update_ops_repo() {
    local version=$1
    local ops_repo="$OPS_REPO_PATH"
    local values_file="$ops_repo/$OPS_VALUES_FILE"

    info "Updating ops repository..."

    # Verify ops repo exists
    if [[ ! -d "$ops_repo" ]]; then
        error "Ops repository not found at: $ops_repo"
        return 1
    fi

    if [[ ! -f "$values_file" ]]; then
        error "Values file not found at: $values_file"
        return 1
    fi

    if [[ "$DRY_RUN" == "true" ]]; then
        info "[DRY RUN] Would update $values_file"
        info "[DRY RUN] Would set image.tag to: \"$version\""
        return 0
    fi

    # Check if yq is available
    if command -v yq &> /dev/null; then
        info "Using yq to update YAML..."
        yq eval ".image.tag = \"$version\"" -i "$values_file"
    else
        info "Using sed to update YAML (yq not found)..."
        sed -i "s/^  tag: .*/  tag: \"$version\"/" "$values_file"
    fi

    # Verify the change
    if grep -q "tag: \"$version\"" "$values_file"; then
        success "Updated image tag to $version in $OPS_VALUES_FILE"
    else
        error "Failed to update image tag in $values_file"
        return 1
    fi

    # Commit changes
    info "Committing changes to ops repo..."
    git -C "$ops_repo" add "$OPS_VALUES_FILE"
    git -C "$ops_repo" commit -m "freedom-sentry: update app version to $version"

    if [[ "$NO_PUSH" == "false" ]]; then
        info "Pushing ops repo changes..."
        git -C "$ops_repo" push
        success "Ops repo updated and pushed successfully"
    else
        warn "Changes committed locally (--no-push flag set)"
    fi
}

# Rollback git tag if something fails
rollback_tag() {
    local version=$1
    local tag="v$version"

    warn "Rolling back tag $tag..."

    if git rev-parse "$tag" >/dev/null 2>&1; then
        git tag -d "$tag" || true
        info "Local tag $tag deleted"
    fi
}

# Main function
main() {
    local bump_type="patch"
    local new_version=""
    local current_version=""

    # Parse arguments
    while [[ $# -gt 0 ]]; do
        case $1 in
            --help|-h)
                show_help
                exit 0
                ;;
            --dry-run)
                DRY_RUN=true
                shift
                ;;
            --no-push)
                NO_PUSH=true
                shift
                ;;
            major|minor|patch)
                bump_type=$1
                shift
                ;;
            [0-9]*.[0-9]*.[0-9]*)
                new_version=$1
                bump_type="explicit"
                shift
                ;;
            *)
                error "Unknown argument: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Print mode
    if [[ "$DRY_RUN" == "true" ]]; then
        warn "DRY RUN MODE - No changes will be made"
    fi

    if [[ "$NO_PUSH" == "true" ]]; then
        warn "NO PUSH MODE - Changes will be local only"
    fi

    # Get current version
    current_version=$(get_latest_version)
    info "Current version: $current_version"

    # Calculate new version
    if [[ -z "$new_version" ]]; then
        new_version=$(bump_version "$current_version" "$bump_type")
    else
        validate_version "$new_version" || exit 1
    fi

    info "New version: $new_version"

    # Check for clean git state
    if ! check_git_clean "." "App repository"; then
        exit 1
    fi

    if ! check_git_clean "$OPS_REPO_PATH" "Ops repository"; then
        exit 1
    fi

    # Show summary
    echo ""
    echo "========================================="
    echo "Release Summary"
    echo "========================================="
    echo "Current version:  $current_version"
    echo "New version:      $new_version"
    echo "Bump type:        ${bump_type:-explicit}"
    echo "Docker image:     $REGISTRY/$APP_NAME:$new_version"
    echo "Git tag:          v$new_version"
    echo "Ops repo:         $OPS_REPO_PATH"
    echo "========================================="
    echo ""

    if [[ "$DRY_RUN" == "true" ]]; then
        info "DRY RUN - Skipping execution"
        exit 0
    fi

    # Execute release workflow
    info "Starting release workflow..."

    # Step 1: Build and push Docker image
    if ! build_and_push_image "$new_version"; then
        error "Failed to build and push Docker image"
        exit 1
    fi

    # Step 2: Create git tag
    if ! create_git_tag "$new_version"; then
        error "Failed to create git tag"
        exit 1
    fi

    # Step 3: Update ops repo
    if ! update_ops_repo "$new_version"; then
        error "Failed to update ops repository"
        # Rollback tag if not pushed
        if [[ "$NO_PUSH" == "true" ]]; then
            rollback_tag "$new_version"
        fi
        exit 1
    fi

    # Success
    echo ""
    echo "========================================="
    success "Release $new_version completed successfully!"
    echo "========================================="
    echo "Docker image: $REGISTRY/$APP_NAME:$new_version"
    echo "Git tag:      v$new_version"
    echo "Ops repo:     Updated and pushed"
    echo ""
    echo "The new version will be deployed automatically."
    echo "========================================="
}

# Run main function
main "$@"
