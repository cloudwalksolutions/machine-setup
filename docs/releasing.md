# Releasing (maintainers)

Every merge to `main` is a release. `.github/workflows/release.yml` computes the
next patch tag from the latest `v*` tag, runs GoReleaser (darwin/linux ×
amd64/arm64 archives + checksums, a GitHub Release, the `tars` cask in
`cloudwalksolutions/homebrew-tap`), refreshes the coverage badge, and pushes the
tag. A run that fails after GoReleaser published the GitHub Release deletes that
release and its tag, so it can simply be re-run (also available via
`workflow_dispatch`).

Merges that only touch Markdown, `docs/`, or `vhs/` do not release.

Bumping minor or major is manual: push the tag yourself and the next merge
continues patching from it.

```bash
git tag v0.2.0 && git push origin v0.2.0   # no workflow runs on tags; the next merge releases v0.2.1
```

Notes:

- The tap push uses `HOMEBREW_TAP_DEPLOY_KEY`, a write deploy key on `homebrew-tap`
  that `devops-admin`'s Terraform creates for every repo in its `cli_tools`. To
  rotate, taint its `tls_private_key.homebrew_tap["machine-setup"]` there.
- `skip_upload: auto` in `.goreleaser.yaml` means snapshots and prerelease tags
  skip the tap push — only full releases update `brew`.
- GoReleaser cannot overwrite existing release assets: to redo a release, delete
  it and its tag (`gh release delete vX.Y.Z --yes --cleanup-tag`) and re-run the
  workflow.

Test the config locally without tagging (builds into `./dist`):

```bash
goreleaser release --snapshot --clean
```
