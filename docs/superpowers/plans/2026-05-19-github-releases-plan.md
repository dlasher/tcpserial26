# GitHub Automated Releases Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build and publish tcpserial binaries for x86, x64, arm, arm64 on every `v*` tag push.

**Architecture:** Single GitHub Actions workflow with a build matrix for 4 architectures. Go cross-compilation with `CGO_ENABLED=0`. Releases via `softprops/action-gh-release`.

**Tech Stack:** GitHub Actions, Go 1.23, softprops/action-gh-release

---

### Task 1: Create release workflow

**Files:**
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: Create the workflow file**

Write `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    name: Build ${{ matrix.arch }}
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux]
        goarch: [386, amd64, arm, arm64]
        include:
          - goarch: arm
            goarm: "6"

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          check-latest: true

      - name: Build
        run: |
          mkdir -p dist
          GOOS=${{ matrix.goos }} GOARCH=${{ matrix.goarch }} GOARM=${{ matrix.goarm }} \
          CGO_ENABLED=0 \
          go build -ldflags "-X main.version=${GITHUB_REF_NAME#v} -X main.commit=${{ github.sha }} -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
          -o dist/tcpserial_${GITHUB_REF_NAME}_${{ matrix.goos }}_${{ matrix.goarch }} ./cmd/tcpserial

      - name: Archive
        run: |
          cd dist
          tar czf tcpserial_${GITHUB_REF_NAME}_${{ matrix.goos }}_${{ matrix.goarch }}.tar.gz \
            tcpserial_${GITHUB_REF_NAME}_${{ matrix.goos }}_${{ matrix.goarch }}

      - uses: softprops/action-gh-release@v2
        if: startsWith(github.ref, 'refs/tags/')
        with:
          files: dist/*.tar.gz
          generate_release_notes: true
```

- [ ] **Step 2: Create the directory and save the file**

```bash
mkdir -p .github/workflows
```

Write `.github/workflows/release.yml` with the content above.

- [ ] **Step 3: Verify the file exists**

```bash
ls -la .github/workflows/release.yml
```

Expected output: `.github/workflows/release.yml` exists and is readable.

- [ ] **Step 4: Stage and commit**

```bash
git add .github/workflows/release.yml docs/superpowers/specs/2026-05-19-github-releases-design.md docs/superpowers/plans/2026-05-19-github-releases-plan.md
git commit -m "ci: add automated release workflow for x86/x64/arm/arm64"
```

- [ ] **Step 5: Push to origin**

```bash
git push origin main
```

- [ ] **Step 6: Verify the workflow is visible on GitHub**

Open `https://github.com/dlasher/tcpserial26/actions` in a browser. The "Release" workflow should appear in the left sidebar.

- [ ] **Step 7: Create a test tag to trigger a release**

```bash
git tag v0.1.0
git push origin v0.1.0
```

Expected: The Release workflow triggers on GitHub Actions. Check:
- `https://github.com/dlasher/tcpserial26/actions` — build running
- `https://github.com/dlasher/tcpserial26/releases` — release appears after completion
