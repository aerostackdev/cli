# Aerostack CLI — Enterprise Release Policy

> Versioning, branching, and release rules for the Aerostack CLI.

---

## Versioning (SemVer)

We follow [Semantic Versioning 2.0.0](https://semver.org/):

- **MAJOR** (x.0.0): Breaking changes, incompatible API changes
- **MINOR** (1.x.0): New features, backward-compatible
- **PATCH** (1.0.x): Bug fixes, backward-compatible

### Examples

| Change | Version bump |
|--------|--------------|
| Remove or rename a command | MAJOR |
| Add new subcommand | MINOR |
| Fix bug in `aerostack dev` | PATCH |
| Change default behavior | MAJOR |
| Add new flag (non-breaking) | MINOR |

---

## Branch Strategy

| Branch | Purpose |
|--------|---------|
| `main` | Production-ready code. All releases are cut from `main`. |
| `develop` | Integration branch for features (optional). |
| `feature/*` | New features (e.g. `feature/db-migrate`). |
| `fix/*` | Bug fixes (e.g. `fix/dev-port-binding`). |
| `release/v*` | Release preparation (e.g. `release/v1.0.0`). |

**Rule:** Only merge to `main` via Pull Request. All PRs must pass CI.

---

## Release Process

### 1. Pre-release checklist

- [ ] All tests pass
- [ ] CHANGELOG.md updated
- [ ] Version bumped in code (if applicable)
- [ ] No uncommitted changes on `main`

### 2. Create release

```bash
# Ensure you're on main and up to date
git checkout main
git pull origin main

# Create and push tag (triggers GitHub Actions)
git tag v1.0.0
git push origin v1.0.0
```

### 3. Automated steps (GitHub Actions)

- Builds binaries for Linux, macOS, Windows (amd64, arm64)
- Creates GitHub Release with changelog
- Uploads artifacts (tar.gz, zip)
- Optionally: Homebrew tap, npm (if applicable)

### 4. Post-release

- [ ] Verify release assets on [Releases](https://github.com/aerostackdev/cli/releases)
- [ ] Update install script / docs if needed
- [ ] Announce in team channels

---

## Commit Convention (Conventional Commits)

All commits should follow [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

### Types

| Type | Use for |
|------|---------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `chore` | Maintenance (deps, config) |
| `refactor` | Code change, no feature/fix |
| `test` | Adding or updating tests |
| `ci` | CI/CD changes |

### Examples

```
feat(dev): add hot reload for template changes
fix(deploy): handle missing aerostack.toml gracefully
docs: update README quick start
chore(deps): bump cobra to v1.8.0
```

---

## Changelog

- Maintain `CHANGELOG.md` in [Keep a Changelog](https://keepachangelog.com/) format
- Update before each release
- GoReleaser can auto-generate from commits (see `.goreleaser.yaml`)

---

## Rollback

If a release has critical issues:

1. **Yank** the release (GitHub: mark as pre-release or delete)
2. Create a **patch** release (e.g. v1.0.1) with the fix
3. Document in CHANGELOG under "Security" or "Fixed"

---

## References

- [GoReleaser](https://goreleaser.com/)
- [Semantic Versioning](https://semver.org/)
- [Conventional Commits](https://www.conventionalcommits.org/)
