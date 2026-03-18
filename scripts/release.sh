#!/bin/bash

# SysMind Release Script
# Usage: ./scripts/release.sh [version]
# Example: ./scripts/release.sh 1.2.3

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Check if version is provided
if [ $# -eq 0 ]; then
    error "Please provide a version number (e.g., 1.2.3)"
fi

VERSION=$1

# Validate version format (semantic versioning)
if ! [[ $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
    error "Invalid version format. Please use semantic versioning (e.g., 1.2.3)"
fi

TAG="v$VERSION"

info "Preparing release $TAG"

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "wails.json" ]; then
    error "Please run this script from the project root directory"
fi

# Check if git working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    error "Git working directory is not clean. Please commit or stash changes first."
fi

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    warning "You're not on the main branch (current: $CURRENT_BRANCH)"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# Check if tag already exists
if git tag -l | grep -q "^$TAG$"; then
    error "Tag $TAG already exists"
fi

# Update version in wails.json
info "Updating version in wails.json..."
if command -v jq >/dev/null 2>&1; then
    jq --arg version "$VERSION" '.info.productVersion = $version' wails.json > wails.json.tmp && mv wails.json.tmp wails.json
else
    # Fallback if jq is not available
    sed -i.bak "s/\"productVersion\": \".*\"/\"productVersion\": \"$VERSION\"/" wails.json && rm wails.json.bak
fi

# Update version in package.json
info "Updating version in frontend/package.json..."
cd frontend
if command -v jq >/dev/null 2>&1; then
    jq --arg version "$VERSION" '.version = $version' package.json > package.json.tmp && mv package.json.tmp package.json
else
    # Fallback if jq is not available
    sed -i.bak "s/\"version\": \".*\"/\"version\": \"$VERSION\"/" package.json && rm package.json.bak
fi
cd ..

# Generate changelog entry
info "Generating changelog entry..."

# Get the previous tag
PREV_TAG=$(git tag --sort=-version:refname | head -n1)

# Create changelog entry
CHANGELOG_ENTRY=""
if [ -n "$PREV_TAG" ]; then
    info "Generating changelog from $PREV_TAG to HEAD..."
    CHANGELOG_ENTRY=$(git log --pretty=format:"- %s (%h)" $PREV_TAG..HEAD | head -20)
else
    info "Generating changelog for initial release..."
    CHANGELOG_ENTRY=$(git log --pretty=format:"- %s (%h)" HEAD | head -20)
fi

# Create or update CHANGELOG.md
if [ ! -f "CHANGELOG.md" ]; then
    cat > CHANGELOG.md << EOF
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [$VERSION] - $(date +%Y-%m-%d)

### Changes
$CHANGELOG_ENTRY

EOF
else
    # Prepend to existing changelog
    {
        echo "## [$VERSION] - $(date +%Y-%m-%d)"
        echo
        echo "### Changes"
        echo "$CHANGELOG_ENTRY"
        echo
        cat CHANGELOG.md
    } > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md
fi

# Commit version changes
info "Committing version changes..."
git add wails.json frontend/package.json CHANGELOG.md
git commit -m "chore: bump version to $VERSION

- Update wails.json productVersion
- Update frontend/package.json version  
- Update CHANGELOG.md with release notes"

# Create and push tag
info "Creating git tag $TAG..."
git tag -a "$TAG" -m "Release $TAG

$(echo "$CHANGELOG_ENTRY" | head -10)"

# Push changes and tag
info "Pushing changes and tag to origin..."
git push origin main
git push origin "$TAG"

success "Release $TAG created successfully!"
info "GitHub Actions will now build and publish the release automatically."
info "Check the progress at: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/actions"

# Optional: Open release page in browser
if command -v xdg-open >/dev/null 2>&1; then
    REPO_URL=$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/https:\/\/github.com\/\1/')
    info "Opening release page..."
    xdg-open "$REPO_URL/releases/tag/$TAG" 2>/dev/null || true
elif command -v open >/dev/null 2>&1; then
    REPO_URL=$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/https:\/\/github.com\/\1/')
    info "Opening release page..."
    open "$REPO_URL/releases/tag/$TAG" 2>/dev/null || true
fi