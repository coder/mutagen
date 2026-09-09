#!/bin/bash

# Signs the Windows binary at the specified path (in place) with Coder's
# Extended Validation code signing certificate, whose private key lives in
# Google Cloud KMS. This is the same mechanism that coder/coder uses to sign
# its Windows releases.
#
# Usage: sign_windows.sh <path>
#
# Requires java and the following environment variables:
#  - JSIGN_PATH: The path to the jsign jar.
#  - EV_KEYSTORE: The Google Cloud KMS key ring containing the signing key.
#  - EV_KEY: The name of the signing key.
#  - EV_CERTIFICATE_PATH: The path to the PEM-encoded signing certificate.
#  - EV_TSA_URL: The URL of the RFC 3161 timestamp server to use.
#  - GCLOUD_ACCESS_TOKEN: A Google Cloud access token authorized to sign with
#    the key.

# Exit immediately on failure.
set -e

# Verify arguments.
if [[ "$#" -ne 1 ]]; then
    echo "usage: $0 <path>" 1>&2
    exit 1
fi

# Verify that the required environment variables are set.
for variable in JSIGN_PATH EV_KEYSTORE EV_KEY EV_CERTIFICATE_PATH EV_TSA_URL GCLOUD_ACCESS_TOKEN; do
    if [[ -z "${!variable}" ]]; then
        echo "${variable} must be set" 1>&2
        exit 1
    fi
done

# Perform signing. Output is redirected to standard error so that it doesn't
# interfere with the build script's output.
java -jar "${JSIGN_PATH}" \
    --storetype GOOGLECLOUD \
    --storepass "${GCLOUD_ACCESS_TOKEN}" \
    --keystore "${EV_KEYSTORE}" \
    --alias "${EV_KEY}" \
    --certfile "${EV_CERTIFICATE_PATH}" \
    --tsmode RFC3161 \
    --tsaurl "${EV_TSA_URL}" \
    "$1" \
    1>&2
