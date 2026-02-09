# Features Research: kubectl Plugin Quality & Distribution

**Research Date**: 2026-02-09
**Context**: Cleaning up kubectl-analyze-images Go plugin for krew distribution
**References**: krew.sigs.k8s.io, ahmetb/kubectx, ahmetb/kubectl-tree, kubernetes-sigs/krew

---

## Executive Summary

For krew distribution, kubectl plugins need three layers:
1. **Table Stakes**: Minimum viable for krew acceptance (manifest, docs, releases)
2. **Differentiators**: Quality signals that separate good plugins from basic ones (tests, CI, automation)
3. **Anti-features**: Things to deliberately avoid in v1 (over-engineering, premature optimization)

---

## Table Stakes Features
*Required for krew listing - without these, plugin won't be accepted*

### 1. Krew Plugin Manifest
**Complexity**: Low
**Effort**: 1-2 hours
**Dependencies**: None

**Requirements**:
- `.krew.yaml` file in root or separate repo
- Required fields:
  - `apiVersion: krew.googlecontainertools.github.com/v1alpha2`
  - `kind: Plugin`
  - `metadata.name` (must match binary name without kubectl- prefix)
  - `spec.version` (semver format)
  - `spec.shortDescription` (under 50 chars)
  - `spec.description` (detailed, markdown supported)
  - `spec.platforms[]` with OS/arch combinations
  - `spec.platforms[].uri` (download URL for each platform)
  - `spec.platforms[].sha256` (checksum for each platform)
  - `spec.platforms[].bin` (binary name in archive)

**Best Practices**:
- Support minimum 3 platforms: linux/amd64, darwin/amd64, windows/amd64
- Add linux/arm64 and darwin/arm64 for modern hardware
- Use GitHub releases as download source
- Single binary per platform (no dependencies)

### 2. Semantic Versioning + GitHub Releases
**Complexity**: Low
**Effort**: 2-3 hours
**Dependencies**: None

**Requirements**:
- Use semver (v1.0.0, v1.1.0, etc.)
- Git tags for each release
- GitHub release with:
  - Release notes (changelog)
  - Compiled binaries for all supported platforms
  - Checksums file (SHA256SUMS)

**Best Practices**:
- Start at v1.0.0 for krew submission
- Keep CHANGELOG.md updated
- Include upgrade notes for breaking changes

### 3. Basic Documentation
**Complexity**: Low
**Effort**: 2-4 hours
**Dependencies**: None

**Requirements**:
- README.md with:
  - Clear description of what plugin does
  - Installation instructions (via krew)
  - Basic usage examples
  - License information
- LICENSE file (Apache-2.0 or MIT preferred)

**Best Practices**:
- Show example output with screenshots/asciinema
- Document all flags and subcommands
- Include troubleshooting section
- Add badges (build status, license, version)

### 4. Cobra CLI Structure
**Complexity**: Low (already done)
**Effort**: N/A
**Dependencies**: cobra, client-go

**Requirements**:
- Root command with --help
- Proper flag handling
- Error messages to stderr
- Exit codes (0 success, 1 error)

**Best Practices**:
- Add --version flag
- Support -v/--verbose for debug output
- Use consistent flag naming (-o/--output, --namespace, etc.)
- Respect KUBECONFIG environment variable

---

## Differentiator Features
*Quality signals that set professional plugins apart*

### 5. Unit & Integration Tests
**Complexity**: Medium
**Effort**: 8-12 hours
**Dependencies**: testing, testify, fake client-go

**Requirements**:
- Unit tests for core logic (>70% coverage)
- Table-driven tests for multiple scenarios
- Mock Kubernetes API responses

**Best Practices**:
- Use client-go's fake clientset
- Test error paths (API failures, timeouts)
- Separate business logic from K8s client code
- Add integration tests with kind/minikube (optional)

**Example Structure**:
```
internal/analyzer/pod_analyzer_test.go
internal/cluster/client_test.go
internal/reporter/report_test.go
```

### 6. CI/CD Pipeline (GitHub Actions)
**Complexity**: Medium
**Effort**: 4-6 hours
**Dependencies**: GitHub Actions

**Requirements**:
- Automated testing on PR
- Build verification for all platforms
- Linting (golangci-lint)

**Best Practices**:
- Test on multiple Go versions (1.22, 1.23)
- Run on PR and main branch
- Cache Go modules
- Fail fast on errors

**Quality Gates**:
- `go test ./...` must pass
- `go vet ./...` must pass
- `golangci-lint run` must pass
- Cross-compilation check: `GOOS=linux go build`, `GOOS=darwin go build`, `GOOS=windows go build`

### 7. Automated Release Pipeline
**Complexity**: Medium
**Effort**: 3-5 hours
**Dependencies**: goreleaser or custom scripts

**Requirements**:
- Automated binary builds on git tag
- Cross-platform compilation
- SHA256 checksum generation
- GitHub release creation

**Best Practices**:
- Use GoReleaser (industry standard)
- Generate .krew.yaml automatically
- Create archives (tar.gz for Unix, zip for Windows)
- Sign binaries (optional for v1)

**GoReleaser Config**:
```yaml
# .goreleaser.yml
builds:
  - env: [CGO_ENABLED=0]
    goos: [linux, darwin, windows]
    goarch: [amd64, arm64]
archives:
  - format: tar.gz
    format_overrides:
      - goos: windows
        format: zip
```

### 8. Code Quality Tools
**Complexity**: Low
**Effort**: 2-3 hours
**Dependencies**: golangci-lint, gofmt

**Requirements**:
- Consistent formatting (gofmt)
- Linting rules (golangci-lint)
- No obvious bugs (go vet)

**Best Practices**:
- `.golangci.yml` config with reasonable rules
- Pre-commit hooks (optional)
- Code coverage reporting (codecov.io)

**Recommended Linters**:
- errcheck (unchecked errors)
- govet (suspicious constructs)
- staticcheck (static analysis)
- unused (unused code)
- gosimple (simplifications)

### 9. Output Formatting Options
**Complexity**: Low-Medium
**Effort**: 3-4 hours
**Dependencies**: Already partially implemented

**Requirements**:
- Multiple output formats (-o json, -o yaml, -o table)
- Structured output for automation
- Human-readable default

**Best Practices**:
- Default to table format
- Support json/yaml for scripting
- Add wide output (-o wide) for additional columns
- Consistent field naming across formats

### 10. Error Handling & User Experience
**Complexity**: Medium
**Effort**: 4-6 hours
**Dependencies**: None

**Requirements**:
- Clear error messages
- Helpful suggestions on common errors
- Graceful degradation

**Best Practices**:
- Detect missing kubeconfig and suggest fix
- Handle API timeout with retry hints
- Validate flags before API calls
- Show progress for slow operations

---

## Anti-Features for v1
*Things to deliberately NOT build - scope creep risks*

### 11. Multi-Cluster Support
**Rationale**: kubectl already handles context switching. Plugin should respect current context.
**Defer to**: v2 if users request it

### 12. Custom Resource Definitions (CRDs)
**Rationale**: Read-only analysis tool shouldn't require cluster-level resources.
**Defer to**: Never - out of scope for this plugin type

### 13. Complex Configuration Files
**Rationale**: Flags are sufficient for simple tool. Config files add maintenance burden.
**Defer to**: v2+ if flag count exceeds 10-15

### 14. Plugin Auto-Update Mechanism
**Rationale**: Krew handles updates. Don't reinvent the wheel.
**Defer to**: Never - krew's responsibility

### 15. Telemetry/Analytics
**Rationale**: Privacy concerns, adds complexity, not needed for personal project.
**Defer to**: Never for personal plugins

### 16. Web UI / Dashboard
**Rationale**: CLI tool should stay CLI. Separate project if visualization needed.
**Defer to**: Separate project if desired

### 17. Write Operations
**Rationale**: Analysis tool should be read-only for safety. No patching/deleting.
**Defer to**: Never - changes the tool's purpose

---

## Feature Dependencies & Sequencing

### Critical Path (Must be sequential):
1. Krew manifest → Requires releases
2. Releases → Requires cross-compilation
3. Cross-compilation → Requires tests passing
4. Tests → Requires refactored code

### Parallel Work Streams:
- **Stream 1**: Testing + CI/CD (can work together)
- **Stream 2**: Documentation + Examples (independent)
- **Stream 3**: Release automation + Manifest (depends on Stream 1)

### Recommended Implementation Order:
1. **Week 1**: Unit tests + refactor for testability (Feature 5)
2. **Week 2**: CI/CD pipeline + linting (Features 6, 8)
3. **Week 3**: Release automation + checksums (Feature 7)
4. **Week 4**: Documentation polish + krew manifest (Features 3, 1)
5. **Week 5**: Submit to krew + monitor feedback

---

## Complexity Matrix

| Feature | Complexity | Effort (hrs) | Priority | Blockers |
|---------|-----------|--------------|----------|----------|
| Krew Manifest | Low | 1-2 | P0 | Releases |
| Releases + Versioning | Low | 2-3 | P0 | None |
| Basic Documentation | Low | 2-4 | P0 | None |
| Cobra CLI | Low | 0 | P0 | Done |
| Unit Tests | Medium | 8-12 | P1 | None |
| CI/CD Pipeline | Medium | 4-6 | P1 | Tests |
| Release Automation | Medium | 3-5 | P1 | Tests, CI |
| Code Quality Tools | Low | 2-3 | P1 | None |
| Output Formatting | Low-Med | 3-4 | P2 | Partial |
| Error Handling | Medium | 4-6 | P2 | None |

**Total Effort Estimate**: 30-45 hours for P0-P1 features

---

## Reference: ahmetb Plugin Analysis

### kubectx/kubens
**Standout Features**:
- Shell script → Go rewrite for portability
- Bash/zsh completion files
- Comprehensive test suite
- Multi-platform releases via GoReleaser
- Active maintenance (issues addressed quickly)

**Distribution Quality**:
- 15k+ stars, top krew plugin
- Clear README with GIFs
- Installation via krew, brew, apt
- Works without kubectl (standalone tool)

### kubectl-tree
**Standout Features**:
- Focused single purpose (ownership tree)
- Colorized output with icons
- Good error messages
- Fast (efficient API queries)

**Distribution Quality**:
- 2.5k+ stars
- Clean codebase, easy to read
- Automated releases
- Responsive to bug reports

**Common Patterns**:
1. Single binary, no dependencies
2. GoReleaser for multi-platform builds
3. Semantic versioning
4. Apache-2.0 license
5. GitHub Actions for CI/CD
6. Unit tests with fake clientset
7. Examples in README
8. Active issue/PR management

---

## Krew Submission Checklist

Before submitting to krew-index:

- [ ] Plugin follows naming convention (kubectl-foo)
- [ ] Binary builds for linux/amd64, darwin/amd64, windows/amd64
- [ ] GitHub release with versioned binaries
- [ ] SHA256 checksums provided
- [ ] .krew.yaml manifest validated (`kubectl krew validate`)
- [ ] README documents installation via krew
- [ ] LICENSE file present
- [ ] Plugin tested on at least 2 platforms
- [ ] No hard dependencies (runs standalone)
- [ ] Respects --kubeconfig flag
- [ ] Help text and --version work

**Submission Process**:
1. Fork kubernetes-sigs/krew-index
2. Add plugin manifest to `plugins/`
3. Open PR with:
   - Plugin manifest
   - Test evidence (screenshots)
   - Link to plugin repo
4. Wait for automated checks
5. Address maintainer feedback
6. Merge and announce

---

## Recommendations for kubectl-analyze-images

### Must-Have (P0):
1. Write comprehensive unit tests (current: 0 tests)
2. Set up GitHub Actions CI/CD
3. Configure GoReleaser for multi-platform releases
4. Create v1.0.0 release with binaries
5. Generate krew manifest with correct checksums
6. Polish README with usage examples

### Should-Have (P1):
7. Add golangci-lint to CI
8. Improve error messages and UX
9. Add integration test with kind (optional)
10. Set up code coverage tracking

### Nice-to-Have (P2):
11. Add shell completion scripts
12. Create asciinema demo
13. Add --watch mode for continuous monitoring (defer to v1.1)
14. Support custom sorting/filtering (defer to v1.1)

### Avoid:
- Configuration files (use flags)
- Write operations (stay read-only)
- Complex dependencies (keep it lean)
- Auto-update mechanisms (let krew handle it)

---

## Success Metrics

**Krew Acceptance**:
- PR merged into krew-index within 2 weeks
- Automated checks pass first try
- Minimal maintainer feedback required

**Quality Indicators**:
- 80%+ test coverage
- All CI checks green
- No golangci-lint warnings
- Releases build successfully for all platforms

**User Adoption** (post-launch):
- 10+ GitHub stars in first month
- 50+ krew installs in first quarter
- No critical bugs reported in first month
- Positive feedback in krew index

---

## External Resources

### Essential Reading:
- https://krew.sigs.k8s.io/docs/developer-guide/
- https://github.com/kubernetes-sigs/krew/blob/master/docs/DEVELOPER_GUIDE.md
- https://goreleaser.com/quick-start/
- https://github.com/ahmetb/kubectx (reference implementation)

### Tools:
- GoReleaser: https://goreleaser.com/
- golangci-lint: https://golangci-lint.run/
- client-go testing: https://github.com/kubernetes/client-go/tree/master/testing

### Community:
- #krew channel in Kubernetes Slack
- kubernetes-sigs/krew GitHub discussions
- kubectl plugin developer guide

---

**Next Steps**: Use this research to define requirements in REQUIREMENTS.md, then create implementation plan in PLAN.md.
