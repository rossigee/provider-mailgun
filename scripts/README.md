# Scripts

## release.sh

`release.sh` validates a release version, confirms the worktree is clean and aligned with `origin/master`, rejects existing local or remote tags, and creates the annotated tag. It does not modify files, commit, push directly to `master`, or publish a plain image.

```bash
./scripts/release.sh v0.23.2
```

The tag-only GitHub Actions workflow performs the dual-architecture xpkg build and publication after the release PR is merged.
