# CI/CD Validation

## Workflow Boundaries

- CI validates generation, lint, tests, security scans, and builds without publishing.
- Release publication is triggered only by an exact `vMAJOR.MINOR.PATCH` tag at the current `origin/master` commit.
- The release workflow publishes the dual-platform xpkg to `ghcr.io/rossigee/provider-mailgun`, aliases `latest`, verifies equal digests and both platforms, and creates the GitHub Release.

## Release Preparation

```bash
make reviewable
make build
make build.all build.artifacts VERSION=v0.23.2 PLATFORMS="linux_amd64 linux_arm64"
for platform in linux_amd64 linux_arm64; do
  make xpkg.build VERSION=v0.23.2 PLATFORMS="linux_amd64 linux_arm64" PLATFORM="$platform"
done
```

Open a pull request from `release/v0.23.2` and wait for review, CI, and security checks. Do not publish from a branch push.
