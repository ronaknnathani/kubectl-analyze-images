# Codebase Concerns

**Analysis Date:** 2026-02-09

## Tech Debt

**Image Registry Integration Not Implemented:**
- Issue: Registry client was completely removed (deleted `internal/registry/client.go`). Code claims to extract sizes "from node status" but abandoned actual OCI registry integration that would provide accurate, up-to-date image sizes.
- Files: `internal/analyzer/pod_analyzer.go` (lines 95-146), `pkg/types/image.go`
- Impact: Plugin only reports sizes that nodes have cached locally. For clusters with image layers not yet pulled, or for accurate size verification across registries, the tool is unreliable. Users may make sizing decisions based on stale node cache data.
- Fix approach: Either restore proper registry client integration (need docker/distribution dependency restored) or clearly document that only node-cached sizes are reported and recommend alternatives for accurate data.

**Duplicate Image Registry/Tag Extraction Logic:**
- Issue: Image name parsing logic (`extractRegistryAndTag`) exists in two places with identical implementation.
- Files: `internal/analyzer/pod_analyzer.go` (lines 149-168), `pkg/types/image.go` (lines 64-83)
- Impact: Maintenance burden - bug fixes or improvements must be applied in both locations. Increases chance of divergence.
- Fix approach: Move to shared utility module `pkg/utils/image.go`, import in both locations.

**Unused Configuration Fields Not Exercised:**
- Issue: `AnalysisConfig` struct defines concurrency, caching, retry logic, and page size settings that are defined but never used.
- Files: `pkg/types/analysis.go` (lines 7-16), `cmd/kubectl-analyze-images/main.go` (line 56)
- Impact: False sense of feature completeness. Users may expect retry/caching behavior that doesn't exist. Dead code maintenance burden.
- Fix approach: Either remove unused fields and simplify config, or implement the features they represent (concurrency for batch registry queries if registry integration restored, actual caching system).

## Known Bugs

**Image Parsing Incorrectly Handles Multi-Part Registry Names:**
- Issue: `extractRegistryAndTag` splits on "/" and assumes first part is registry, but multi-part registry names (e.g., `gcr.io/project/image:tag` or `registry.example.com:5000/image:tag`) are not correctly parsed.
- Files: `internal/analyzer/pod_analyzer.go` (lines 150-167), `pkg/types/image.go` (lines 65-82)
- Trigger: Running on cluster with images from registries with multiple path components (very common with container registries).
- Current behavior: `gcr.io/my-project/app:v1` would extract registry as `gcr.io` correctly, but `internal-registry.company.com:5000/team/app:v1` extracts registry as `internal-registry.company.com:5000` then fails to extract tag correctly due to string split order.
- Fix approach: Use proper Docker reference parsing library (already in deleted registry client), don't rely on naive string splitting.

**SHA Digest Detection Overly Specific:**
- Issue: `containsSHA()` function checks for exactly 64-character hex strings, but SHA256 digests in image names may be truncated or use different formats.
- Files: `internal/cluster/client.go` (lines 235-257)
- Trigger: Images with non-standard digest formats or images pulled by short digest.
- Current behavior: Falls back to first name when digest detection fails, potentially returning non-canonical names.
- Fix approach: More robust digest detection (check common prefixes like `sha256:`, `sha512:`).

## Security Considerations

**No TLS Verification Configuration for Cluster Access:**
- Risk: If implementing registry integration in future, there's no TLS/certificate validation settings exposed. Plugin would inherit only what client-go provides.
- Files: `internal/cluster/client.go` (lines 28-51)
- Current mitigation: Only queries Kubernetes API (node status), not external registries. kubeconfig TLS handled by client-go defaults.
- Recommendations: If adding registry access, explicitly handle certificate validation, add `--insecure` flag options, document certificate requirements.

**No Input Validation on Label Selectors or Namespace Names:**
- Risk: Label selectors and namespace values passed directly to Kubernetes API without validation. Malformed selectors could cause unexpected behavior or expose cluster information.
- Files: `cmd/kubectl-analyze-images/main.go` (lines 33, 75), `internal/cluster/client.go` (lines 54-77)
- Current mitigation: Kubernetes API validates/rejects invalid selectors. Error handling wraps errors.
- Recommendations: Add client-side validation for common selector errors, provide helpful error messages.

**Inaccessible Images Silently Dropped from Metrics:**
- Risk: No clear user indication that images couldn't be accessed. Metrics show "INACCESSIBLE" in table but totals exclude them silently.
- Files: `internal/analyzer/pod_analyzer.go` (lines 100-125), `internal/reporter/report.go` (lines 115-126)
- Current mitigation: "INACCESSIBLE" label in table output, but totals (TotalSize, counts) only count accessible images.
- Recommendations: Add explicit counter for inaccessible images in performance metrics, warn users in summary.

## Performance Bottlenecks

**Inefficient Percentile Calculation in Visualization:**
- Problem: Linear scan through all bin items to find actual image sizes during percentile calculation.
- Files: `pkg/types/visualization.go` (lines 217-229)
- Cause: Histogram bins store image names as strings, then linear search through all images to get sizes.
- Current: For each item in bins, iterates through entire analysis.Images array - O(n²) worst case.
- Improvement path: Store image sizes directly in histogram bins, or pre-sort sizes array during histogram creation.

**Map Lookups Not Optimized:**
- Problem: Checking image existence in `imageSizes` map for every image in analysis (line 101-103).
- Files: `internal/analyzer/pod_analyzer.go` (lines 100-125)
- Cause: Sequential map lookups in tight loop for 1000s of images.
- Current behavior: Works but not leveraging Go map efficiency (should be fast, but pattern could be streamlined).
- Improvement path: Reverse logic - iterate imageSizes map and check if in pods list, rather than iterating pods and checking map.

**Spinner Updates Too Frequent:**
- Problem: Spinner text updated every 100 pods, but for clusters with 50-100 pods this means constant terminal writes.
- Files: `internal/cluster/client.go` (lines 93-97), `internal/cluster/client.go` (lines 167-170)
- Cause: Fixed-interval updates without rate limiting.
- Impact: Small clusters see constant flickering, minimal impact on large clusters.
- Fix approach: Implement time-based rate limiting (e.g., update every 100ms) instead of count-based.

## Fragile Areas

**Image Name Parsing Tightly Coupled Across Package Boundaries:**
- Files: `internal/analyzer/pod_analyzer.go`, `pkg/types/image.go`, `pkg/types/visualization.go` (multiple formatting functions)
- Why fragile: Same logic duplicated, different functions assume different parsing strategies (split on "/" vs split on ":" vs split on "@"), no shared contract for what constitutes valid image name.
- Safe modification: Create `pkg/utils/imageref/parser.go` with single source of truth. Add comprehensive test suite for edge cases (registries with ports, digest formats, custom domains, ECR ARNs).
- Test coverage: Zero - no unit tests for image parsing at all.

**Histogram Rendering Depends on Specific Analysis Structure:**
- Files: `pkg/types/visualization.go` (lines 127-253, especially 220-228)
- Why fragile: `RenderASCII()` needs access to original `analysis.Images` to cross-reference bin items back to sizes. If Image structure changes or bin storage changes, rendering breaks.
- Safe modification: Refactor histogram generation to store computed sizes directly rather than item names, making rendering self-contained.
- Test coverage: Zero - no unit tests for visualization.

**Spinner Error Handling Inconsistent:**
- Files: `internal/cluster/client.go` (lines 102-105 for ListPods, 175-177 for GetImageSizesFromNodes)
- Why fragile: Both manually call `s.Stop()` before returning error, but this pattern could be missed in new code. If spinner not stopped, terminal state corrupted.
- Safe modification: Wrap spinner in `defer` pattern consistently, or create helper `withSpinner()` function that handles defer automatically.
- Test coverage: Zero - spinner behavior not tested.

**Missing Pod Ephemeral Container Handling:**
- Files: `pkg/types/pod.go` (lines 27-39)
- Why fragile: Only extracts images from regular containers and init containers. Ephemeral containers (added in K8s 1.16+, stable 1.25+) are ignored.
- Safe modification: Add extraction for `k8sPod.Spec.EphemeralContainers`. Add test case for pod with ephemeral container.
- Test coverage: Zero - no pod extraction tests.

## Scaling Limits

**No Memory Limits on Image Collection:**
- Current capacity: Untested. Theoretical limit is available system RAM.
- Limit: Cluster with 100,000+ unique images could consume significant memory (each Image struct ~200 bytes, adds up quickly).
- Scaling path: Implement streaming output for JSON, add pagination to table report, or add `--max-images` flag to limit results.

**No Pagination for Large Clusters:**
- Current: Uses kubernetes client-go pager with PageSize=1000 for lists, which is good.
- Limit: If cluster has 10,000 nodes (rare but possible in hyperscale setups), performance may degrade.
- Scaling path: Make PageSize configurable, add concurrent node processing.

## Dependencies at Risk

**Multiple Outdated Kubernetes Client-Go Versions:**
- Risk: Using `k8s.io/client-go v0.29.0` (released Aug 2023). Current stable is 0.31+.
- Impact: Missing performance improvements, potential security patches, compatibility issues with newer Kubernetes clusters.
- Migration plan: Update to latest stable v0.31 or v0.32, run integration tests against target cluster versions.

**Removed Docker Distribution Dependency:**
- Risk: Deleted `internal/registry/client.go` that used `github.com/docker/distribution`. This dependency is no longer in go.mod but was core to registry functionality.
- Impact: No path to restore registry integration without re-implementing image reference parsing.
- Migration plan: Add back `github.com/docker/distribution` or migrate to `github.com/opencontainers/image-spec`.

**Color Library with Minimal Maintenance:**
- Risk: `github.com/fatih/color v1.15.0` is stable but low-activity. Alternatives like github.com/charmbracelet/lipgloss are more active.
- Impact: No active bugs expected, but styling improvements limited. Cross-platform color support depends on fatih/color.
- Migration plan: Low priority, but consider if terminal color support needs enhancement.

**Spinner Library Version Locked:**
- Risk: `github.com/briandowns/spinner v1.23.2` is old. Repository shows minimal recent updates.
- Impact: Spinner characters may not work on newer terminal emulators or in certain CI environments.
- Migration plan: Evaluate, test compatibility if users report terminal issues.

## Missing Critical Features

**No Test Suite:**
- Problem: Zero test coverage. No unit tests, no integration tests, no end-to-end tests.
- Blocks: Can't safely refactor, can't detect regressions, can't verify behavior across Kubernetes versions.
- Priority: **HIGH** - Essential before expanding feature set or accepting contributions.

**No Error Recovery or Retry Logic:**
- Problem: Configuration defines retry count (3) but it's never used. Single API failure causes entire analysis to fail.
- Blocks: Plugin unusable in unreliable network conditions, large clusters with intermittent node status query failures.
- Priority: **MEDIUM** - Would improve reliability significantly.

**No Caching System:**
- Problem: Configuration defines caching (24hr TTL, cache directory) but never implemented. Disabled features in config.
- Blocks: Large clusters re-queried on every invocation - poor performance. Node status doesn't change frequently, caching would speed up repeated runs.
- Priority: **LOW** - works without it, but would improve performance significantly for users running repeatedly.

**No Export Capabilities:**
- Problem: README promises "Export capabilities" in roadmap but not implemented.
- Blocks: Users can't save results for comparison, trending, or reporting.
- Priority: **LOW** - CSV/Excel export would be useful but not blocking.

**No Deduplication Across Nodes:**
- Problem: Same image on multiple nodes counted separately. No layer-level deduplication.
- Blocks: Total size reported is meaningless for understanding actual disk usage (overcounts).
- Priority: **MEDIUM** - Size metrics misleading for users making resource decisions.

## Test Coverage Gaps

**No unit tests for image parsing:**
- What's not tested: `extractRegistryAndTag()` with edge cases - registry names with ports, image names with multiple slashes, SHA digests, OCI artifact names.
- Files: `internal/analyzer/pod_analyzer.go` (lines 149-168), `pkg/types/image.go` (lines 64-83)
- Risk: **HIGH** - Bugs in parsing manifest as wrong registries/tags, misleading results.
- Priority: **HIGH**

**No tests for Kubernetes API integration:**
- What's not tested: Pod listing with various label selectors, node status reading, pager behavior with large result sets, error cases.
- Files: `internal/cluster/client.go`
- Risk: **HIGH** - Can't verify behavior against real cluster, can't catch API changes.
- Priority: **HIGH**

**No tests for report generation:**
- What's not tested: JSON marshaling correctness, table formatting, histogram rendering, formatting of large sizes.
- Files: `internal/reporter/report.go`, `pkg/types/visualization.go`
- Risk: **MEDIUM** - Output format drift undetected, users see broken formatting.
- Priority: **MEDIUM**

**No integration test with mock Kubernetes API:**
- What's not tested: Full end-to-end flow with realistic mock cluster data, interaction between analyzer and reporter.
- Files: Entire codebase
- Risk: **MEDIUM** - Bugs in integration hidden until real cluster use.
- Priority: **MEDIUM**

---

*Concerns audit: 2026-02-09*
