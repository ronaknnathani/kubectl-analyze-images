# Technology Stack Research: Kubectl Plugin Distribution

**Research Date:** 2025-02-09
**Research Scope:** Distribution stack for kubectl plugins via krew
**Target Audience:** Downstream roadmap and implementation planning

---

## Executive Summary

This document defines the prescriptive technology stack for distributing `kubectl-analyze-images` via krew. The stack is based on 2025 industry standards, with ahmetb's kubectl plugins (kubectx, kubectl-tree) as reference implementations.

**Key Decision:** Use GoReleaser v2 + GitHub Actions for a fully automated release pipeline that generates krew-compatible artifacts.

---

## 1. Build & Release Automation

### GoReleaser v2.x (Recommended: v2.4+)

**Confidence Level:** 🟢 High (Industry Standard)

**Why GoReleaser v2:**
- **Native krew support**: Built-in `.krew.yaml` manifest generation via `krew` announcement type
- **Multi-platform builds**: Automatic cross-compilation for linux/darwin/windows on amd64/arm64
- **Checksum automation**: SHA256 generation for krew plugin verification
- **Archive standardization**: Creates tar.gz/zip formats expected by krew
- **GitHub integration**: Seamless release creation and asset uploads
- **Version 2 improvements**: Better performance, SLSA provenance support, improved templates

**Why NOT alternatives:**
- ❌ **Manual builds**: Error-prone, doesn't scale, no checksum automation
- ❌ **Makefile-only**: Requires custom scripting for multi-platform builds and krew manifest generation
- ❌ **GoReleaser v1**: v2 has better krew integration and active maintenance

**Configuration Requirements:**
```yaml
# .goreleaser.yaml (v2 format)
version: 2

project_name: kubectl-analyze-images

before:
  hooks:
    - go mod tidy
    - go test ./...

builds:
  - id: kubectl-analyze-images
    main: ./cmd/kubectl-analyze-images
    binary: kubectl-analyze_images  # Note: underscore for krew compatibility
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w
      - -X main.version={{.Version}}
      - -X main.commit={{.Commit}}
      - -X main.date={{.Date}}

archives:
  - id: kubectl-analyze-images
    format_overrides:
      - goos: windows
        format: zip
    files:
      - LICENSE
      - README.md

checksum:
  name_template: 'checksums.txt'

release:
  github:
    owner: <your-username>
    name: kubectl-analyze-images
  draft: false
  prerelease: auto

announce:
  krew:
    enabled: true
    name: analyze-images
    index:
      owner: kubernetes-sigs
      name: krew-index
    commit_author:
      name: goreleaserbot
      email: bot@goreleaser.com
    commit_msg_template: "Krew plugin update for {{ .ProjectName }} version {{ .Tag }}"
    pull_request:
      enabled: true
      draft: false
```

**Version to Use:** v2.4.0 or later (January 2025 stable)

---

## 2. CI/CD Pipeline

### GitHub Actions

**Confidence Level:** 🟢 High (De facto standard for Go OSS)

**Why GitHub Actions:**
- **Native GitHub integration**: No external service dependencies
- **Free for public repos**: Unlimited minutes for OSS projects
- **Go ecosystem maturity**: `actions/setup-go@v5` is well-maintained
- **GoReleaser action**: Official `goreleaser/goreleaser-action@v6` for releases
- **Matrix builds**: Easy parallel testing across Go versions and platforms
- **Secrets management**: Built-in `GITHUB_TOKEN` and PAT support for krew-index PRs

**Why NOT alternatives:**
- ❌ **GitLab CI**: No advantage for GitHub-hosted project
- ❌ **CircleCI/Travis**: Requires external account, less Go ecosystem integration
- ❌ **Jenkins**: Overkill for simple Go plugin, maintenance burden

**Required Workflows:**

#### a) Continuous Integration (.github/workflows/ci.yml)
```yaml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    name: Test
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go-version: ['1.23', '1.24']  # Test current + next

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go-version }}
          cache: true

      - name: Verify dependencies
        run: go mod verify

      - name: Build
        run: go build -v ./...

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        if: matrix.os == 'ubuntu-latest' && matrix.go-version == '1.23'
        with:
          files: ./coverage.out

  lint:
    name: Lint
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: golangci-lint
        uses: golangci/golangci-lint-action@v6
        with:
          version: v1.61  # Latest as of Jan 2025
```

#### b) Release Workflow (.github/workflows/release.yml)
```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

permissions:
  contents: write  # For creating releases
  pull-requests: write  # For krew-index PR (if automated)

jobs:
  goreleaser:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history for changelog

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.23'
          cache: true

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: '~> v2'
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          # KREW_GITHUB_TOKEN: ${{ secrets.KREW_GITHUB_TOKEN }}  # Optional: for automated krew-index PR
```

**Action Versions (January 2025):**
- `actions/checkout@v4`
- `actions/setup-go@v5`
- `goreleaser/goreleaser-action@v6`
- `golangci/golangci-lint-action@v6`
- `codecov/codecov-action@v4`

---

## 3. Krew Plugin Manifest

### Krew Index Format (v0.4.x)

**Confidence Level:** 🟢 High (Stable specification)

**Why this format:**
- **Kubernetes SIG standard**: Official krew-index repository format
- **SHA256 verification**: Ensures binary integrity
- **Multi-platform support**: Single manifest for all OS/arch combinations
- **Semantic versioning**: Required for plugin updates

**Manifest Structure:**
```yaml
# plugins/analyze-images.yaml (in krew-index repo)
apiVersion: krew.googlecode.com/v1alpha2
kind: Plugin
metadata:
  name: analyze-images
spec:
  version: v0.1.0
  homepage: https://github.com/<username>/kubectl-analyze-images
  shortDescription: Analyze container image sizes in Kubernetes clusters
  description: |
    kubectl-analyze-images analyzes container image sizes across your
    Kubernetes cluster by examining node status and container runtime data.
    Provides insights into storage usage and identifies large images.
  caveats: |
    Requires cluster-admin or read access to nodes.
    Works with containerd and CRI-O runtimes.

  platforms:
  - selector:
      matchLabels:
        os: linux
        arch: amd64
    uri: https://github.com/<username>/kubectl-analyze-images/releases/download/v0.1.0/kubectl-analyze-images_linux_amd64.tar.gz
    sha256: <checksum>
    bin: kubectl-analyze_images

  - selector:
      matchLabels:
        os: linux
        arch: arm64
    uri: https://github.com/<username>/kubectl-analyze-images/releases/download/v0.1.0/kubectl-analyze-images_linux_arm64.tar.gz
    sha256: <checksum>
    bin: kubectl-analyze_images

  - selector:
      matchLabels:
        os: darwin
        arch: amd64
    uri: https://github.com/<username>/kubectl-analyze-images/releases/download/v0.1.0/kubectl-analyze-images_darwin_amd64.tar.gz
    sha256: <checksum>
    bin: kubectl-analyze_images

  - selector:
      matchLabels:
        os: darwin
        arch: arm64
    uri: https://github.com/<username>/kubectl-analyze-images/releases/download/v0.1.0/kubectl-analyze-images_darwin_arm64.tar.gz
    sha256: <checksum>
    bin: kubectl-analyze_images

  - selector:
      matchLabels:
        os: windows
        arch: amd64
    uri: https://github.com/<username>/kubectl-analyze-images/releases/download/v0.1.0/kubectl-analyze-images_windows_amd64.zip
    sha256: <checksum>
    bin: kubectl-analyze_images.exe
```

**Key Manifest Requirements:**
1. **Binary naming**: Must be `kubectl-<plugin-name>` with underscores (not hyphens) for multi-word names
2. **Archive format**: tar.gz for unix, zip for windows
3. **SHA256 checksums**: Mandatory for security verification
4. **Platform matrix**: Minimum linux/darwin amd64, recommended to include arm64
5. **Descriptions**: Short (< 120 chars) and long descriptions for discoverability

**GoReleaser Integration:**
- GoReleaser v2's `announce.krew` section auto-generates this manifest
- Checksums automatically populated from `checksums.txt`
- Can create PR to krew-index automatically with proper token setup

---

## 4. Testing Framework

### Go Standard Library + Testify

**Confidence Level:** 🟢 High (Best practice for CLI tools)

**Why this stack:**
- **Native testing**: Go's built-in `testing` package is mature and fast
- **Testify assertions**: Cleaner test syntax without magic
- **Table-driven tests**: Standard Go pattern for comprehensive coverage
- **Cobra testing**: `cobra.Command.ExecuteC()` enables CLI testing
- **Kubernetes fake clients**: `client-go/kubernetes/fake` for unit tests
- **No external dependencies**: Tests run in CI without special setup

**Why NOT alternatives:**
- ❌ **Ginkgo/Gomega**: Overkill for CLI tool, BDD style unnecessary
- ❌ **GoConvey**: Dated, less active maintenance
- ❌ **Pure stdlib**: Testify improves readability without adding complexity

**Testing Stack:**

#### Core Dependencies:
```go
// go.mod additions
require (
    github.com/stretchr/testify v1.9.0  // Assertions and mocking
    k8s.io/client-go v0.29.0  // Includes fake clients
)
```

#### Test Structure:
```
.
├── cmd/kubectl-analyze-images/
│   └── main_test.go              # Integration tests
├── internal/
│   ├── analyzer/
│   │   └── pod_analyzer_test.go  # Unit tests with fake k8s client
│   ├── cluster/
│   │   └── client_test.go        # Client initialization tests
│   └── reporter/
│       └── report_test.go        # Output formatting tests
└── pkg/types/
    └── image_test.go             # Type validation tests
```

#### Test Patterns:

**1. Unit Tests with Table-Driven Pattern:**
```go
func TestPodAnalyzer_AnalyzeImages(t *testing.T) {
    tests := []struct {
        name    string
        pods    []*corev1.Pod
        want    []types.ImageInfo
        wantErr bool
    }{
        {
            name: "single pod with one image",
            pods: []*corev1.Pod{...},
            want: []types.ImageInfo{...},
            wantErr: false,
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Use fake k8s client
            clientset := fake.NewSimpleClientset()
            analyzer := NewPodAnalyzer(clientset)

            got, err := analyzer.AnalyzeImages(context.Background(), "default")

            if tt.wantErr {
                assert.Error(t, err)
                return
            }

            assert.NoError(t, err)
            assert.Equal(t, tt.want, got)
        })
    }
}
```

**2. CLI Integration Tests:**
```go
func TestRootCommand(t *testing.T) {
    tests := []struct {
        name    string
        args    []string
        wantOut string
        wantErr bool
    }{
        {
            name:    "help flag",
            args:    []string{"--help"},
            wantOut: "Analyze container image sizes",
            wantErr: false,
        },
        // More cases...
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            cmd := NewRootCommand()
            cmd.SetArgs(tt.args)

            output := new(bytes.Buffer)
            cmd.SetOut(output)

            err := cmd.Execute()

            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Contains(t, output.String(), tt.wantOut)
            }
        })
    }
}
```

**3. E2E Tests (Optional, for post-distribution validation):**
```bash
# test/e2e/install_test.sh
#!/bin/bash
set -euo pipefail

# Test krew installation
kubectl krew install --manifest=test/fixtures/analyze-images.yaml
kubectl analyze-images --help
kubectl krew uninstall analyze-images
```

#### Test Coverage Goals:
- **Unit tests**: 80%+ coverage target
- **Integration tests**: Core CLI flows (help, version, basic execution)
- **Fake kubernetes clients**: All cluster interactions
- **Edge cases**: Error handling, invalid inputs, missing permissions

#### CI Test Execution:
```bash
# Makefile targets
test:
	go test -v -race ./...

test-coverage:
	go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
	go tool cover -html=coverage.out -o coverage.html

test-integration:
	go test -v -tags=integration ./test/integration/...
```

---

## 5. Additional Tooling

### golangci-lint v1.61+

**Confidence Level:** 🟢 High (Go standard)

**Why golangci-lint:**
- **Aggregates 100+ linters**: One tool for all static analysis
- **Fast**: Parallel execution and caching
- **Configurable**: `.golangci.yml` for project-specific rules
- **GitHub Action**: `golangci-lint-action@v6` integrates seamlessly
- **Pre-commit hook support**: Catch issues before CI

**Configuration (.golangci.yml):**
```yaml
run:
  timeout: 5m
  go: '1.23'

linters:
  enable:
    - gofmt
    - goimports
    - govet
    - errcheck
    - staticcheck
    - unused
    - gosimple
    - ineffassign
    - typecheck
    - revive
    - misspell
    - gocyclo
    - dupl
    - gosec

linters-settings:
  gocyclo:
    min-complexity: 15
  govet:
    enable-all: true
  revive:
    rules:
      - name: exported
        disabled: false
```

### Pre-commit Hooks (Optional but Recommended)

**Tool:** `pre-commit` framework with `golangci-lint` hook

**Why:**
- Catches formatting and lint issues before commits
- Reduces CI failures and feedback loop time
- Enforces code quality standards locally

**Configuration (.pre-commit-config.yaml):**
```yaml
repos:
  - repo: https://github.com/golangci/golangci-lint
    rev: v1.61.0
    hooks:
      - id: golangci-lint
        args: [--fix]

  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v4.6.0
    hooks:
      - id: trailing-whitespace
      - id: end-of-file-fixer
      - id: check-yaml
      - id: check-added-large-files
```

---

## 6. Documentation Tooling

### Standard Markdown + GitHub Pages (Optional)

**Confidence Level:** 🟡 Medium (Depends on documentation needs)

**Minimal Approach:**
- `README.md`: Installation, usage, examples
- `CONTRIBUTING.md`: Development setup, testing
- `docs/`: Extended guides if needed
- GitHub wiki for troubleshooting

**Why NOT complex doc frameworks:**
- ❌ **MkDocs/Docusaurus**: Overkill for kubectl plugin documentation
- ❌ **Sphinx**: Python ecosystem, not Go-native

**Essential Documentation:**
1. **README.md sections:**
   - Installation (krew + manual)
   - Quick start
   - Usage examples
   - Flags and options
   - Troubleshooting
   - License

2. **CONTRIBUTING.md:**
   - Development setup
   - Running tests
   - Submitting PRs
   - Release process

---

## 7. Version Management

### Semantic Versioning + Git Tags

**Confidence Level:** 🟢 High (Universal standard)

**Why:**
- **Krew requirement**: Plugins must use semver (v1.2.3 format)
- **GoReleaser trigger**: Releases on git tag push
- **Go module compatibility**: Follows Go versioning conventions

**Versioning Scheme:**
- **v0.y.z**: Pre-1.0 development (current phase)
- **v1.0.0**: First stable release after krew listing
- **vX.Y.Z**: Major.Minor.Patch per semver rules

**Version Embedding (ldflags):**
```go
// cmd/kubectl-analyze-images/main.go
var (
    version = "dev"
    commit  = "none"
    date    = "unknown"
)

func main() {
    rootCmd.Version = fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date)
    // ...
}
```

**Release Process:**
1. Update CHANGELOG.md
2. Create git tag: `git tag -a v0.1.0 -m "Release v0.1.0"`
3. Push tag: `git push origin v0.1.0`
4. GitHub Action triggers GoReleaser
5. GoReleaser creates GitHub release with binaries
6. (Manual or automated) Submit/update krew-index PR

---

## 8. What NOT to Use (Anti-Patterns)

### ❌ Docker/Container-Based Build Systems
- **Why not:** Adds complexity for Go cross-compilation that works natively
- **Alternative:** Use GoReleaser's native cross-compilation

### ❌ Custom Release Scripts
- **Why not:** Reinventing what GoReleaser does better, harder to maintain
- **Alternative:** Use GoReleaser with proper configuration

### ❌ Makefile for Cross-Platform Builds
- **Why not:** Prone to errors, lacks checksum automation, no krew manifest generation
- **Alternative:** Makefile for dev tasks only (test, lint, build-local), use GoReleaser for releases

### ❌ Bats (Bash Automated Testing System) for Go Tests
- **Why not:** Go has excellent native testing, shell tests are brittle
- **Alternative:** Go testing with fake kubernetes clients

### ❌ Manual Krew Manifest Maintenance
- **Why not:** Error-prone, especially checksums and version updates
- **Alternative:** Let GoReleaser generate the manifest

### ❌ Multiple CI Systems
- **Why not:** Maintenance burden, complexity for minimal gain
- **Alternative:** GitHub Actions is sufficient for all CI/CD needs

---

## 9. Dependency Versions (January 2025)

### Core Dependencies
```
Go: 1.23.x (current stable, testing with 1.24 for future compatibility)
k8s.io/client-go: v0.29.x (aligned with Kubernetes 1.29)
k8s.io/api: v0.29.x
k8s.io/apimachinery: v0.29.x
github.com/spf13/cobra: v1.8.x
github.com/stretchr/testify: v1.9.x
```

### Tool Versions
```
GoReleaser: v2.4.x or later
golangci-lint: v1.61.x
actions/checkout: v4
actions/setup-go: v5
goreleaser/goreleaser-action: v6
golangci/golangci-lint-action: v6
```

### Version Update Strategy
- **Go**: Update to latest minor within major version (1.23.x)
- **Kubernetes libraries**: Match latest stable K8s release (currently 1.29)
- **Tools**: Auto-update patch versions in CI, review major/minor updates
- **GitHub Actions**: Use major version tags (v4, v5) for auto-updates

---

## 10. Implementation Checklist

### Phase 1: Build & Release Automation
- [ ] Add `.goreleaser.yaml` with v2 configuration
- [ ] Create `.github/workflows/release.yml` for tag-triggered releases
- [ ] Test GoReleaser locally: `goreleaser release --snapshot --clean`
- [ ] Create initial v0.1.0 tag and verify release artifacts

### Phase 2: CI/CD
- [ ] Add `.github/workflows/ci.yml` for PR/push testing
- [ ] Configure matrix builds (Linux/macOS/Windows, Go 1.23/1.24)
- [ ] Add golangci-lint with `.golangci.yml` configuration
- [ ] Set up codecov or similar for coverage tracking

### Phase 3: Testing Infrastructure
- [ ] Add `testify` dependency to go.mod
- [ ] Write unit tests for core analyzer logic (target 80% coverage)
- [ ] Add CLI integration tests for cobra commands
- [ ] Create test fixtures and fake kubernetes clients
- [ ] Document testing approach in CONTRIBUTING.md

### Phase 4: Krew Distribution
- [ ] Generate krew manifest from first release (manual or via GoReleaser)
- [ ] Test installation: `kubectl krew install --manifest=./analyze-images.yaml`
- [ ] Submit PR to kubernetes-sigs/krew-index
- [ ] Address krew-index review feedback
- [ ] Document installation in README.md

### Phase 5: Polish
- [ ] Add pre-commit hooks configuration
- [ ] Update README with badges (build status, coverage, krew)
- [ ] Create CHANGELOG.md following Keep a Changelog format
- [ ] Add LICENSE file (Apache 2.0 recommended for kubectl ecosystem)
- [ ] Document release process in CONTRIBUTING.md

---

## 11. Reference Implementations

### Gold Standard Projects (ahmetb)

1. **kubectx** (https://github.com/ahmetb/kubectx)
   - GoReleaser configuration
   - GitHub Actions CI/CD
   - Krew plugin manifests for multiple plugins
   - Comprehensive testing

2. **kubectl-tree** (https://github.com/ahmetb/kubectl-tree)
   - Simpler single-binary structure (closer to your project)
   - Clean GoReleaser setup
   - Minimal but effective CI

### Other Reference Projects
- **stern** (https://github.com/stern/stern): Advanced CLI with multiple output formats
- **kubectl-node-shell** (https://github.com/kvaps/kubectl-node-shell): Complex plugin with dependencies

---

## 12. Confidence Levels Summary

| Component | Confidence | Rationale |
|-----------|-----------|-----------|
| GoReleaser v2 | 🟢 High | Industry standard, proven krew integration |
| GitHub Actions | 🟢 High | De facto for GitHub-hosted Go projects |
| Krew manifest format | 🟢 High | Stable v1alpha2 specification |
| Go testing + testify | 🟢 High | Best practice for CLI tools |
| golangci-lint | 🟢 High | Standard Go linting tool |
| Semantic versioning | 🟢 High | Universal standard, krew requirement |
| Pre-commit hooks | 🟡 Medium | Beneficial but optional |
| Documentation tooling | 🟡 Medium | Simple markdown sufficient |

---

## 13. Risks & Mitigation

### Risk 1: Krew Index Review Delays
- **Impact:** Weeks-long wait for PR review by maintainers
- **Mitigation:** Submit early, ensure manifest is perfect (checksums, descriptions), engage with reviewers promptly
- **Fallback:** Users can install via manual manifest file while waiting for index merge

### Risk 2: GoReleaser Breaking Changes
- **Impact:** CI failures on version updates
- **Mitigation:** Pin major version in GitHub Action (v6), test updates in separate branch, monitor GoReleaser changelog
- **Rollback:** Use explicit version tag if latest breaks: `goreleaser/goreleaser-action@v6.0.0`

### Risk 3: Kubernetes API Changes
- **Impact:** client-go compatibility issues with newer K8s versions
- **Mitigation:** Test against multiple K8s versions (1.27-1.29), follow Kubernetes deprecation notices
- **Strategy:** Update client-go quarterly, document supported K8s versions

### Risk 4: Multi-Platform Build Issues
- **Impact:** Binary doesn't work on specific OS/arch combination
- **Mitigation:** Test releases on all platforms before krew submission, use GitHub Actions matrix for validation
- **Coverage:** Minimum test on linux-amd64, darwin-amd64, darwin-arm64

---

## 14. Timeline Estimate

Based on current project state (working PoC):

- **Week 1**: GoReleaser setup, first release, local testing (Phase 1)
- **Week 2**: CI/CD workflows, golangci-lint integration (Phase 2)
- **Week 3**: Testing infrastructure, coverage to 80%+ (Phase 3)
- **Week 4**: Krew manifest, submission to krew-index, documentation polish (Phase 4-5)
- **Ongoing**: Krew index PR review (2-6 weeks typical response time)

**Total to krew submission:** ~4 weeks active development
**Total to krew listing:** 6-10 weeks including review

---

## Conclusion

This stack prioritizes:
1. **Automation**: GoReleaser + GitHub Actions eliminate manual release steps
2. **Standards compliance**: Krew manifest format, semver, Go conventions
3. **Quality**: Comprehensive testing, linting, CI coverage
4. **Maintainability**: Industry-standard tools with active communities
5. **Simplicity**: Avoid over-engineering for a kubectl plugin

**Next Step:** Proceed to roadmap creation (ROADMAP.md) with implementation phases based on this stack.

---

**Document Version:** 1.0
**Last Updated:** 2025-02-09
**Review Cycle:** Quarterly or on major tool updates
