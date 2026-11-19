# Releasing

This fork's release artifacts are the files that Coder Desktop downloads from
`https://storage.googleapis.com/coder-desktop/mutagen/<version>/`:
`mutagen-windows-{amd64,arm64}.exe`, `mutagen-darwin-{amd64,arm64}`, and a
trimmed `mutagen-agents.tar.gz`. All of them are produced by the
[release workflow](.github/workflows/release.yml), which does not sign them.

## Cutting a release

1. Bump the version in `pkg/mutagen/version.go` and merge that change.
2. Tag the merged commit and push the tag. The tag must match the version
   (`v` prefix plus `mutagen version` output). Only repository admins,
   maintainers and the `coder-desktop-release-managers` and
   `release-managers` teams may push `v*` tags (the `Release tags` ruleset
   in coder/github-org-terraform):

   ```bash
   git tag -a v0.18.4 -m v0.18.4
   git push origin v0.18.4
   ```

   The release workflow builds the artifacts and creates a GitHub release
   with them attached. Running the workflow manually from the Actions tab
   builds the same artifacts and attaches them to the workflow run instead,
   which is useful for a dry run.
3. Run the [Upload release workflow](.github/workflows/upload-release.yml)
   manually from the Actions tab on `main`, with the existing release tag (for example,
   `v0.18.4`) as the `tag` input. It downloads the five release payloads and
   `SHA256SUMS`, verifies the complete checksum manifest, and uploads the files
   to `gs://coder-desktop/mutagen/<tag>/` without rebuilding them.
   The bucket permissions described below must be in place first.

   Uploads use `gcloud storage cp` with generation preconditions to refuse
   overwrites, and upload `SHA256SUMS` last. Re-running a successful upload
   fails; after a partial failure, an operator must inspect and remove the
   partial upload before retrying.

4. Bump the Mutagen version in Coder Desktop: `$mutagenVersion` in
   `scripts/Get-Mutagen.ps1` (coder/coder-desktop-windows) and
   `Coder-Desktop/Resources/.mutagenversion` (coder/coder-desktop-macos).

## Upload setup

The upload workflow authenticates to Google Cloud through Workload Identity
Federation and runs in the `code-signing` environment, which only `main` and
`v*` tags may use. It needs these repository variables:

- `GCP_CODE_SIGNING_WORKLOAD_ID_PROVIDER`
- `GCP_CODE_SIGNING_SERVICE_ACCOUNT`:
  `mutagen@coder-ci.iam.gserviceaccount.com`

The service account needs `roles/storage.objectCreator` on the
`coder-desktop` bucket, with an IAM condition limiting writes to objects
whose resource name starts with
`projects/_/buckets/coder-desktop/objects/mutagen/`.

The account also needs `roles/storage.objectViewer` on the bucket so
`gcloud storage cp` can inspect destination objects and list the bucket.
This read binding must be unconditional: object-prefix conditions cannot
restrict bucket listing. It permits reading and listing the whole bucket,
while writes remain limited to `mutagen/`. Neither role permits deletion.

The `coder-desktop` bucket already exists in project `coder-ci`. Its bucket
resource, IAM policy, and import declarations are managed in coder/gcp under
`projects/production/coder-ci`. Add both uploader bindings to that existing
policy before running the upload workflow; do not create another bucket.
No service account key or new GitHub secret is needed.
