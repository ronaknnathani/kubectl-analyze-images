# kubectl-analyze-images

## What This Is

A kubectl plugin that analyzes container image sizes and distribution across a Kubernetes cluster by querying node status data. It provides histogram visualizations, size statistics, and filterable reports in table or JSON format. Distributed via krew with automated CI/CD and multi-platform binaries. Built for platform engineers and cluster operators who need visibility into image bloat.

## Core Value

Quickly show operators what images are in their cluster and how big they are -- using only node status data, no registry credentials needed.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. -->

- ✓ Query image sizes from node status (`node.Status.Images`) -- existing
- ✓ List pods with namespace and label selector filtering -- existing
- ✓ Cross-reference pod images with node image sizes -- existing
- ✓ Table output with formatted image sizes -- existing
- ✓ JSON output for programmatic consumption -- existing
- ✓ Color-coded histogram visualization of image size distribution -- existing
- ✓ Performance metrics (query time, analysis time) -- existing
- ✓ Top N images by size report -- existing
- ✓ Multi-cluster support via `--context` flag -- existing
- ✓ Progress spinners during cluster queries -- existing
- ✓ Kubernetes pagination for large clusters -- existing
- ✓ Clean, modular codebase ready for open-source contribution -- v1.0
- ✓ Comprehensive test suite (unit tests for all packages, 80.4% coverage) -- v1.0
- ✓ krew plugin manifest for distribution -- v1.0
- ✓ goreleaser configuration for multi-platform binaries (6 targets) -- v1.0
- ✓ GitHub Actions CI/CD pipeline (test, build, release) -- v1.0
- ✓ Remove dead code and unused config fields -- v1.0
- ✓ Deduplicate image parsing logic -- v1.0
- ✓ Proper LICENSE file (Apache-2.0) -- v1.0

### Active

<!-- Next milestone scope. -->

(None yet -- define with /gsd:new-milestone)

### Out of Scope

- Registry integration (direct image registry queries) -- intentional design choice; node status is sufficient and requires no credentials
- Layer-level analysis or deduplication -- v2+ consideration
- Caching system -- not needed with node status approach
- Scheduled/automated runs -- this is a CLI tool, not a daemon
- Interactive TUI -- standard CLI output is the target

## Context

- v1.0 shipped with 3,214 lines Go, 80.4% test coverage, 5 phases completed
- Architecture: thin CLI (50 lines) → plugin Complete/Validate/Run → kubernetes.Interface → cluster → analyzer → reporter
- All dependencies injectable; FakeClient enables full testing without real clusters
- CI/CD: GitHub Actions (test+lint on PR, goreleaser release on tags)
- Distribution: krew manifest ready (SHA256 populated post-release), Apache-2.0 license
- Go 1.23, Cobra CLI, client-go v0.29.0, testify v1.11.1

## Constraints

- **Tech stack**: Go, Cobra, client-go -- established, don't change
- **Distribution**: Must work as kubectl plugin (binary named `kubectl-analyze-images`)
- **Compatibility**: Kubernetes clusters v1.29+ (matching client-go)
- **Dependencies**: Minimize external dependencies for clean krew distribution
- **Permissions**: Read-only cluster access (pods, nodes) -- no write operations

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Node status for image sizes | No registry credentials needed, fast, works everywhere | ✓ Good |
| Remove registry client | Simplifies design, node status is sufficient for v1 | ✓ Good |
| Personal GitHub + krew | Own the roadmap, lower barrier than kubernetes-sigs | ✓ Good -- CI/CD and manifest ready |
| Code cleanup before krew packaging | Inverted original plan; clean architecture first proved valuable | ✓ Good -- enabled 80.4% test coverage |
| Full CI/CD pipeline | GitHub Actions + goreleaser is standard for krew plugins | ✓ Good -- 6 platform targets automated |
| Complete/Validate/Run pattern | Standard kubectl plugin architecture | ✓ Good -- clean DI, fully testable |
| Interface-driven Kubernetes abstraction | Enables FakeClient for testing without real cluster | ✓ Good -- 86-93% coverage on cluster/analyzer |
| GoReleaser v2 + CGO_ENABLED=0 | Static binaries, modern config format | ✓ Good -- portable across all targets |
| Apache-2.0 license | Standard for Kubernetes ecosystem projects | ✓ Good |

---
*Last updated: 2026-02-10 after v1.0 milestone*
