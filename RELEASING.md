# Releasing

This fork's release artifacts are the files that Coder Desktop downloads from
`https://storage.googleapis.com/coder-desktop/mutagen/<version>/`:
`mutagen-windows-{amd64,arm64}.exe`, `mutagen-darwin-{amd64,arm64}`, and a
trimmed `mutagen-agents.tar.gz`. The macOS and Windows binaries (including
the agents inside the bundle) are signed with the same certificates as Coder
Desktop. All of them are produced by the
[release workflow](.github/workflows/release.yml).

## Cutting a release

1. Bump the version in `pkg/mutagen/version.go` and merge that change.
2. Tag the merged commit and push the tag. The tag must match the version
   (`v` prefix plus `mutagen version` output):

   ```bash
   git tag -a v0.18.4 -m v0.18.4
   git push origin v0.18.4
   ```

   The release workflow builds and signs the artifacts and creates a GitHub
   release with them attached. Running the workflow manually from the Actions
   tab builds the same artifacts from any ref and attaches them to the workflow
   run instead, which is useful for a dry run.
3. Copy the artifacts to the bucket that Coder Desktop reads from:

   ```bash
   gh release download v0.18.4 -R coder/mutagen -p 'mutagen-*' -D mutagen-v0.18.4
   gsutil cp mutagen-v0.18.4/* gs://coder-desktop/mutagen/v0.18.4/
   ```

4. Bump the Mutagen version in Coder Desktop: `$mutagenVersion` in
   `scripts/Get-Mutagen.ps1` (coder/coder-desktop-windows) and
   `Coder-Desktop/Resources/.mutagenversion` (coder/coder-desktop-macos).

## Signing setup

macOS binaries are signed with the `Developer ID Application: Coder
Technologies Inc` certificate (the identity is hardcoded in the workflow) and
need these secrets:

- `MACOS_CERTIFICATE`: base64-encoded PKCS#12 export of the certificate and
  its private key
- `MACOS_CERTIFICATE_PWD`: the PKCS#12 password

The binaries are not notarized here. Coder Desktop embeds them as-is and
notarizes the whole app, which is why they have to carry this signature.

Windows binaries use the same setup as coder/coder-desktop-windows: jsign with
a Google Cloud KMS keystore, authenticated through Workload Identity
Federation. It needs the following repository configuration:

- Variable `GCP_CODE_SIGNING_WORKLOAD_ID_PROVIDER`: same value as coder/coder
- Variable `GCP_CODE_SIGNING_SERVICE_ACCOUNT`:
  `mutagen@coder-ci.iam.gserviceaccount.com`
- Secrets `EV_SIGNING_CERT`, `EV_KEYSTORE`, `EV_KEY`, `EV_TSA_URL`: same values
  as coder/coder

The service account, its Cloud KMS roles, and its workload identity binding
for this repository are managed in coder/gcp under
`projects/production/coder-ci`.
