# Scripts

Automation scripts for provider-mailgun development and release management.

## release.sh

Automated release script that consistently updates all version references across the codebase.

### Usage

```bash
# Create a new release
./scripts/release.sh v0.8.2

# The script will:
# 1. Validate version format (vX.Y.Z)
# 2. Check for uncommitted changes
# 3. Update VERSION file
# 4. Update all documentation and example references
# 5. Create git commit with standardized message
# 6. Create git tag
# 7. Display next steps for pushing and monitoring
```

### What gets updated

- `VERSION` file
- `CLAUDE.md` - All version references in examples and deployment info
- `README.md` - Build commands and examples
- `package/crossplane.yaml` - Controller image reference
- `examples/provider-config.yaml` - Package reference

### Next steps after running

1. Review changes: `git show --name-only`
2. Push commit: `git push`
3. Push tag: `git push origin v0.8.2`
4. Monitor workflow: `gh run list`

The GitHub Actions release workflow will automatically:
- Build and push Docker images (`versioned` + `latest`)
- Publish Crossplane packages
- Create GitHub release with notes
