#!/bin/bash

# Steam Account Switcher - Build Script
# Cross-compiles the Go application for Windows from macOS/Linux

set -e  # Exit on any error

echo "🔨 Building Steam Account Switcher for Windows..."

# Set build environment for Windows cross-compilation
export GOOS=windows
export GOARCH=amd64

# Build the application
echo "📦 Compiling with Go..."
go build -o steam-account-switcher.exe

# Check if build was successful
if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo "📁 Executable created: steam-account-switcher.exe"
    echo "📏 File size: $(ls -lh steam-account-switcher.exe | awk '{print $5}')"
    echo ""
    echo "🚀 Ready to deploy to Windows!"
else
    echo "❌ Build failed!"
    exit 1
fi 