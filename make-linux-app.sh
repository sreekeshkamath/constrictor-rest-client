#!/bin/bash

# Script to create a proper Linux app from the Wails build
# This creates a desktop entry and optionally installs the app

set -e

APP_NAME="constrictor-rest-client"
BUILD_DIR="wails/build/bin"
DESKTOP_FILE="wails/build/${APP_NAME}.desktop"
INSTALL_DIR="${HOME}/.local/share/applications"
BIN_INSTALL_DIR="${HOME}/.local/bin"

echo "🔨 Building Linux app..."
make wails-build-linux

if [ ! -f "${BUILD_DIR}/${APP_NAME}" ]; then
    echo "❌ Build failed! Binary not found at ${BUILD_DIR}/${APP_NAME}"
    exit 1
fi

echo "✅ Build successful!"

# Create desktop entry
echo "📝 Creating desktop entry..."
cat > "${DESKTOP_FILE}" << EOF
[Desktop Entry]
Version=1.0
Type=Application
Name=Constrictor REST Client
Comment=REST API testing tool - Native desktop application
Exec=${BIN_INSTALL_DIR}/${APP_NAME}
Icon=${APP_NAME}
Terminal=false
Categories=Development;Network;
StartupNotify=true
EOF

echo "✅ Desktop entry created at ${DESKTOP_FILE}"

# Ask user if they want to install
read -p "Do you want to install the app to ~/.local? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "📦 Installing app..."
    
    # Create directories if they don't exist
    mkdir -p "${BIN_INSTALL_DIR}"
    mkdir -p "${INSTALL_DIR}"
    
    # Copy binary
    cp "${BUILD_DIR}/${APP_NAME}" "${BIN_INSTALL_DIR}/${APP_NAME}"
    chmod +x "${BIN_INSTALL_DIR}/${APP_NAME}"
    echo "✅ Binary installed to ${BIN_INSTALL_DIR}/${APP_NAME}"
    
    # Copy desktop entry
    cp "${DESKTOP_FILE}" "${INSTALL_DIR}/${APP_NAME}.desktop"
    echo "✅ Desktop entry installed to ${INSTALL_DIR}/${APP_NAME}.desktop"
    
    # Update desktop database
    if command -v update-desktop-database >/dev/null 2>&1; then
        update-desktop-database "${INSTALL_DIR}" 2>/dev/null || true
        echo "✅ Desktop database updated"
    fi
    
    echo ""
    echo "🎉 App installed successfully!"
    echo "   You can now find 'Constrictor REST Client' in your application menu"
    echo "   Or run it from terminal: ${APP_NAME}"
else
    echo ""
    echo "📦 App built but not installed"
    echo "   Binary: ${BUILD_DIR}/${APP_NAME}"
    echo "   Desktop entry: ${DESKTOP_FILE}"
    echo ""
    echo "To install manually:"
    echo "  cp ${BUILD_DIR}/${APP_NAME} ~/.local/bin/"
    echo "  cp ${DESKTOP_FILE} ~/.local/share/applications/"
    echo "  update-desktop-database ~/.local/share/applications/"
fi
