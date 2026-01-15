#!/bin/bash
# Sign the dev binary after build
# Called automatically by Wails build hooks

BINARY_PATH="$1"
CERT_NAME="Boot Builder Dev Code Signing"
KC=~/Library/Keychains/login.keychain-db

if [ -z "${BINARY_PATH}" ]; then
    echo "Usage: $0 <path-to-binary>"
    exit 1
fi

if [ ! -f "${BINARY_PATH}" ]; then
    echo "Binary not found: ${BINARY_PATH}"
    exit 1
fi

# Check if certificate exists
if ! security find-certificate -c "${CERT_NAME}" -a "${KC}" >/dev/null 2>&1; then
    echo "⚠️  Dev certificate '${CERT_NAME}' not found!"
    echo "   See DEV_SETUP.md for instructions on creating it via Keychain Access."
    echo "   Skipping code signing..."
    exit 0
fi

echo "Signing ${BINARY_PATH} with ${CERT_NAME}..."
codesign --force --sign "${CERT_NAME}" \
    --identifier "dev.kyleupton.boot-builder" \
    --options runtime \
    "${BINARY_PATH}"

echo "✅ Binary signed successfully"
