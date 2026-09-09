#!/bin/bash

# Assembles the artifacts that Coder Desktop downloads from
# https://storage.googleapis.com/coder-desktop/mutagen/<version>/ into
# build/coder-desktop. It must be run after a release build (see build.sh).
#
# The agent bundle is limited to the platforms that Coder Desktop can
# realistically sync with, which keeps it around 17 MB instead of the roughly
# 100 MB of the full bundle.

# Exit immediately on failure.
set -e

# Compute the output directory and start with a clean slate.
MUTAGEN_CODER_DESKTOP_PATH="build/coder-desktop"
rm -rf "${MUTAGEN_CODER_DESKTOP_PATH}"
mkdir -p "${MUTAGEN_CODER_DESKTOP_PATH}"

# Copy the CLI binaries that the desktop apps embed.
cp build/cli/windows_amd64 "${MUTAGEN_CODER_DESKTOP_PATH}/mutagen-windows-amd64.exe"
cp build/cli/windows_arm64 "${MUTAGEN_CODER_DESKTOP_PATH}/mutagen-windows-arm64.exe"
cp build/cli/darwin_amd64 "${MUTAGEN_CODER_DESKTOP_PATH}/mutagen-darwin-amd64"
cp build/cli/darwin_arm64 "${MUTAGEN_CODER_DESKTOP_PATH}/mutagen-darwin-arm64"

# Build the trimmed agent bundle. COPYFILE_DISABLE keeps macOS tar from adding
# AppleDouble (._*) entries that the agent bundle reader doesn't expect.
COPYFILE_DISABLE=1 tar -czf "${MUTAGEN_CODER_DESKTOP_PATH}/mutagen-agents.tar.gz" -C build/agent \
    darwin_amd64 \
    darwin_arm64 \
    linux_amd64 \
    linux_arm64 \
    windows_amd64 \
    windows_arm64

# Compute SHA256 digests.
pushd "${MUTAGEN_CODER_DESKTOP_PATH}" > /dev/null
sha256sum mutagen-* > SHA256SUMS
popd > /dev/null
