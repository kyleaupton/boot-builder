# Development Setup for XPC Privileged Helper

This guide explains how to set up your development environment to work with the privileged helper without needing to build for production each time.

## Overview

The app uses an XPC privileged helper to perform disk operations without requiring the entire app to run as root. For development, we need:

1. A self-signed code signing certificate
2. A manually installed helper (stays running between dev sessions)
3. Auto-signing of the main binary after each build

## One-Time Setup

### Step 1: Create Development Certificate

You need to manually create a code signing certificate using Keychain Access:

1. Open **Keychain Access** (Applications → Utilities → Keychain Access)
2. From the menu: **Keychain Access** → **Certificate Assistant** → **Create a Certificate...**
3. Fill in the form:
   - **Name**: `Boot Builder Dev Code Signing` (exactly this name)
   - **Identity Type**: Self Signed Root
   - **Certificate Type**: Code Signing ⚠️ **Make sure this says "Code Signing" not "Root Certificate"**
   - Click **Create**
   - Click **Continue** when prompted about adding to keychain
   - Click **Done**

The certificate will be added to your login keychain and is valid for 1 year by default.

**Important**: The certificate must have "Code Signing" as its type. If you see "Self-signed root certificate" in the details, delete it and create it again, making sure to select "Code Signing" from the Certificate Type dropdown.

**Verify it worked:**

```bash
security find-certificate -c "Boot Builder Dev Code Signing" -a ~/Library/Keychains/login.keychain-db
```

You should see output showing the certificate details. If the command returns without error, the certificate was created successfully.

### Step 2: Install Privileged Helper

Build and install the helper (requires sudo password):

```bash
./scripts/install-helper-dev.sh
```

This:

- Builds the helper signed with your dev certificate
- Installs it to `/Library/PrivilegedHelperTools/`
- Loads it with launchd

**Verify it's running:**

```bash
sudo launchctl list | grep dev.kyleupton.boot-builder.helper
```

You should see it listed with a PID.

### Step 3: Update Helper Info.plist (Done)

The helper's `SMAuthorizedClients` is already configured to accept binaries signed with the dev certificate:

```xml
<key>SMAuthorizedClients</key>
<array>
    <string>identifier "dev.kyleupton.boot-builder"</string>
</array>
```

## Daily Development Workflow

Once setup is complete, your workflow is simple:

```bash
wails3 dev
```

**What happens automatically:**

1. Wails builds your Go binary
2. **Post-build hook signs the binary** with your dev certificate (`scripts/sign-dev-binary.sh`)
3. libwim.dylib is copied
4. App runs and connects to the already-running helper via XPC

**No sudo required!** (except for the one-time helper install)

## When to Reinstall Helper

You only need to run `./scripts/install-helper-dev.sh` again if:

- You modify helper code (`cmd/privileged-helper/*.m`)
- You update the helper's Info.plist
- You update the launchd.plist

For changes to the main app, the auto-signing hook handles it.

## Monitoring & Debugging

### View Helper Logs

```bash
log stream --predicate 'process == "dev.kyleupton.boot-builder.helper"' --level debug
```

### Check Helper Status

```bash
sudo launchctl list | grep dev.kyleupton.boot-builder.helper
```

### Restart Helper

```bash
sudo launchctl unload /Library/LaunchDaemons/dev.kyleupton.boot-builder.helper.plist
sudo launchctl load /Library/LaunchDaemons/dev.kyleupton.boot-builder.helper.plist
```

### Verify Binary Signature

```bash
codesign -vvv bin/boot-builder
```

Should show:

```
identifier=dev.kyleupton.boot-builder
...
valid on disk
satisfies its Designated Requirement
```

## Uninstalling Dev Helper

To clean up:

```bash
sudo launchctl unload /Library/LaunchDaemons/dev.kyleupton.boot-builder.helper.plist
sudo rm /Library/PrivilegedHelperTools/dev.kyleupton.boot-builder.helper
sudo rm /Library/LaunchDaemons/dev.kyleupton.boot-builder.helper.plist
```

## Testing Disk Operations

### Create a Test Disk Image

Instead of using a real USB drive:

```bash
hdiutil create -size 2g -fs FAT32 -volname "TEST" -type SPARSE test-usb.sparseimage
hdiutil attach test-usb.sparseimage
```

This creates `/dev/diskN` that you can safely test against.

**Note:** Your drive detection filters for USB protocol, so you may need to temporarily relax that filter to see disk images.

### Unmount When Done

```bash
hdiutil detach /dev/diskN
```

## Troubleshooting

### Certificate not showing as signing identity

If you created the certificate but `security find-certificate -c "Boot Builder Dev Code Signing" -a ~/Library/Keychains/login.keychain-db` fails:

1. Open Keychain Access and find "Boot Builder Dev Code Signing"
2. Double-click it and check if it says "Self-signed root certificate" in the title
3. If so, delete it (also delete the associated private key)
4. Create it again, making absolutely sure you select **"Code Signing"** from the Certificate Type dropdown
5. The dropdown has many options - scroll down to find "Code Signing"

### "SMJobBless failed: Error Domain=CFErrorDomainLaunchd Code=2"

This means the helper isn't installed or the signatures don't match.

**Fix:**

1. Verify dev certificate exists: `security find-certificate -c "Boot Builder Dev Code Signing" -a ~/Library/Keychains/login.keychain-db`
2. Reinstall helper: `./scripts/install-helper-dev.sh`
3. Verify binary is signed: `codesign -vvv bin/boot-builder`

### "XPC connection failed"

Check if helper is running:

```bash
sudo launchctl list | grep dev.kyleupton.boot-builder.helper
```

If not listed, reinstall with `./scripts/install-helper-dev.sh`.

Check Console.app for XPC errors.

### Binary Not Auto-Signing

The hook runs after every build. If it's not working:

1. Check `build/config.yml` has the signing step
2. Run manually: `./scripts/sign-dev-binary.sh bin/boot-builder`
3. Check script has execute permission: `ls -la scripts/sign-dev-binary.sh`

## Production Builds

For production, the helper is:

- Embedded in the app bundle at `Contents/Library/LaunchServices/`
- Signed with a proper Developer ID certificate (or ad-hoc for local testing)
- Installed via SMJobBless when first needed

The production flow is handled by `task package` in `build/darwin/Taskfile.yml`.
