# PITFALLS: kubectl Plugin Krew Distribution

Common mistakes when preparing kubectl plugins for krew distribution, based on krew index requirements and patterns from successful plugins.

---

## 1. KREW MANIFEST STRUCTURE

### Pitfall: Invalid or Incomplete Plugin Manifest
**What goes wrong:**
- Missing required fields (shortDescription, homepage, platforms)
- Invalid platform/architecture combinations (e.g., only linux/amd64)
- Incorrect binary path or selector in manifest
- SHA256 checksums don't match released artifacts
- Version format doesn't follow semver (must be v1.2.3)

**Warning signs:**
- `krew-index` CI validation fails on PR
- Plugin installs but binary not found in PATH
- Installation fails on specific platforms
- Users report "checksum mismatch" errors

**Prevention strategy:**
- Use `krew-release-bot` or validated template for manifest generation
- Include ALL mainstream platforms: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- Test manifest locally with `kubectl krew install --manifest=<file> --archive=<local-archive>`
- Validate manifest structure against krew schema before submission
- Automate SHA256 generation in release pipeline (GoReleaser does this)

**Phase: CI/CD Setup & Manifest Creation**

---

## 2. BINARY NAMING CONVENTION

### Pitfall: Incorrect Binary Names or Structure
**What goes wrong:**
- Binary not named exactly `kubectl-<plugin>` (breaks kubectl plugin discovery)
- Different binary name in archives vs manifest selector
- Binary at wrong path in archive (not at root or specified bin path)
- Executable permissions not set in archive

**Warning signs:**
- `kubectl <plugin>` fails with "unknown command"
- Krew installs but binary not in `~/.krew/bin/`
- Works locally but fails after krew installation

**Prevention strategy:**
- Binary MUST be named `kubectl-analyze-images` (matches your project)
- Ensure GoReleaser `binary:` field matches exactly
- Archives must contain binary at root or manifest `bin` field must specify path
- Test: `tar tzf <archive> | grep kubectl-analyze-images` should show the binary
- Set file mode to 0755 in GoReleaser archive config

**Phase: Build Configuration (GoReleaser)**

**Example from manifest:**
```yaml
spec:
  platforms:
  - bin: kubectl-analyze-images  # Must match binary name
    files:
    - from: kubectl-analyze-images  # From archive root
      to: .
```

---

## 3. GORELEASER CONFIGURATION

### Pitfall: Missing or Incomplete Cross-Platform Builds
**What goes wrong:**
- Only building for developer's platform (darwin/arm64)
- Missing windows/.exe suffix handling
- Not using CGO_ENABLED=0 (creates dynamic linking issues)
- Archive format wrong (tar.gz for unix, zip for windows)
- Missing checksums file generation

**Warning signs:**
- Krew manifest submission rejected for missing platforms
- Users on Windows/Linux/ARM report installation failures
- "shared library not found" errors on some platforms
- Release assets don't include checksums

**Prevention strategy:**
- Configure builds for: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, windows/amd64
- Use GoReleaser's platform-specific archive formats:
  ```yaml
  archives:
    - format_overrides:
        - goos: windows
          format: zip
  ```
- Set environment: `CGO_ENABLED=0` for static binaries
- Enable checksum generation: `checksum: { name_template: 'checksums.txt' }`
- Test on non-development platforms (use GitHub Actions matrix)

**Phase: Build Configuration (GoReleaser)**

---

## 4. GITHUB ACTIONS CI/CD

### Pitfall: Release Workflow Failures
**What goes wrong:**
- GoReleaser runs without Git tag (version detection fails)
- GITHUB_TOKEN lacks release permissions
- Workflow triggers on every push (should be tags only)
- Building before running tests (ship broken code)
- Not caching Go modules (slow CI)

**Warning signs:**
- Release fails with "no tag found"
- Assets not uploaded to GitHub release
- Every commit triggers release workflow
- CI takes 10+ minutes for simple build

**Prevention strategy:**
- Trigger only on tags: `on: push: tags: ['v*']`
- Run tests BEFORE GoReleaser in workflow
- Use actions/checkout with `fetch-depth: 0` for full git history
- Cache Go modules: `actions/setup-go` with `cache: true`
- Grant workflow write permissions:
  ```yaml
  permissions:
    contents: write
  ```
- Validate workflow with `act` locally or use branch protection

**Phase: CI/CD Setup**

**Example workflow structure:**
```yaml
on:
  push:
    tags: ['v*']
jobs:
  test:
    runs-on: ubuntu-latest
    steps: [checkout, setup-go, test]
  release:
    needs: test
    runs-on: ubuntu-latest
    steps: [checkout, setup-go, goreleaser]
```

---

## 5. GO MODULE AND DEPENDENCIES

### Pitfall: Kubernetes Client Version Mismatches
**What goes wrong:**
- Using outdated client-go (you have v0.29.0, latest is v0.32+)
- Mixing incompatible k8s.io/* module versions
- Direct dependencies on internal k8s packages
- Large binary size from unnecessary dependencies

**Warning signs:**
- "unsupported API version" errors with newer clusters
- Binary size >50MB (Go binaries should be 10-20MB)
- Dependency conflicts during `go mod tidy`
- Deprecated API warnings in logs

**Prevention strategy:**
- Align ALL k8s.io/* packages to same version (client-go, api, apimachinery)
- Use client-go's recommended version matrix for kubectl compatibility
- Review dependencies: `go mod graph | grep k8s.io`
- Consider `go mod vendor` if using replace directives
- Test against multiple Kubernetes versions (1.28, 1.29, 1.30+)

**Phase: Dependencies Audit**

**Current issue:**
```
Your go.mod shows client-go v0.29.0 (Jan 2024)
Latest stable: v0.32.x (supports K8s 1.32)
Consider updating unless you need K8s 1.28 compatibility
```

---

## 6. PLUGIN NAMING AND DISCOVERY

### Pitfall: Plugin Name Conflicts or Poor Discoverability
**What goes wrong:**
- Plugin name too generic (conflicts with existing plugins)
- Name doesn't reflect functionality
- Poor shortDescription in manifest (users can't find it)
- Missing or unclear homepage/repository links

**Warning signs:**
- Krew index rejects due to name conflict
- Plugin hard to find with `kubectl krew search`
- Users don't understand what plugin does from listing
- No documentation link in krew output

**Prevention strategy:**
- Check existing plugins: `kubectl krew search <keyword>`
- Verify name available: search krew-index for `name: <yourname>`
- ShortDescription must be <80 chars, action-focused
- Homepage must link to GitHub repo with README
- Consider including "kubectl" in GitHub repo description for SEO

**Phase: Manifest Creation**

**Your plugin:**
```yaml
# Good: specific, clear purpose
name: analyze-images
shortDescription: Analyze container images used across cluster resources
description: |
  Scans cluster pods, deployments, statefulsets to report image usage,
  versions, registries, and identify outdated or vulnerable images.
homepage: https://github.com/yourusername/kubectl-analyze-images
```

---

## 7. RELEASE PROCESS AND VERSIONING

### Pitfall: Manual or Inconsistent Releases
**What goes wrong:**
- Forgetting to create Git tag before release
- Tag format inconsistent (v1.0.0 vs 1.0.0 vs v1.0)
- Manually editing release assets (breaks automation)
- No changelog or release notes
- Krew manifest update forgotten after release

**Warning signs:**
- Release exists but no corresponding tag
- Manifest version out of sync with latest release
- Users report using old version after update
- No clear upgrade path documented

**Prevention strategy:**
- ALWAYS use semantic versioning with 'v' prefix: v1.0.0, v1.1.0, v2.0.0
- Automate: `git tag v1.0.0 && git push origin v1.0.0` triggers release
- Use GoReleaser's changelog generation
- Create krew manifest PR immediately after release
- Document release process in CONTRIBUTING.md
- Consider release-please or semantic-release for automation

**Phase: Release Process Documentation**

**Release checklist:**
1. Update CHANGELOG.md
2. Create tag: `git tag -a v1.0.0 -m "Release v1.0.0"`
3. Push tag: `git push origin v1.0.0`
4. Verify GitHub Actions release succeeds
5. Update krew-index PR with new manifest
6. Test installation: `kubectl krew install <plugin>`

---

## 8. TESTING AND VALIDATION

### Pitfall: No Pre-Release Testing
**What goes wrong:**
- Shipping untested cross-platform builds
- Not testing actual krew installation flow
- No integration tests with real Kubernetes cluster
- Breaking changes not caught before release

**Warning signs:**
- Users report immediate crashes on specific platforms
- Plugin works in development but fails after krew install
- Different behavior between `go run` and installed binary
- Permissions issues on Windows/macOS

**Prevention strategy:**
- Unit tests for core logic (you have zero tests currently)
- Integration tests against kind/minikube cluster
- Test installation from krew manifest locally:
  ```bash
  kubectl krew install --manifest=kubectl-analyze-images.yaml \
    --archive=dist/kubectl-analyze-images-linux-amd64.tar.gz
  ```
- Matrix testing in CI across platforms
- Manual testing on each platform before manifest submission

**Phase: Testing Infrastructure Setup**

**Minimum test coverage needed:**
- Parse flags and arguments correctly
- Handle kubeconfig and context switching
- Error handling for missing cluster access
- Image data extraction from pod specs

---

## 9. DOCUMENTATION AND USER EXPERIENCE

### Pitfall: Poor or Missing Documentation
**What goes wrong:**
- README doesn't explain krew installation method
- No usage examples for common scenarios
- Missing flag documentation
- No troubleshooting guide
- License file missing (krew requirement)

**Warning signs:**
- High volume of "how do I..." issues
- Krew manifest rejected for missing LICENSE
- Users can't figure out basic usage
- No clear installation instructions

**Prevention strategy:**
- README must include:
  - Krew installation: `kubectl krew install analyze-images`
  - Manual installation as fallback
  - Clear usage examples
  - All flags documented
  - Output format examples
- Add LICENSE file (Apache-2.0 or MIT recommended)
- Include troubleshooting section
- Link to krew plugin guidelines

**Phase: Documentation**

**Required sections:**
```markdown
## Installation
### Via Krew (recommended)
kubectl krew install analyze-images

### Manual Installation
[Download from releases]

## Usage
kubectl analyze-images [flags]

## Examples
[3-5 common scenarios]

## Troubleshooting
[Common issues and solutions]
```

---

## 10. SECURITY AND BEST PRACTICES

### Pitfall: Security and Maintenance Issues
**What goes wrong:**
- Hardcoded credentials or secrets in code
- No security scanning in CI
- Vulnerable dependencies not updated
- No SECURITY.md or vulnerability reporting process
- Excessive permissions requested by plugin

**Warning signs:**
- Dependabot/Snyk alerts on dependencies
- Plugin requests cluster-admin when read-only sufficient
- Credentials accidentally committed
- No regular dependency updates

**Prevention strategy:**
- Use kubeconfig credentials, never hardcode
- Enable GitHub Dependabot for go.mod updates
- Add security scanning to CI (gosec, trivy)
- Follow principle of least privilege for RBAC
- Create SECURITY.md with vulnerability reporting process
- Regular dependency audits: `go list -m all | nancy sleuth`

**Phase: Security Setup**

**Minimum security checklist:**
- [ ] No secrets in code or history
- [ ] Dependabot enabled
- [ ] Security scanning in CI
- [ ] SECURITY.md present
- [ ] Dependencies up to date

---

## PHASE MAPPING SUMMARY

**Phase 1: Dependencies Audit**
- Update client-go and k8s.io/* packages (Pitfall 5)
- Align module versions
- Run `go mod tidy`

**Phase 2: Build Configuration (GoReleaser)**
- Set up cross-platform builds (Pitfall 3)
- Configure binary naming correctly (Pitfall 2)
- Set CGO_ENABLED=0
- Configure archives and checksums

**Phase 3: CI/CD Setup**
- Create GitHub Actions workflow (Pitfall 4)
- Run tests before build
- Trigger on tags only
- Grant release permissions

**Phase 4: Testing Infrastructure**
- Add unit tests (Pitfall 8)
- Set up integration tests
- Test krew installation locally

**Phase 5: Manifest Creation**
- Generate krew manifest (Pitfall 1)
- Validate naming and description (Pitfall 6)
- Test manifest with local archives

**Phase 6: Documentation**
- Update README with krew installation (Pitfall 9)
- Add LICENSE file
- Create SECURITY.md (Pitfall 10)
- Document release process (Pitfall 7)

**Phase 7: Security Setup**
- Enable Dependabot
- Add security scanning to CI
- Review permissions and credentials

**Phase 8: First Release**
- Create v1.0.0 tag
- Verify GitHub release
- Submit krew-index PR
- Test end-to-end installation

---

## CRITICAL PATH ISSUES

**These must be fixed BEFORE krew submission:**

1. **Missing LICENSE file** - krew requires open source license
2. **No cross-platform builds** - must support linux/darwin/windows on amd64/arm64
3. **No tests** - risk shipping broken code
4. **No goreleaser config** - manual releases won't scale
5. **No krew manifest** - can't submit without it

**Order of operations:**
1. Add LICENSE (5 minutes)
2. Set up GoReleaser (1 hour)
3. Configure GitHub Actions (1 hour)
4. Write basic tests (2-4 hours)
5. Generate and test krew manifest (1 hour)
6. Create first release (30 minutes)
7. Submit to krew-index (30 minutes + review wait time)

---

## REFERENCES

Key resources to prevent these pitfalls:

- Krew Developer Guide: https://krew.sigs.k8s.io/docs/developer-guide/
- Krew Plugin Manifest: https://krew.sigs.k8s.io/docs/developer-guide/plugin-manifest/
- GoReleaser kubectl plugin example: https://goreleaser.com/cookbooks/kubectl-plugin/
- Example plugins to study:
  - kubectl-neat (clean manifest output)
  - kubectl-tree (resource relationships)
  - kubectl-images (similar to your plugin)
- Krew index validation: https://github.com/kubernetes-sigs/krew-index/blob/master/.github/workflows/plugin-validation.yaml

---

## DETECTION CHECKLIST

Use this before each phase to catch pitfalls early:

**Pre-Build:**
- [ ] Binary name matches `kubectl-<plugin>` exactly
- [ ] Go modules aligned to same k8s.io version
- [ ] LICENSE file exists
- [ ] README has installation instructions

**Pre-Release:**
- [ ] Tests pass on CI
- [ ] GoReleaser builds for all platforms
- [ ] Checksums file generated
- [ ] GitHub Actions triggers only on tags
- [ ] Tag follows v1.2.3 format

**Pre-Krew Submission:**
- [ ] Manifest passes `krew validate-plugin`
- [ ] Tested installation on Linux/macOS/Windows
- [ ] ShortDescription under 80 chars
- [ ] Homepage links to repo with docs
- [ ] All platform SHA256 checksums match

**Post-Release:**
- [ ] GitHub release has all platform archives
- [ ] Krew manifest PR submitted
- [ ] Installation tested: `kubectl krew install <plugin>`
- [ ] Plugin works: `kubectl <plugin> --help`
