# Releasing

This fork's release artifacts are the files that Coder Desktop downloads from
`https://storage.googleapis.com/coder-desktop/mutagen/<version>/`:
`mutagen-windows-{amd64,arm64}.exe`, `mutagen-darwin-{amd64,arm64}`, and a
trimmed `mutagen-agents.tar.gz`. All of them are produced by the
[release workflow](.github/workflows/release.yml).

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
3. Run the [Release workflow](.github/workflows/release.yml) manually from
   the Actions tab on `main`, with the tag (for example, `v0.18.4`) as the
   `tag` input. It checks that the tag is on `main` and matches the version,
   builds the artifacts from the tag, and creates a GitHub release with them
   attached. The publishing job runs in the `release` environment, so the run
   waits for approval before the release is created.

   Running the workflow without a `tag` builds the selected branch and
   attaches the artifacts to the workflow run only, which is useful for a dry
   run.
4. Run the [Upload release workflow](.github/workflows/upload-release.yml)
   manually from the Actions tab on `main`, with the same tag as the `tag`
   input. It downloads the five release payloads and `SHA256SUMS`, verifies
   the complete checksum manifest, and uploads the files to
   `gs://coder-desktop/mutagen/<tag>/` without rebuilding them. It also runs
   in the `release` environment and waits for approval, and it uses the
   bucket permissions described below.

   Uploads use `gcloud storage cp` with generation preconditions to refuse
   overwrites, and upload `SHA256SUMS` last. Re-running a successful upload
   fails; after a partial failure, an operator must inspect and remove the
   partial upload before retrying.

5. Bump the Mutagen version in Coder Desktop: `$mutagenVersion` in
   `scripts/Get-Mutagen.ps1` (coder/coder-desktop-windows) and
   `Coder-Desktop/Resources/.mutagenversion` (coder/coder-desktop-macos).

## Upload setup

The upload workflow authenticates to Google Cloud through Workload Identity
Federation and runs in the `release` environment, which only `main` may use.
It reads these repository variables:

- `GCP_CODE_SIGNING_WORKLOAD_ID_PROVIDER`
- `GCP_CODE_SIGNING_SERVICE_ACCOUNT`:
  `mutagen@coder-ci.iam.gserviceaccount.com`

The service account needs two roles on the `coder-desktop` bucket:

- `roles/storage.objectCreator`, with an IAM condition limiting writes to
  objects whose resource name starts with
  `projects/_/buckets/coder-desktop/objects/mutagen/`
- `roles/storage.objectViewer`, without a condition

The bucket is in project `coder-ci`. Its resource and IAM policy are managed
in coder/gcp under `projects/production/coder-ci`.
