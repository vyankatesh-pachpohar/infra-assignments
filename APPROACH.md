# Approach — Kubernetes Config Service

## Understanding
Build a Go HTTP service (GET /ping, GET /configs/:id, POST /configs) backed by PostgreSQL,
deployed to a local Kubernetes cluster (kind), provisioned via Terraform, packaged as a Helm chart.

## Architecture
Client -> Service (ClusterIP) -> Go app Deployment (2 replicas) -> Postgres Deployment (1 replica, PVC-backed)
Config: ConfigMap (app_name, log_level defaults)
Secrets: K8s Secret (DB credentials)

## Key Decisions
- kind for local cluster: fast, Docker-native, easy CI parity
- Terraform provisions: namespace, Secret, ConfigMap, Postgres (via kubernetes provider)
- Helm packages: the Go app deployment (templated, versioned, values.yaml per future environment)
- Postgres: single instance with PVC for data persistence across pod restarts
- /ping returns 503 if DB unreachable (fail-fast signal for readiness probe)
- Schema: configs(id TEXT PRIMARY KEY, host TEXT NOT NULL, port INT NOT NULL, app_name TEXT NOT NULL, log_level TEXT NOT NULL DEFAULT 'INFO', updated_at TIMESTAMPTZ NOT NULL DEFAULT now())

## Open Questions / Assumptions
- Treating POST /configs as full upsert (INSERT ... ON CONFLICT (id) DO UPDATE)
- Single DB instance acceptable for local scope; would use managed Postgres (RDS-equivalent) in production
- K8s Secret used for local dev; would use a real secrets manager in production

## Known Limitations Given Time Scope
- No HPA/autoscaling
- No CI pipeline (bonus, time-permitting)
- Single-node local cluster only.
