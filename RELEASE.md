# Release Process Guide

This document outlines the complete process for releasing new versions of SysMind.

## 🏗️ CI/CD Overview

SysMind uses GitHub Actions for automated testing, building, and releasing:

- **CI Pipeline** (`.github/workflows/ci.yml`): Runs on every push and PR
- **Release Pipeline** (`.github/workflows/release.yml`): Triggers on version tags

### CI Pipeline Features
- ✅ **Multi-OS Testing**: Linux, macOS, Windows
- ✅ **Code Quality**: golangci-lint, frontend linting
- ✅ **Security Scanning**: Gosec security analysis
- ✅ **Test Coverage**: Automated coverage reports
- ✅ **Dependency Scanning**: Automated vulnerability checks

### Release Pipeline Features
- 🚀 **Cross-Platform Builds**: 5 platform targets
- 📦 **Automated Packaging**: Compressed archives
- 📝 **Auto-Generated Changelogs**: From commit history
- 🔒 **Signed Releases**: GitHub-verified builds
- 📊 **Build Artifacts**: UPX-compressed binaries

## 📋 Supported Platforms

| Platform | Architecture | File Extension | Package Format |
|----------|-------------|----------------|----------------|
| Linux | AMD64 | - | `.tar.gz` |
| Linux | ARM64 | - | `.tar.gz` |
| macOS | Intel (AMD64) | - | `.tar.gz` |
| macOS | Apple Silicon (ARM64) | - | `.tar.gz` |
| Windows | AMD64 | `.exe` | `.zip` |

## 🚀 How to Release New Versions

### Prerequisites
- ✅ Clean git working directory
- ✅ All changes committed and pushed
- ✅ On `main` branch (recommended)
- ✅ CI pipeline passing
- ✅ `jq` installed (optional, for JSON parsing)

### Quick Release (Automated Script)

The easiest way to create a new release:

```bash
# Check current version and get suggestions
./scripts/version-check.sh

# Create and publish a new release
./scripts/release.sh 1.2.3
```

### Manual Release Process

If you prefer manual control:

#### Step 1: Prepare the Release
```bash
# Ensure clean state
git status
git pull origin main

# Check current version
./scripts/version-check.sh
```

#### Step 2: Update Version Numbers
```bash
# Update wails.json
jq '.info.productVersion = "1.2.3"' wails.json > temp.json && mv temp.json wails.json

# Update frontend/package.json  
cd frontend
jq '.version = "1.2.3"' package.json > temp.json && mv temp.json package.json
cd ..
```

#### Step 3: Generate Changelog
```bash
# Create/update CHANGELOG.md with new version entry
# Include notable changes, bug fixes, and new features
```

#### Step 4: Commit and Tag
```bash
# Commit version changes
git add wails.json frontend/package.json CHANGELOG.md
git commit -m "chore: bump version to 1.2.3"

# Create and push tag
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin main
git push origin v1.2.3
```

#### Step 5: Monitor Release
- GitHub Actions will automatically trigger
- Monitor progress in the Actions tab
- Release will be published automatically when builds complete

## 📝 Version Numbering

SysMind follows [Semantic Versioning](https://semver.org/):

```
MAJOR.MINOR.PATCH
```

### When to Increment:
- **MAJOR** (1.0.0 → 2.0.0): Breaking changes, API changes
- **MINOR** (1.0.0 → 1.1.0): New features, backward compatible
- **PATCH** (1.0.0 → 1.0.1): Bug fixes, backward compatible

### Examples:
- `v0.1.0` - Initial release
- `v0.1.1` - Bug fix release
- `v0.2.0` - New features added
- `v1.0.0` - Stable API, production ready
- `v2.0.0` - Breaking changes introduced

## 🛠️ Development Scripts

### Build Script
```bash
# Development build
./scripts/build.sh

# Production build
./scripts/build.sh --prod

# Cross-compile for specific platform
./scripts/build.sh --prod --platform linux/amd64

# Clean build
./scripts/build.sh --prod --clean
```

### Version Management
```bash
# Check current version and get suggestions
./scripts/version-check.sh

# Create a new release
./scripts/release.sh 1.2.3
```

## 🔍 Quality Gates

Before releasing, ensure all quality gates pass:

### Automated Checks (CI)
- ✅ All tests pass (`go test ./...`)
- ✅ Code linting passes (`golangci-lint run`)
- ✅ Security scan clean (`gosec`)
- ✅ Frontend tests pass (`npm test`)
- ✅ Build succeeds on all platforms

### Manual Checks
- ✅ Test core functionality locally
- ✅ Verify AI providers work correctly
- ✅ Check system monitoring accuracy
- ✅ Test on multiple platforms if possible
- ✅ Review changelog for accuracy
- ✅ Verify documentation is up-to-date

## 📦 Release Artifacts

Each release automatically generates:

### Binary Packages
- `sysmind-v1.2.3-linux-amd64.tar.gz`
- `sysmind-v1.2.3-linux-arm64.tar.gz`
- `sysmind-v1.2.3-darwin-amd64.tar.gz`
- `sysmind-v1.2.3-darwin-arm64.tar.gz`
- `sysmind-v1.2.3-windows-amd64.zip`

### Package Contents
- Compressed, optimized binary (UPX)
- Platform-specific executable
- SHA256 checksums
- Release notes and changelog

## 🚨 Rollback Process

If a release has critical issues:

### Immediate Actions
1. **Mark as Pre-release**: Edit GitHub release, check "pre-release"
2. **Add Warning**: Update release description with issue details
3. **Prepare Hotfix**: Create hotfix branch for urgent fixes

### Hotfix Release
```bash
# Create hotfix branch
git checkout -b hotfix/1.2.4 v1.2.3

# Make critical fixes
git commit -m "fix: critical issue description"

# Release hotfix
./scripts/release.sh 1.2.4
```

## 📊 Release Metrics

Monitor these metrics for each release:

- **Download counts** by platform
- **Issue reports** related to new version
- **Performance metrics** from user feedback
- **Adoption rate** across versions

## 🔗 Integration Points

### GitHub Repository Settings
Ensure these are configured:
- **Branch Protection**: Require PR reviews for `main`
- **Required Checks**: CI pipeline must pass
- **Merge Settings**: Squash merges preferred
- **Release Permissions**: Admin/write access required

### External Services
- **Codecov**: Automatic coverage reporting
- **Security Scanning**: Dependabot and Gosec integration
- **Package Registries**: Future integration with package managers

## 📞 Support Channels

After release, monitor these channels:
- 🐛 GitHub Issues for bug reports
- 💬 GitHub Discussions for questions  
- 📊 Download metrics and usage analytics
- 📧 Email feedback from users

---

## 🎯 Quick Reference

### Create Release Commands
```bash
# Check what's changed
./scripts/version-check.sh

# Release patch version (bug fixes)
./scripts/release.sh 1.2.4

# Release minor version (new features) 
./scripts/release.sh 1.3.0

# Release major version (breaking changes)
./scripts/release.sh 2.0.0
```

### Monitor Release
- **Actions**: https://github.com/yourusername/sysmind/actions
- **Releases**: https://github.com/yourusername/sysmind/releases
- **Issues**: https://github.com/yourusername/sysmind/issues

### Emergency Contacts
- Release Manager: [Your Email]
- CI/CD Admin: [Admin Email]
- Security Contact: [Security Email]