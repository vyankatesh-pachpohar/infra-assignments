## Summary

A Go HTTP config service backed by PostgreSQL, deployed to a local Kubernetes
cluster (kind), with infrastructure provisioned via Terraform and application
deployment packaged as a Helm chart.

## 1. Infrastructure Design

- **Cluster choice:** kind (Kubernetes-in-Docker) — fast to create/tear down,
  Docker-native, good CI parity if this were ever automated.
- **Split of responsibility:** Terraform provisions the namespace, ConfigMap,
  and Secret (the environment-setup layer). Helm packages and deploys the
  application and Postgres workloads on top. This follows the assignment's
  suggested pattern of "Terraform mainly for environment setup, with
  Kubernetes manifests/Helm for deployment" — chosen because Terraform's
  state/plan model suits slow-changing environment config, while Helm's
  templating and release/rollback model suits the app/DB workloads that
  change more often.
- **Database provisioning:** Postgres runs as a single-replica Deployment
  inside the cluster, backed by a PersistentVolumeClaim so data survives pod
  restarts. For production, this would move to a managed service (e.g. RDS)
  rather than self-hosted in-cluster Postgres.
- **Service exposure:** App exposed via a NodePort Service on a fixed port
  (30080) for predictable local access; Postgres exposed only via a
  ClusterIP Service, reachable only from inside the cluster.

## 2. Configuration and Secret Handling

- Non-sensitive config (ports, hostnames, DB name) lives in a Kubernetes
  ConfigMap, provisioned by Terraform.
- DB credentials live in a Kubernetes Secret, also provisioned by Terraform.
- Both are injected into pods as environment variables — the app never
  hardcodes connection details.
- **Known limitation:** Kubernetes Secrets are only base64-encoded, not
  encrypted at rest by default. Acceptable for this local assignment; in
  production I'd use a dedicated secrets manager (AWS Secrets Manager, Vault,
  or the External Secrets Operator) rather than relying on K8s Secrets alone.

## 3. Operational Readiness

- **Health checks:** `/ping` checks actual DB connectivity (not just process
  liveness), returning 503 if Postgres is unreachable.
- **Readiness vs. liveness:** both probes currently hit `/ping`. This is a
  deliberate simplification with a known trade-off: tying liveness to a
  downstream dependency (DB) means a Postgres outage could cause the app pod
  to be restarted repeatedly, which won't fix the underlying outage. A
  stricter design would give liveness its own lightweight check independent
  of DB reachability, and let readiness alone reflect DB status. Observed
  this directly during testing — app pods restarted once during initial
  rollout while Postgres was still starting.
- **Failure handling:** if the DB is down, `/ping` returns 503 and
  `GET/POST /configs` will surface a 500 with a logged error rather than
  hanging or crashing.
- **Data persistence:** Postgres uses a PVC, so pod restarts don't lose data.
- **Resource limits:** app containers have explicit CPU/memory
  requests/limits set to avoid one pod starving the node.

## 4. Repository Structure

\`\`\`
cmd/server/          — main.go, wiring/entrypoint
internal/domain/     — Config struct (data model)
internal/repository/ — all Postgres/SQL access, isolated here only
internal/service/    — business logic (validation, defaults), no HTTP or SQL specifics
internal/handler/    — thin HTTP handlers, translate HTTP <-> service layer
terraform/           — namespace, configmap, secret provisioning
chart/               — Helm chart: app + Postgres Deployments/Services/PVC
migrations/          — SQL schema
Makefile             — cluster-up, infra-apply, deploy, migrate, validate, cluster-down
APPROACH.md          — design doc written before implementation
\`\`\`
This layered structure (handler → service → repository → domain) keeps HTTP
concerns, business logic, and persistence cleanly separated, so each layer
can be tested or changed independently.

## 5. Responsible AI Usage Disclosure

I used Claude (Anthropic) as a learning and pairing tool while building this
assignment. My hands-on production experience is with AWS ECS/Fargate; I have
not previously operated Terraform or Kubernetes in a production setting, so
I used AI assistance to:

- Learn the Terraform Kubernetes-provider syntax and Helm chart structure
- Debug specific errors as they came up (Docker/WSL permissions, kind image
  loading, Go/Docker base-image version mismatches)
- Discuss design trade-offs (e.g., Terraform/Helm split of responsibility,
  readiness vs. liveness probe design)

I wrote, ran, and personally verified every command and piece of code in this
repo myself — every file was tested against a real local cluster (kind), and
the endpoint behavior shown above was validated with live `curl` requests
against a running Postgres instance, not assumed or copied blind. Where I was
uncertain of a concept (e.g., the liveness/readiness trade-off above), I've
flagged it explicitly rather than presenting it as fully solved.

## Testing

Given the 3-5 hour time scope, I prioritized infra/deployment depth
(Terraform, Helm, full teardown-and-rebuild validation) over an automated
test suite. What I validated instead:

- Live `curl` smoke tests against all 3 endpoints + the 404 path, run twice:
  once on initial deploy, once after a full clean-slate rebuild
  (`kind delete cluster` -> full re-provision), proving reproducibility.
- `make validate` captures this as a repeatable smoke-test target.

What I'd add next with more time:
- Table-driven Go tests for the handler layer (mocking the repository)
- An integration test using a real Postgres via `docker-compose` or
  testcontainers, covering the upsert-then-get round trip and the
  not-found path
- A CI workflow (GitHub Actions) running `go vet`, `go test`, and
  `terraform validate` on every PR
