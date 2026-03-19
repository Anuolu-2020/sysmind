#!/bin/bash

# Generate sysmind.desktop file with proper paths
# This script is called during the build process to create a desktop entry file
# that users can install to integrate SysMind into their desktop environment

set -e

# Get the installation prefix (default to /usr/local for system-wide, or $HOME/.local for user)
PREFIX="${1:-.}"

# Determine icon path based on installation location
if [ "$PREFIX" = "." ]; then
    # Development/local build
    ICON_PATH="$(pwd)/build/icons/icon_512.png"
    EXEC_PATH="$(pwd)/build/bin/sysmind"
else
    # System installation
    ICON_PATH="$PREFIX/share/pixmaps/sysmind.png"
    EXEC_PATH="$PREFIX/bin/sysmind"
fi

# Get version info
VERSION="${VERSION:-dev}"
BUILD_DATE="$(date -u +"%Y-%m-%d")"

# Create the desktop file
cat > sysmind.desktop << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=SysMind
Comment=AI-powered system monitoring desktop application
Comment[en_US]=AI-powered system monitoring desktop application
Exec=$EXEC_PATH
Icon=sysmind
Terminal=false
Categories=System;Monitor;Utility;
StartupWMClass=sysmind
Keywords=system;monitor;performance;ai;cpu;memory;network;process;
X-AppImage-Version=$VERSION
X-AppImage-BuildDate=$BUILD_DATE
EOF

echo "✅ Generated sysmind.desktop with paths:"
echo "   Exec: $EXEC_PATH"
echo "   Icon: $ICON_PATH"
