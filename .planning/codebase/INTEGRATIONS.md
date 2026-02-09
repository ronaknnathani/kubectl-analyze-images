# External Integrations

**Analysis Date:** 2026-02-09

## APIs & External Services

**Kubernetes API:**
- Kubernetes API Server - Pod and Node resource queries
  - SDK/Client: k8s.io/client-go v0.29.0
  - Authentication: Via kubeconfig (OAuth2 token, client cert, or cloud provider)
  - Endpoints used:
    - `GET /api/v1/pods` - List pods with optional namespace and label selectors (`internal/cluster/client.go:ListPods`)
    - `GET /api/v1/nodes` - List nodes and query image status (`internal/cluster/client.go:GetImageSizesFromNodes`)

**Image Registry APIs:**
- Not used directly
- Image sizes extracted from Kubernetes node status (`node.Status.Images`)
- No direct calls to image registries (Docker Hub, ECR, GCR, etc.)

## Data Storage

**Databases:**
- Not applicable - No database integration

**File Storage:**
- Kubernetes cluster etcd (via k8s.io/client-go)
  - Source of truth for pod and node data
  - Read-only access pattern

**Caching:**
- Optional local file system cache (disabled by default in current code)
- Cache directory configurable via `AnalysisConfig.CacheDir`
- Cache TTL: 24 hours (default per `pkg/types/analysis.go`)
- Current implementation does not actively use cache - placeholder for future enhancement

## Authentication & Identity

**Auth Provider:**
- Kubernetes native authentication
  - Implementation: k8s.io/client-go kubeconfig loader (`internal/cluster/client.go:NewClient`)
  - Supports multiple auth methods:
    - OAuth2 tokens (bearer token in kubeconfig)
    - Client certificates + CA cert
    - Cloud provider credentials (GKE, EKS, AKS)
    - Service account tokens (when running in-cluster)

**Kubeconfig Loading:**
- Standard location: `~/.kube/config`
- Context selection via `--context` flag or default current context
- Respects `KUBECONFIG` environment variable if set

## Monitoring & Observability

**Error Tracking:**
- None - All errors returned as Go error types and printed to stderr

**Logs:**
- Console output to stdout for reports
- Progress/status output to stderr via spinners
  - Library: github.com/briandowns/spinner v1.23.2
  - Used in `internal/cluster/client.go` and `internal/analyzer/pod_analyzer.go`
- Performance metrics included in output:
  - Pod query time
  - Node query time
  - Image analysis time
  - Total execution time

## CI/CD & Deployment

**Hosting:**
- Not a hosted service - Distributed as compiled binary for kubectl plugin installation
- Installation target: `~/.kube/plugins/analyze-images/` or `~/.local/bin/`

**CI Pipeline:**
- None configured - Build via Makefile only

## Environment Configuration

**Required env vars:**
- None strictly required
- `KUBECONFIG` (optional) - Overrides default kubeconfig location
- Kubernetes authentication credentials via kubeconfig file

**Secrets location:**
- Kubernetes API credentials stored in kubeconfig file
- Default: `~/.kube/config`
- No in-application secret storage

## Webhooks & Callbacks

**Incoming:**
- None

**Outgoing:**
- None - Read-only plugin, no external API calls initiated

## API Usage Pattern

**Query Pattern:**
- Kubernetes API calls use `ResourceVersion=0` for watch cache optimization
- Pagination via `k8s.io/client-go/tools/pager` with page size of 1000 items
- List options applied:
  - Optional namespace filter
  - Optional label selector filter
  - Watch cache enabled for performance

**Image Size Source:**
- Pod spec image names obtained from `Pod.Spec.Containers[].Image` and `Pod.Spec.InitContainers[].Image`
- Image sizes obtained from `Node.Status.Images[]` array (no registry queries)
- Image name canonicalization to avoid SHA digest duplicates via `selectBestImageName()` in `internal/cluster/client.go`

## Rate Limiting & Quotas

**Kubernetes API:**
- No explicit rate limiting configuration
- Default client-go rate limiter applies
- Query efficiency optimized via:
  - Watch cache usage (ResourceVersion=0)
  - Efficient pagination (1000 items per request)
  - Spinner-based progress feedback for long-running queries

---

*Integration audit: 2026-02-09*
