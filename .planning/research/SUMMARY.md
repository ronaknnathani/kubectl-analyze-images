# Project Research Summary

**Project:** kubectl-analyze-images
**Domain:** Kubectl plugin distribution via krew
**Researched:** 2026-02-09
**Confidence:** HIGH

## Executive Summary

kubectl-analyze-images is a functional kubectl plugin that needs professional packaging for krew distribution. The research reveals a clear path: this is a standard Go kubectl plugin cleanup and distribution project with well-established patterns from the kubernetes-sigs/krew ecosystem. The recommended approach is GoReleaser + GitHub Actions for automated multi-platform releases, combined with architectural refactoring for testability (interface-driven design, Complete/Validate/Run pattern).

The research identified strong consensus across all dimensions. For technology stack, GoReleaser v2 with GitHub Actions is the industry standard for krew plugins (high confidence from ahmetb's kubectl-tree, kubectx references). For architecture, the golang-standards/project-layout with kubectl-specific patterns (pkg/ for public APIs, internal/ for private logic, interface abstraction for testing) is universally adopted. For features, the research clearly delineates table stakes (krew manifest, releases, tests) from differentiators (CI/CD automation, comprehensive testing) and anti-features (config files, write operations, telemetry).

The key risk is scope creep during code cleanup—there's a temptation to over-engineer. Mitigation: follow the phased approach strictly, validating after each phase. Start with low-risk utility extraction, then interface abstraction for testing, then plugin restructuring, and finally comprehensive tests. The critical path to krew submission is ~4 weeks with clear gates: tests passing, builds working, manifest validated, documentation complete.

## Key Findings

### Recommended Stack

The research converges on a mature, well-supported toolchain with native krew integration. GoReleaser v2 provides automated cross-platform builds, checksum generation, and built-in krew manifest generation. GitHub Actions offers free CI/CD for public repos with excellent Go ecosystem support. The testing stack (Go stdlib + testify + fake client-go) enables fast unit tests without cluster dependencies.

**Core technologies:**
- GoReleaser v2.4+: Multi-platform builds with native krew support — industry standard, proven with kubectl-tree
- GitHub Actions: CI/CD pipeline with matrix testing — free for OSS, native GitHub integration
- Go testing + testify: Unit test framework with table-driven tests — enables 80%+ coverage without cluster
- golangci-lint v1.61+: Static analysis and linting — catches bugs before CI
- client-go v0.29.x: Kubernetes API client aligned with K8s 1.29 — current project uses this, stable
- cobra v1.8.x: CLI framework — already in use, standard for kubectl plugins

### Expected Features

The research clearly separates must-have features (blocking krew acceptance) from nice-to-have quality improvements and explicitly identified anti-features to avoid.

**Must have (table stakes):**
- Krew plugin manifest with all required fields — without this, can't submit to krew-index
- Semantic versioning + GitHub releases with binaries — krew requirement for distribution
- Multi-platform support (linux/darwin/windows on amd64/arm64) — users expect cross-platform
- LICENSE file (Apache-2.0 or MIT) — krew-index submission requirement
- Basic README documentation with krew installation — users can't install otherwise
- Proper binary naming (kubectl-analyze_images) — breaks plugin discovery if wrong

**Should have (competitive):**
- Unit tests with 80%+ coverage — quality signal, enables confident refactoring
- GitHub Actions CI/CD with automated testing — professional projects have this
- GoReleaser automation for releases — manual releases don't scale, error-prone
- golangci-lint integration — catches bugs, enforces Go best practices
- Clear error messages and progress indicators — improves user experience
- Multiple output formats (table/json/yaml) — flexibility for automation

**Defer (v2+):**
- Shell completion scripts — nice but not essential for v1.0
- Watch mode for continuous monitoring — feature creep, defer to v1.1
- Interactive TUI — standard CLI is sufficient for v1
- Custom configuration files — flags are sufficient for now
- Multi-cluster support — kubectl handles context switching

### Architecture Approach

Well-structured kubectl plugins follow golang-standards/project-layout with interface-driven design for testability. The pattern is: thin CLI layer in cmd/ (just wiring), public orchestration layer in pkg/plugin/ with Complete/Validate/Run pattern, abstracted Kubernetes client in pkg/kubernetes/ with Interface and FakeClient, output formatting in pkg/output/ using strategy pattern, and shared utilities in pkg/image/ or pkg/util/.

**Major components:**
1. **CLI Layer (cmd/)** — Command definition and flag parsing only, <100 lines in main.go
2. **Plugin Layer (pkg/plugin/)** — Business logic orchestration with Options struct (Complete/Validate/Run pattern)
3. **Kubernetes Layer (pkg/kubernetes/)** — Interface abstraction over client-go with FakeClient for testing
4. **Output Layer (pkg/output/)** — Printer interface with strategy pattern for table/json/yaml formats
5. **Utilities Layer (pkg/image/, internal/util/)** — Shared functions like image parsing, spinners

### Critical Pitfalls

The most common mistakes when distributing kubectl plugins via krew fall into three categories: manifest/release automation issues, testing gaps, and documentation failures.

1. **Invalid krew manifest or checksums** — SHA256 mismatches block installation. Prevention: use GoReleaser's automatic checksum generation, test manifest locally before submission with `kubectl krew install --manifest=file --archive=local`.

2. **Missing cross-platform builds or wrong binary naming** — Plugin won't work on user platforms. Prevention: configure GoReleaser for all platforms (linux/darwin/windows on amd64/arm64), ensure binary name is exactly `kubectl-analyze_images`, test archives contain binary at correct path.

3. **Shipping untested code or no test coverage** — Breaks on user systems but works in dev. Prevention: write unit tests with fake Kubernetes client before refactoring, add GitHub Actions matrix testing across platforms, test actual krew installation flow locally.

4. **Release workflow failures or manual releases** — Can't scale, error-prone. Prevention: trigger GitHub Actions only on tags (`on: push: tags: ['v*']`), run tests before GoReleaser, use semantic versioning strictly (v1.0.0 format), grant workflow `contents: write` permission.

5. **Missing LICENSE or poor documentation** — krew-index rejects submissions. Prevention: add LICENSE file immediately (Apache-2.0 recommended for kubectl ecosystem), document krew installation in README, include usage examples and troubleshooting section.

## Implications for Roadmap

Based on research, suggested phase structure follows dependency order and risk mitigation principles. Start with low-risk utility cleanup and test infrastructure, then abstract dependencies for testability, then restructure business logic, and finally polish for distribution.

### Phase 1: Foundation & Testing Infrastructure
**Rationale:** Must have test infrastructure before refactoring. Utility extraction is low-risk and removes code duplication, enabling safe refactoring in later phases.
**Delivers:** Test framework, deduped utilities, baseline test coverage
**Addresses:** Zero test coverage (pitfall #3), duplicated image parsing logic
**Avoids:** Shipping untested code, breaking existing functionality during refactoring
**Duration:** Week 1

Key tasks:
- Add testify dependency, create test/ structure with fixtures
- Extract image parsing to pkg/image/parser.go (dedupe from analyzer + types)
- Create pkg/output/ with Printer interface (strategy pattern)
- Write unit tests for image parsing and output formatters
- Validate: make build && go test ./... passes

### Phase 2: Kubernetes Abstraction Layer
**Rationale:** Interface abstraction enables testing without cluster. This is medium risk but essential before restructuring business logic. Having tests from Phase 1 provides safety net.
**Delivers:** kubernetes.Interface, FakeClient, testable cluster interactions
**Uses:** Go interfaces, client-go/fake patterns
**Implements:** Kubernetes Layer from architecture (component #3)
**Duration:** Week 1-2

Key tasks:
- Create pkg/kubernetes/interface.go with ListPods(), ListNodes() methods
- Implement pkg/kubernetes/client.go wrapping client-go
- Create pkg/kubernetes/fake.go for testing
- Refactor internal/cluster/ to use Interface
- Add tests using FakeClient
- Validate: existing functionality unchanged, new tests pass

### Phase 3: Plugin Restructuring
**Rationale:** With interfaces and tests in place, can safely restructure business logic. Complete/Validate/Run is kubectl standard pattern. This is highest risk phase but dependencies are stable.
**Delivers:** Clean plugin architecture, minimal main.go, testable business logic
**Implements:** Plugin Layer architecture (component #2), CLI Layer (component #1)
**Duration:** Week 2-3

Key tasks:
- Create pkg/plugin/options.go with Complete/Validate/Run pattern
- Move internal/analyzer/ logic to pkg/plugin/analyzer.go
- Inject kubernetes.Interface and Printer dependencies
- Refactor main.go to <100 lines (just command setup)
- Add comprehensive plugin tests with FakeClient
- Validate: all features work, test coverage >70%

### Phase 4: Build & Release Automation
**Rationale:** Now that code is clean and tested, set up professional distribution. GoReleaser and GitHub Actions are table stakes for krew plugins.
**Delivers:** Automated multi-platform builds, release pipeline, krew manifest
**Uses:** GoReleaser v2, GitHub Actions, semantic versioning
**Addresses:** Manual releases (pitfall #4), cross-platform builds (pitfall #2)
**Duration:** Week 3-4

Key tasks:
- Create .goreleaser.yaml with multi-platform configuration
- Add .github/workflows/ci.yml (test, lint on PR)
- Add .github/workflows/release.yml (build on tag)
- Configure golangci-lint with .golangci.yml
- Test local release with goreleaser --snapshot
- Validate: CI green, release artifacts correct

### Phase 5: Krew Distribution
**Rationale:** Code is ready, builds work, now package for distribution. Krew manifest generation and submission are final steps.
**Delivers:** Krew manifest, documentation, first release
**Addresses:** Krew submission requirements (pitfall #1, #5)
**Duration:** Week 4 + review time

Key tasks:
- Generate krew manifest (GoReleaser or manual)
- Add LICENSE file (Apache-2.0)
- Polish README with krew installation, usage examples
- Create SECURITY.md, CONTRIBUTING.md
- Tag v1.0.0, verify release succeeds
- Test local install: kubectl krew install --manifest=file
- Submit PR to kubernetes-sigs/krew-index
- Validate: manifest validates, installation works

### Phase Ordering Rationale

- **Testing first (Phase 1):** Can't safely refactor without tests. Low-risk utility work builds confidence.
- **Interfaces before restructuring (Phase 2):** Need abstraction layer before moving business logic. Fake clients enable fast tests.
- **Architecture before automation (Phase 3 before 4):** Clean code is easier to build and test. Automation locks in quality.
- **Distribution last (Phase 5):** All other phases are prerequisites for krew submission. Can't submit incomplete work.
- **Parallel opportunities:** Documentation can progress alongside phases 1-3. Security setup (Dependabot, scanning) can happen in Phase 4.

This ordering follows the research consensus: test infrastructure → abstraction → restructuring → automation → distribution. It minimizes risk by validating at each phase and follows the critical path identified in PITFALLS.md.

### Research Flags

Phases with well-documented patterns (skip phase-specific research):
- **Phase 1-3:** Standard Go patterns, well-covered by ARCHITECTURE.md. No additional research needed.
- **Phase 4:** GoReleaser and GitHub Actions extensively documented in STACK.md. Follow cookbook.
- **Phase 5:** Krew submission process well-documented in PITFALLS.md and krew developer guide.

All phases can proceed with research already completed. No blocking unknowns identified.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | HIGH | GoReleaser + GitHub Actions is industry standard, verified with kubectl-tree/kubectx examples |
| Features | HIGH | Clear separation of table stakes vs nice-to-have from krew guidelines and plugin analysis |
| Architecture | HIGH | golang-standards/project-layout + kubectl patterns are well-established, multiple references |
| Pitfalls | HIGH | Based on krew-index requirements and common submission failures, verified with krew docs |

**Overall confidence:** HIGH

All research dimensions have high confidence due to:
- Multiple corroborating sources (ahmetb plugins, krew docs, Go standards)
- Clear industry consensus (GoReleaser for releases, GitHub Actions for CI)
- Well-documented patterns (Complete/Validate/Run, interface-driven design)
- Specific to domain (kubectl plugins have mature ecosystem)

### Gaps to Address

No significant gaps requiring validation during planning. Research covers all aspects of kubectl plugin cleanup and krew distribution. Minor points to validate during implementation:

- **Go version compatibility:** Research assumes Go 1.23, verify 1.24 compatibility in CI matrix testing (Phase 4)
- **Kubernetes version support:** Current client-go v0.29.0 supports K8s 1.29. Decide if need to update to v0.32+ for K8s 1.32 support. (Phase 2)
- **Binary size optimization:** Research doesn't cover size optimization. If binary >50MB, investigate with `go tool nm -size` during Phase 4.

These are implementation details, not blocking research gaps. Proceed to roadmap definition with current findings.

## Key Decisions

Based on research consensus, recommend these choices for project success:

| Decision | Rationale | Alternative Rejected | Risk |
|----------|-----------|----------------------|------|
| Use GoReleaser v2 for releases | Native krew support, proven with kubectl-tree | Manual scripts (error-prone, no checksum automation) | Low |
| GitHub Actions for CI/CD | Free for OSS, native integration | CircleCI/Travis (external account needed) | Low |
| Interface-driven architecture | Enables fast tests without cluster | Direct client-go usage (untestable) | Medium |
| Complete/Validate/Run pattern | kubectl standard, familiar to contributors | Custom structure (harder to understand) | Low |
| Defer configuration files to v2 | Flags sufficient for v1, avoid complexity | Config files now (YAGNI, premature) | Low |
| Start at v1.0.0 for krew submission | Signals production-ready | v0.x (suggests beta, less trust) | Low |
| Apache-2.0 license | Standard for kubectl ecosystem | MIT (also fine, but less common) | Low |

These decisions align with research findings and minimize project risk.

## Risk Register

Top risks and mitigations for roadmap execution:

| Risk | Impact | Likelihood | Mitigation | Phase |
|------|--------|------------|------------|-------|
| Breaking existing functionality during refactoring | HIGH | Medium | Test at each phase, keep old code working during transition, validate manually | Phase 2-3 |
| krew-index submission rejected | Medium | Low | Follow manifest checklist from PITFALLS.md, test locally before submission | Phase 5 |
| Cross-platform build failures | Medium | Low | Use GitHub Actions matrix, test on linux/darwin/windows before release | Phase 4 |
| Test coverage insufficient for confident refactoring | HIGH | Low | Phase 1 focused on test infrastructure, aim for 80%+ before Phase 3 | Phase 1 |
| Scope creep (adding features instead of cleanup) | Medium | Medium | Strict adherence to phased approach, defer features to v1.1+ explicitly | All phases |
| GoReleaser configuration errors | Low | Low | Test with --snapshot locally, follow cookbook from STACK.md | Phase 4 |
| Release workflow permissions issues | Low | Low | Grant contents:write permission explicitly in workflow | Phase 4 |

Overall project risk: **LOW**. Well-researched domain with established patterns, functional codebase, clear scope.

## Sources

### Primary (HIGH confidence)
- Krew Developer Guide (krew.sigs.k8s.io/docs/developer-guide/) — manifest requirements, submission process
- GoReleaser Docs (goreleaser.com) — v2 configuration, krew integration
- golang-standards/project-layout (github.com/golang-standards/project-layout) — directory structure standards
- kubectl-tree by ahmetb (github.com/ahmetb/kubectl-tree) — reference implementation, proven patterns
- kubectx by ahmetb (github.com/ahmetb/kubectx) — distribution patterns, release automation
- k8s.io/cli-runtime patterns — Complete/Validate/Run pattern, genericclioptions

### Secondary (MEDIUM confidence)
- kubectl-images by chenjiandongx — similar plugin, architecture patterns
- kubernetes-sigs/krew-index — submission requirements from actual PRs
- Go testing best practices — table-driven tests, interface-driven design

### Tertiary (LOW confidence)
- None. All research backed by official sources or proven implementations.

---
*Research completed: 2026-02-09*
*Ready for roadmap: yes*
