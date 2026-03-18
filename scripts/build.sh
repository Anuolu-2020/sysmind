#!/bin/bash

# Development Build Script
# Builds the application locally for testing

set -e

# Colors for output
BLUE='\033[0;34m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
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

error() {
    echo -e "${RED}❌ $1${NC}"
    exit 1
}

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -f "wails.json" ]; then
    error "Please run this script from the project root directory"
fi

# Check if Wails CLI is installed
if ! command -v wails >/dev/null 2>&1; then
    error "Wails CLI is not installed. Run: go install github.com/wailsapp/wails/v2/cmd/wails@latest"
fi

# Parse command line options
BUILD_MODE="dev"
PLATFORM=""
CLEAN=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --prod|--production)
            BUILD_MODE="prod"
            shift
            ;;
        --platform)
            PLATFORM="$2"
            shift 2
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        --help|-h)
            echo "Usage: $0 [OPTIONS]"
            echo ""
            echo "Options:"
            echo "  --prod, --production    Build for production"
            echo "  --platform PLATFORM     Target platform (e.g., linux/amd64, windows/amd64)"
            echo "  --clean                 Clean build directory first"
            echo "  --help, -h             Show this help message"
            echo ""
            echo "Examples:"
            echo "  $0                     # Development build"
            echo "  $0 --prod              # Production build"
            echo "  $0 --prod --clean      # Clean production build"
            echo "  $0 --platform linux/amd64  # Cross-compile for Linux"
            exit 0
            ;;
        *)
            error "Unknown option: $1"
            ;;
    esac
done

info "Starting SysMind build..."

# Install frontend dependencies if needed
if [ ! -d "frontend/node_modules" ]; then
    info "Installing frontend dependencies..."
    cd frontend && npm install && cd ..
fi

# Clean build directory if requested
if [ "$CLEAN" = true ]; then
    info "Cleaning build directory..."
    rm -rf build/
fi

# Install Go dependencies
info "Installing Go dependencies..."
go mod download

# Build command
BUILD_CMD="wails build"

if [ "$BUILD_MODE" = "prod" ]; then
    BUILD_CMD="$BUILD_CMD -clean -upx -s"
    info "Building for production with optimizations..."
else
    info "Building for development..."
fi

if [ -n "$PLATFORM" ]; then
    BUILD_CMD="$BUILD_CMD -platform $PLATFORM"
    info "Target platform: $PLATFORM"
fi

# Execute build
BUILD_FLAGS=""

# Get version information for ldflags
VERSION=$(git describe --tags --always 2>/dev/null || echo "dev")
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
GIT_TAG=$(git describe --tags --exact-match 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
BUILD_USER=$(whoami)

# Build ldflags
LDFLAGS="-X 'sysmind/internal/version.Version=${VERSION}' \
         -X 'sysmind/internal/version.GitCommit=${GIT_COMMIT}' \
         -X 'sysmind/internal/version.GitTag=${GIT_TAG}' \
         -X 'sysmind/internal/version.BuildDate=${BUILD_DATE}' \
         -X 'sysmind/internal/version.BuildUser=${BUILD_USER}'"

if [ "$BUILD_MODE" = "prod" ]; then
    BUILD_CMD="$BUILD_CMD -clean -upx -s -ldflags \"$LDFLAGS\""
    info "Building for production with optimizations..."
else
    BUILD_CMD="$BUILD_CMD -ldflags \"$LDFLAGS\""
    info "Building for development..."
fi

if [ -n "$PLATFORM" ]; then
    BUILD_CMD="$BUILD_CMD -platform $PLATFORM"
    info "Target platform: $PLATFORM"
fi

info "Version: $VERSION"
info "Git Commit: ${GIT_COMMIT:0:8}"
info "Build Date: $BUILD_DATE"

eval $BUILD_CMD

if [ $? -eq 0 ]; then
    success "Build completed successfully!"
    
    # Show build info
    if [ -f "build/bin/sysmind" ]; then
        BINARY_PATH="build/bin/sysmind"
    elif [ -f "build/bin/sysmind.exe" ]; then
        BINARY_PATH="build/bin/sysmind.exe"
    else
        warning "Binary not found in expected location"
        exit 1
    fi
    
    info "Binary location: $BINARY_PATH"
    info "Binary size: $(du -h "$BINARY_PATH" | cut -f1)"
    
    if [ "$BUILD_MODE" = "dev" ]; then
        echo ""
        echo "To run the application:"
        echo "  ./$BINARY_PATH"
        echo ""
        echo "Or start in development mode:"
        echo "  wails dev"
    fi
else
    error "Build failed!"
fi