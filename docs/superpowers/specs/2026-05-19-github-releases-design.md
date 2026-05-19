# GitHub Automated Releases Design

**Status:** Approved  
**Date:** 2026-05-19

## Goal

Automatically build Go binaries for x86, x64, arm, and arm64 on every `v*` tag push and publish them as a GitHub Release.

## Architecture

Single GitHub Actions workflow triggered by tag pushes matching `v*`. Uses Go's built-in cross-compilation (`GOOS`/`GOARCH`/`GOARM`) with `CGO_ENABLED=0` for fully static binaries. Archives are uploaded as release assets via `softprops/action-gh-release`.

## Build Matrix

| Target | GOARCH | GOARM | Notes |
|--------|--------|-------|-------|
| x86 | 386 | - | 32-bit Intel |
| x64 | amd64 | - | 64-bit Intel/AMD |
| arm | arm | 6 | Raspberry Pi Zero/1 |
| arm64 | arm64 | - | Raspberry Pi 3+/4/5 |

## Asset Naming

`tcpserial_<VERSION>_linux_<ARCH>.tar.gz` — each archive contains a single static binary.

## Version Injection

- `main.version` = tag name with `v` prefix stripped (e.g., tag `v0.1.0` → binary reports `0.1.0`)
- `main.commit` = full commit SHA
- `main.date` = ISO8601 build timestamp

## Workflow

```
User pushes tag v0.1.0 → GitHub Actions triggers
  → 4 parallel builds (x86, x64, arm, arm64)
  → Each produces a tar.gz archive
  → softprops/action-gh-release creates Release with all 4 assets
  → Auto-generates release notes from commits since last tag
```

## Files

- **Create:** `.github/workflows/release.yml` (single workflow file)

## Edge Cases

- **Non-tag pushes:** Workflow only triggers on `v*` tags — non-tag pushes are ignored
- **Tag without `v` prefix:** Tag `0.1.0` won't trigger (intentional — prevents accidental releases)
- **Empty release notes:** If no commits since last tag, release notes will just show the tag name
- **Build failure in one architecture:** Each matrix job is independent — arm64 can succeed while arm fails, and the release will still include the successful builds
- **Re-tagging the same version:** Pushing a tag that already exists will fail at the git level before reaching Actions
