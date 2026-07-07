# kubectl-analyze-images

Go-based kubectl plugin for visualizing the largest and most-used container images in a Kubernetes cluster. The tool is read-only: it lists pods and nodes, combines pod usage with node status image-size data, and writes reports to stdout.

## Build, test, and lint

Use the Makefile targets; they pin and bootstrap tooling where needed.

```bash
make build          # go mod tidy, tests, lint, then build
make test           # unit tests
make lint           # pinned golangci-lint v1.64.8
make ci             # local equivalent of GitHub Actions checks
make test-coverage  # coverage report
make snapshot       # local GoReleaser snapshot
```

Run `make ci` before pushing meaningful code, workflow, or release changes. For docs-only changes, at least run `git diff --check`; run more if docs touch commands, install, release, or krew instructions.

## Development conventions

- Follow standard Go conventions and keep changes focused.
- Add or update tests for behavior changes.
- Prefer table-driven tests and `testify` assertions, matching existing tests.
- Use fake Kubernetes clients for cluster interaction tests; do not require a live cluster for unit tests.
- Keep stdout clean for command output. Progress and status messages belong on stderr, especially for JSON mode.
- Preserve the plugin's read-only behavior; do not add cluster mutations.
- Do not commit generated artifacts such as `kubectl-analyze-images`, `coverage.out`, `dist/`, or `.tools/`.

## Git, commits, and PRs

- Keep PRs focused on one logical change.
- Use descriptive commit messages, e.g. `Add krew manifest checksums` or `Fix image summary output`.
- Include an appropriate co-author trailer when committing from an AI-assisted session, using the identity of the agent that made the change:

  ```text
  Co-authored-by: <Agent Name> <agent-email@example.com>
  ```

- Before changing GitHub repo metadata, release text, or other user-facing descriptions, show the exact proposed text first.
- Prefer `gh` for GitHub operations.
- After pushing, check GitHub Actions and report whether CI passed.

## Release and distribution notes

- Tags are annotated semver tags (`vX.Y.Z`) and trigger the Release workflow.
- GoReleaser publishes archives and `checksums.txt`.
- After a release, update `plugins/analyze-images.yaml` with the new version, asset URLs, and SHA256 checksums before krew distribution.

## Krew best-practices checklist

Before submitting or updating the krew manifest, confirm the plugin follows the krew developer best practices:

- Help and usage text should show the kubectl form (`kubectl analyze-images`), not just the binary name.
- Support common kubectl flags where practical: `-h`/`--help`, `-n`/`--namespace`, `--context`, and `--kubeconfig`.
- Do not add `-A`/`--all-namespaces` unless the default behavior changes; this plugin already analyzes all namespaces when `--namespace` is omitted.
- Import Kubernetes client-go auth plugins with `_ "k8s.io/client-go/plugin/pkg/client/auth"` so cloud-provider kubeconfigs work.
- Keep manifest `metadata.name` as the plugin name without the `kubectl-` prefix.
- Manifest URLs must point to immutable versioned release artifacts, never `latest`.
- Every manifest platform entry must include the SHA256 from the release `checksums.txt`.
- Keep `caveats` accurate about required RBAC; this plugin needs read/list access to pods and nodes.
- Validate the manifest YAML parses, all asset URLs return HTTP 200, and manifest checksums match the release before opening a krew-index PR.
