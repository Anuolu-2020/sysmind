#!/bin/bash

# Version Checker Script
# Checks current version and suggests next version

set -e

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

success() {
    echo -e "${GREEN}✅ $1${NC}"
}

warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Get current version from git tags
CURRENT_TAG=$(git tag --sort=-version:refname | head -n1)

if [ -z "$CURRENT_TAG" ]; then
    info "No previous tags found. This will be the first release."
    success "Suggested version: 0.1.0"
    exit 0
fi

CURRENT_VERSION=${CURRENT_TAG#v}
info "Current version: $CURRENT_VERSION"

# Parse current version
IFS='.' read -ra VERSION_PARTS <<< "$CURRENT_VERSION"
MAJOR=${VERSION_PARTS[0]}
MINOR=${VERSION_PARTS[1]}
PATCH=${VERSION_PARTS[2]}

# Calculate next versions
NEXT_PATCH="$MAJOR.$MINOR.$((PATCH + 1))"
NEXT_MINOR="$MAJOR.$((MINOR + 1)).0"
NEXT_MAJOR="$((MAJOR + 1)).0.0"

# Show commit messages since last tag
info "Changes since $CURRENT_TAG:"
git log --pretty=format:"  - %s" $CURRENT_TAG..HEAD | head -10

echo
echo "Suggested next versions:"
echo "  🐛 Patch (bug fixes):     $NEXT_PATCH"
echo "  ✨ Minor (new features):  $NEXT_MINOR" 
echo "  💥 Major (breaking):      $NEXT_MAJOR"
echo

# Analyze commit messages for suggestions
COMMITS=$(git log --pretty=format:"%s" $CURRENT_TAG..HEAD)

if echo "$COMMITS" | grep -qi "BREAKING\|breaking"; then
    warning "Breaking changes detected - consider major version bump"
elif echo "$COMMITS" | grep -qi "feat\|feature\|add"; then
    success "New features detected - consider minor version bump"
elif echo "$COMMITS" | grep -qi "fix\|bug"; then
    success "Bug fixes detected - consider patch version bump"
fi

echo "To create a release, run:"
echo "  ./scripts/release.sh [version]"