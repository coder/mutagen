# Releasing

This fork's release artifacts are the files that Coder Desktop downloads from
`https://storage.googleapis.com/coder-desktop/mutagen/<version>/`:
`mutagen-windows-{amd64,arm64}.exe`, `mutagen-darwin-{amd64,arm64}`, and a
trimmed `mutagen-agents.tar.gz`. All of them are produced by the
[release workflow](.github/workflows/release.yml).

## Cutting a release

1. Bump the version in `pkg/mutagen/version.go` and merge that change.
2. Run the [Release workflow](.github/workflows/release.yml) on `main` with
   the version (`v` prefix plus `mutagen version` output) and the full SHA of
   the merged commit:

   ```bash
   gh workflow run release.yml --repo coder/mutagen --ref main \
     -f version=v0.18.6 -f commit="$(git rev-parse origin/main)"
   ```
3. Ask another member of `coder-desktop-release-managers` to approve the run.
   It waits in the `release-approval` environment until then.
4. Once approved, the workflow creates the `vX.Y.Z` tag at the commit, builds
   the artifacts, publishes them with `SHA256SUMS` as a GitHub release, and
   uploads the same files to `gs://coder-desktop/mutagen/<version>/`,
   `SHA256SUMS` last.
5. Bump the Mutagen version and checksums in Coder Desktop:
   `scripts/mutagen.json` (coder/coder-desktop-windows), and
   `Coder-Desktop/Resources/.mutagenversion` plus the `MUTAGEN_SHA256_*`
   values in the `Makefile` (coder/coder-desktop-macos). The checksums come
   from the release's `SHA256SUMS`.

Running the workflow without a version builds the selected branch and
attaches the artifacts to the workflow run only, which is useful for a dry
run.

The workflow accepts a release when all of these hold:

- The run is on `main`, and the commit is on `main`.
- The version looks like `vX.Y.Z`, is higher than the latest release, and is
  the version in `pkg/mutagen/version.go` of the commit.
- The commit is not older than the latest release.

### When a release run fails

Re-run the failed jobs. The run goes on from what the earlier attempt did: it
keeps the tag, a GitHub release that holds the files of the run, and the
files already in the bucket.

Starting the same version for the same commit again also finds the tag. It
takes a new approval and builds the files again, and it stops if a published
GitHub release holds other files.

## Setup

| Environment        | Jobs                | Reviewers                        |
| ------------------ | ------------------- | -------------------------------- |
| `release-approval` | `approve`           | `coder-desktop-release-managers` |
| `release`          | `release`, `upload` | None                             |

Both environments only accept `main`.

The `approve` job creates the tag as the `coder-release-publisher` GitHub
App, with these secrets of the `release-approval` environment:

| Secret                                | What it is                                           |
| ------------------------------------- | ---------------------------------------------------- |
| `CODER_RELEASE_PUBLISHER_APP_ID`      | The App's client ID                                  |
| `CODER_RELEASE_PUBLISHER_APP_KMS_KEY` | Resource name of the App's key in Cloud KMS          |
| `APP_KEYS_WORKLOAD_ID_PROVIDER`       | Google Cloud workload identity provider for App keys |

The `upload` job authenticates to Google Cloud through Workload Identity
Federation, with these repository variables:

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
