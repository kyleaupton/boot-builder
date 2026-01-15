#!/bin/bash
# Build and install the privileged helper for development
# Run this once (or when helper code changes)

set -e

CERT_NAME="Boot Builder Dev Code Signing"
KC=~/Library/Keychains/login.keychain-db
HELPER_NAME="dev.kyleupton.boot-builder.helper"
HELPER_SRC="build/helpers/${HELPER_NAME}"
HELPER_DEST="/Library/PrivilegedHelperTools/${HELPER_NAME}"
PLIST_DEST="/Library/LaunchDaemons/${HELPER_NAME}.plist"

echo "Building privileged helper for development..."
echo ""

# Check if certificate exists
if ! security find-certificate -c "${CERT_NAME}" -a "${KC}" >/dev/null 2>&1; then
    echo "❌ Certificate '${CERT_NAME}' not found!"
    echo ""
    echo "You need to create a code signing certificate first."
    echo "See DEV_SETUP.md for instructions on creating it via Keychain Access."
    exit 1
fi

echo "Using certificate: ${CERT_NAME}"
echo ""

# Build helper
echo "Building helper..."
cd cmd/privileged-helper
task clean >/dev/null 2>&1 || true
task compile
task create:bundle

# Sign with dev certificate
echo ""
echo "Signing helper with dev certificate..."
codesign --force --sign "${CERT_NAME}" --options runtime \
    "../../${HELPER_SRC}/Contents/MacOS/${HELPER_NAME}"

codesign --force --sign "${CERT_NAME}" \
    "../../${HELPER_SRC}"

cd ../..

# Verify signature
echo ""
echo "Verifying helper signature..."
codesign -vvv --deep --strict "${HELPER_SRC}"

# Unload existing helper
echo ""
echo "Unloading existing helper (if running)..."
sudo launchctl unload "${PLIST_DEST}" 2>/dev/null || true

# Install helper
echo ""
echo "Installing helper to ${HELPER_DEST}..."
echo "⚠️  You'll be prompted for your password"
sudo rm -rf "${HELPER_DEST}"
sudo cp -r "${HELPER_SRC}" "${HELPER_DEST}"

# Install launchd plist
echo ""
echo "Installing launchd plist..."
sudo cp "${HELPER_SRC}/Contents/Resources/launchd.plist" "${PLIST_DEST}"

# Load helper
echo ""
echo "Loading helper..."
sudo launchctl load "${PLIST_DEST}"

# Wait a moment for it to start
sleep 1

# Verify it's running
if sudo launchctl list | grep -q "${HELPER_NAME}"; then
    echo ""
    echo "✅ Helper installed and running!"
    echo ""
    sudo launchctl list | grep "${HELPER_NAME}"
else
    echo ""
    echo "⚠️  Helper installed but not running!"
    echo ""
    echo "Check logs:"
    echo "  sudo launchctl list | grep ${HELPER_NAME}"
    echo "  log stream --predicate 'process == \"${HELPER_NAME}\"' --level debug"
    exit 1
fi

echo ""
echo "Monitor logs:"
echo "  log stream --predicate 'process == \"${HELPER_NAME}\"' --level debug"
echo ""
echo "To uninstall:"
echo "  sudo launchctl unload ${PLIST_DEST}"
echo "  sudo rm ${HELPER_DEST}"
echo "  sudo rm ${PLIST_DEST}"
