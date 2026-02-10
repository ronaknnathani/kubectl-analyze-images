# Roadmap

## Milestone: v1.0.0 — Krew-Ready Distribution

### Phase 1: Foundation & Testing Infrastructure
**Goal:** Establish test framework and deduplicate utilities to enable safe refactoring.
**Plans:** 2 plans

Plans:
- [x] 01-01-PLAN.md -- Shared utility package, deduplication, config cleanup, and utility tests
- [x] 01-02-PLAN.md -- Printer interface abstraction, reporter refactor, and printer tests

**Scope:**
- Add Go testing framework with testify dependency
- Extract and deduplicate image parsing logic into shared utility package
- Create output formatting abstraction with Printer interface
- Write unit tests for image parsing and output formatters
- Remove unused configuration fields (concurrency, caching, retry)

**Success criteria:**
- `go test ./...` passes with baseline test coverage (>30%)
- Image parsing logic exists in exactly one location
- Output formatters (table/JSON) have unit test coverage
- Build completes successfully with `make build`
- No code duplication for registry/tag extraction

**Depends on:** None
**Research needed:** no

---

### Phase 2: Kubernetes Abstraction Layer
**Goal:** Enable testable cluster interactions through interface-driven design.
**Plans:** 2 plans

Plans:
- [x] 02-01-PLAN.md -- Kubernetes interface definition with real and fake client implementations
- [x] 02-02-PLAN.md -- Refactor cluster/analyzer for dependency injection and add unit tests

**Scope:**
- Create `pkg/kubernetes/interface.go` defining ListPods(), ListNodes() methods
- Implement real client wrapper in `pkg/kubernetes/client.go`
- Create FakeClient implementation for testing
- Refactor `internal/cluster/` to use kubernetes.Interface
- Add unit tests using FakeClient for all cluster operations

**Success criteria:**
- All Kubernetes API interactions go through interface abstraction
- FakeClient enables testing without real cluster
- Existing functionality unchanged (backward compatibility verified)
- Test coverage for cluster operations >60%
- `make build` passes, manual testing shows no regressions

**Depends on:** Phase 1
**Research needed:** no

---

### Phase 3: Plugin Restructuring
**Goal:** Refactor business logic into clean, testable plugin architecture following kubectl patterns.
**Plans:** 2 plans

Plans:
- [x] 03-01-PLAN.md -- Create plugin options with Complete/Validate/Run pattern, refactor main.go
- [x] 03-02-PLAN.md -- Add comprehensive plugin tests, make Reporter output testable, achieve >70% coverage

**Scope:**
- Create `pkg/plugin/options.go` with Complete/Validate/Run pattern
- Move analyzer logic from `internal/analyzer/` to `pkg/plugin/`
- Inject kubernetes.Interface and Printer dependencies
- Refactor `main.go` to minimal command setup (<100 lines)
- Add comprehensive plugin tests with mocked dependencies

**Success criteria:**
- Plugin follows Complete/Validate/Run kubectl standard pattern
- Main.go is thin CLI layer with no business logic
- All dependencies injected (no globals or singletons)
- Test coverage >70% across all packages
- All existing features functional (table output, JSON, histograms, filtering)

**Depends on:** Phase 2
**Research needed:** no

---

### Phase 4: Build & Release Automation
**Goal:** Automate multi-platform builds and establish professional CI/CD pipeline.
**Plans:** 2 plans

Plans:
- [x] 04-01-PLAN.md -- GoReleaser config, golangci-lint config, and Makefile updates
- [x] 04-02-PLAN.md -- GitHub Actions CI and Release workflows

**Scope:**
- Create `.goreleaser.yaml` with multi-platform configuration (linux/darwin/windows, amd64/arm64)
- Add `.github/workflows/ci.yml` for test and lint on pull requests
- Add `.github/workflows/release.yml` for automated releases on tags
- Configure golangci-lint with `.golangci.yml`
- Test local release with `goreleaser --snapshot`

**Success criteria:**
- GoReleaser produces binaries for all target platforms
- GitHub Actions CI runs tests on every PR
- Release workflow creates GitHub release with artifacts on version tags
- golangci-lint passes with zero errors
- Checksums generated automatically for all binaries

**Depends on:** Phase 3
**Research needed:** no

---

### Phase 5: Krew Distribution
**Goal:** Package plugin for krew distribution with complete documentation and first public release.
**Plans:** 2 plans

Plans:
- [x] 05-01-PLAN.md -- LICENSE, krew manifest, SECURITY.md, CONTRIBUTING.md
- [x] 05-02-PLAN.md -- README polish with krew instructions and usage examples

**Scope:**
- Generate and validate krew plugin manifest
- Add LICENSE file (Apache-2.0)
- Polish README with krew installation instructions and usage examples
- Create SECURITY.md and CONTRIBUTING.md
- Tag v1.0.0 and verify release pipeline
- Test local installation with `kubectl krew install --manifest`
- Submit PR to kubernetes-sigs/krew-index

**Success criteria:**
- Krew manifest validates with `kubectl krew install --manifest=file`
- LICENSE file present (Apache-2.0)
- README documents installation via krew
- v1.0.0 release created with all platform binaries
- Local krew installation works end-to-end
- Plugin discoverable as `kubectl analyze-images`

**Depends on:** Phase 4
**Research needed:** no

---

## Progress Tracking

| Phase | Status | Completion |
|-------|--------|------------|
| Phase 1: Foundation & Testing Infrastructure | ✓ Complete (2026-02-09) | 100% |
| Phase 2: Kubernetes Abstraction Layer | ✓ Complete (2026-02-10) | 100% |
| Phase 3: Plugin Restructuring | ✓ Complete (2026-02-10) | 100% |
| Phase 4: Build & Release Automation | ✓ Complete (2026-02-10) | 100% |
| Phase 5: Krew Distribution | ✓ Complete (2026-02-10) | 100% |

**Overall Milestone Progress:** 5/5 phases complete (100%)

---

*Roadmap created: 2026-02-09*
*Target completion: v1.0.0 release ready for krew submission*
