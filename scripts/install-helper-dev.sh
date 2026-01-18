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
# REMOVED: task create:bundle  ← Don't create a bundle!
cd ../..

# Verify it's a single executable, not a bundle
if [ -d "${HELPER_SRC}" ]; then
    echo "❌ ERROR: Helper is a directory/bundle, should be a single executable!"
    echo "   Check your build task - remove any bundle creation steps."
    exit 1
fi

if ! file "${HELPER_SRC}" | grep -q "Mach-O"; then
    echo "❌ ERROR: Helper is not a Mach-O executable!"
    file "${HELPER_SRC}"
    exit 1
fi

# Verify plists are embedded
echo ""
echo "Verifying embedded plists..."
if ! otool -s __TEXT __info_plist "${HELPER_SRC}" | grep -q "Contents"; then
    echo "❌ ERROR: Info.plist not embedded in binary!"
    echo "   Add -sectcreate __TEXT __info_plist to your clang command."
    exit 1
fi

if ! otool -s __TEXT __launchd_plist "${HELPER_SRC}" | grep -q "Contents"; then
    echo "❌ ERROR: launchd.plist not embedded in binary!"
    echo "   Add -sectcreate __TEXT __launchd_plist to your clang command."
    exit 1
fi
echo "✅ Plists embedded correctly"

# Sign with dev certificate
echo ""
echo "Signing helper with dev certificate..."
codesign --force --sign "${CERT_NAME}" --options runtime "${HELPER_SRC}"

# Verify signature
echo ""
echo "Verifying helper signature..."
codesign -vvv --strict "${HELPER_SRC}"

# Unload existing helper
echo ""
echo "Unloading existing helper (if running)..."
sudo launchctl bootout system/${HELPER_NAME} 2>/dev/null || true

# Remove old installation (might be a bundle from before)
echo ""
echo "Removing old installation..."
sudo rm -rf "${HELPER_DEST}"
sudo rm -f "${PLIST_DEST}"

# Install helper (single file, not directory)
echo ""
echo "Installing helper to ${HELPER_DEST}..."
echo "⚠️  You'll be prompted for your password"
sudo cp "${HELPER_SRC}" "${HELPER_DEST}"
sudo chown root:wheel "${HELPER_DEST}"
sudo chmod 544 "${HELPER_DEST}"

# Install launchd plist (add ProgramArguments for manual install)
echo ""
echo "Installing launchd plist..."
cat > /tmp/${HELPER_NAME}.plist << EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple Computer//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>${HELPER_NAME}</string>

    <key>ProgramArguments</key>
    <array>
        <string>${HELPER_DEST}</string>
    </array>

    <key>MachServices</key>
    <dict>
        <key>${HELPER_NAME}</key>
        <true/>
    </dict>
</dict>
</plist>
EOF

sudo cp /tmp/${HELPER_NAME}.plist "${PLIST_DEST}"
sudo chown root:wheel "${PLIST_DEST}"
sudo chmod 644 "${PLIST_DEST}"
rm /tmp/${HELPER_NAME}.plist

# Load helper
echo ""
echo "Loading helper..."
sudo launchctl bootstrap system "${PLIST_DEST}"

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
    echo "⚠️  Helper installed but may not be running yet (on-demand)"
    echo "   It will start when the app connects via XPC."
fi

# Final verification
echo ""
echo "Installation verification:"
echo "  Helper: $(file ${HELPER_DEST} | cut -d: -f2)"
echo "  Plist:  ${PLIST_DEST}"
ls -la "${HELPER_DEST}"
ls -la "${PLIST_DEST}"

echo ""
echo "Monitor logs:"
echo "  log stream --predicate 'process == \"${HELPER_NAME}\"' --level debug"
echo ""
echo "To uninstall:"
echo "  sudo launchctl bootout system/${HELPER_NAME}"
echo "  sudo rm ${HELPER_DEST}"
echo "  sudo rm ${PLIST_DEST}"