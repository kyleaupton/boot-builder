#!/bin/bash -e

# This script builds wimlib binaries for Linux, macOS. (Windows is not working)
# It is intended to be ran on a Apple Silicon Mac.
# You need to have the necessary build tools installed.
# You can do this by running the following command:
#   brew install autoconf automake libtool pkg-config FiloSottile/musl-cross/musl-cross
# (If Windows is working, then you need to install mingw-w64 too)

# Define the output directory for binaries
OUTPUT_DIR=$(pwd)/wimlib-binaries

# Create the necessary directory structure
mkdir -p "$OUTPUT_DIR/linux/wimlib/x64"
mkdir -p "$OUTPUT_DIR/mac/wimlib/x64"
mkdir -p "$OUTPUT_DIR/mac/wimlib/arm64"
mkdir -p "$OUTPUT_DIR/win/wimlib/x64"

#####################################
# Setup wimlib source
#####################################
git clone https://github.com/ebiggers/wimlib.git wimlib-source || true
cd wimlib-source || exit 1

# Get new tags from remote
git fetch --tags

# Get latest tag name
latestTag=$(git describe --tags "$(git rev-list --tags --max-count=1)")

# Checkout latest tag
git checkout $latestTag

# Define the source directory (change this if wimlib source is elsewhere)
WIMLIB_SRC_DIR=$(pwd)/wimlib-source

# Bootstrap the build system
bash bootstrap

# Clean previous builds
make clean || true

#####################################
# macOS (arm64)
#####################################
echo "Building for macOS arm64..."
./configure --prefix="$WIMLIB_SRC_DIR/wimlib-install-arm64" --enable-static --without-fuse --without-ntfs-3g
make -j$(sysctl -n hw.ncpu)
make install

# Copy macOS arm64 binaries
cp -r "$WIMLIB_SRC_DIR/wimlib-install-arm64/bin" "$OUTPUT_DIR/mac/wimlib/arm64/"

make clean

#####################################
# macOS (x86_64)
#####################################
echo "Building for macOS x86_64..."
./configure --prefix="$WIMLIB_SRC_DIR/wimlib-install-x64" --enable-static --without-fuse --without-ntfs-3g CC="clang -arch x86_64"
make -j$(sysctl -n hw.ncpu)
make install

# Copy macOS x86_64 binaries
cp -r "$WIMLIB_SRC_DIR/wimlib-install-x64/bin" "$OUTPUT_DIR/mac/wimlib/x64/"

make clean

#####################################
# Linux (x86_64)
#####################################
echo "Building for Linux x86_64..."
./configure --host=x86_64-linux-musl --prefix="$WIMLIB_SRC_DIR/wimlib-install-linux-x64" --enable-static --without-fuse --without-ntfs-3g
make -j$(nproc)
make install

# Copy Linux x86_64 binaries
cp -r "$WIMLIB_SRC_DIR/wimlib-install-linux-x64/bin" "$OUTPUT_DIR/linux/wimlib/x64/"

make clean

#####################################
# Windows (x86_64)
#####################################
# echo "Building for Windows x86_64..."
# ./configure --host=x86_64-w64-mingw32 --prefix="$WIMLIB_SRC_DIR/wimlib-install-win-x64" --enable-static --without-fuse --without-ntfs-3g
# make -j$(nproc)
# make install

# # Copy Windows x86_64 binaries
# cp -r "$WIMLIB_SRC_DIR/wimlib-install-win-x64/bin" "$OUTPUT_DIR/win/wimlib/x64/"

# make clean

# echo "All builds completed successfully. Binaries are located in $OUTPUT_DIR."
