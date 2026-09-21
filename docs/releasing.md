# Releasing (maintainers)

Releases are cut by GoReleaser on a semver tag:

```bash
git tag v0.2.0
git push origin v0.2.0     # triggers .github/workflows/release.yml
```

This builds darwin/linux × amd64/arm64 archives + checksums, publishes a GitHub
Release, and pushes the `tars` cask to `cloudwalksolutions/homebrew-tap`.

Notes:

- The tap push authenticates with the `HOMEBREW_TAP_TOKEN` Actions secret — a
  fine-grained PAT scoped to `homebrew-tap` (Contents: read/write). Both the tap
  repo and the secret are already provisioned; to rotate, mint a new PAT and
  `gh secret set HOMEBREW_TAP_TOKEN -R cloudwalksolutions/machine-setup`.
- The tap repo itself is Terraform-managed in `devops-admin`
  (`terraform/terraform.auto.tfvars`, `open_source_repos`).
- `skip_upload: auto` in `.goreleaser.yaml` means snapshots and prerelease tags
  skip the tap push — only full releases update `brew`.
- GoReleaser cannot overwrite existing release assets: to re-run a release for
  the same tag, delete the release first (`gh release delete vX.Y.Z --yes`,
  the tag survives) and re-run the workflow.

Test the config locally without tagging (builds into `./dist`):

```bash
HOMEBREW_TAP_TOKEN=x goreleaser release --snapshot --clean
```
