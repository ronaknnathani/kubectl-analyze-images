# kubectl-analyze-images

## What This Is

A kubectl plugin that analyzes container image sizes and distribution across a Kubernetes cluster by querying node status data. It provides histogram visualizations, size statistics, and filterable reports in table or JSON format. Built for platform engineers and cluster operators who need visibility into image bloat.

## Core Value

Quickly show operators what images are in their cluster and how big they are -- using only node status data, no registry credentials needed.

## Requirements

### Validated

<!-- Shipped and confirmed valuable. Inferred from existing codebase. -->

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

### Active

<!-- Current scope. Building toward these. -->

- [ ] Clean, modular codebase ready for open-source contribution
- [ ] Comprehensive test suite (unit tests for all packages)
- [ ] krew plugin manifest for distribution
- [ ] goreleaser configuration for multi-platform binaries
- [ ] GitHub Actions CI/CD pipeline (test, build, release)
- [ ] Remove dead code and unused config fields
- [ ] Deduplicate image parsing logic (currently in two places)
- [ ] Proper LICENSE file

### Out of Scope

- Registry integration (direct image registry queries) -- intentional design choice; node status is sufficient and requires no credentials
- Layer-level analysis or deduplication -- v2+ consideration
- Caching system -- not needed with node status approach
- Scheduled/automated runs -- this is a CLI tool, not a daemon
- Interactive TUI -- standard CLI output is the target

## Context

- Working PoC with all core features functional
- Code has some duplication (image parsing in two places) and unused config fields (concurrency, caching, retry)
- Deleted registry client was intentional -- node status approach is the design
- Zero test coverage currently
- Target distribution: personal GitHub repo with krew index listing
- Go 1.23, Cobra CLI, client-go v0.29.0

## Constraints

- **Tech stack**: Go, Cobra, client-go -- established, don't change
- **Distribution**: Must work as kubectl plugin (binary named `kubectl-analyze_images` or `kubectl-analyze-images`)
- **Compatibility**: Kubernetes clusters v1.29+ (matching client-go)
- **Dependencies**: Minimize external dependencies for clean krew distribution
- **Permissions**: Read-only cluster access (pods, nodes) -- no write operations

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| Node status for image sizes | No registry credentials needed, fast, works everywhere | ✓ Good |
| Remove registry client | Simplifies design, node status is sufficient for v1 | ✓ Good |
| Personal GitHub + krew | Own the roadmap, lower barrier than kubernetes-sigs | -- Pending |
| krew packaging before code cleanup | Get distributable first, then iterate on quality | -- Pending |
| Full CI/CD pipeline | GitHub Actions + goreleaser is standard for krew plugins | -- Pending |

---
*Last updated: 2026-02-09 after initialization*
