# db-migrator Helm chart

Reusable Helm chart that runs [`raoptimus/db-migrator`](https://hub.docker.com/r/raoptimus/db-migrator)
as a Kubernetes `Job`.

- **`release`** — applies all pending migrations atomically. Rendered as a Helm hook
  (`pre-install,pre-upgrade` by default), so the schema is ready **before** the application
  pods roll out.
- **`rollback`** — reverts the latest release batch. Rendered as an opt-in plain `Job`
  (disabled by default), triggered explicitly. It is not a lifecycle hook, so it works the
  same under both plain Helm and [werf](https://werf.io) `converge`.

The chart works standalone (`helm install`) and as a subchart dependency of an application
chart. `INTERACTIVE` is always forced to `false` because a Job has no TTY.

## Prerequisites

- Kubernetes cluster and Helm 3.
- A container image that contains your migration files at `migrator.path`
  (default `/migrations`). The base image ships **no** migrations — see
  [Providing migrations](#providing-migrations).
- An existing `Secret` holding the database DSN.

## Installation

```bash
kubectl create secret generic db-migrator-dsn \
  --from-literal=dsn='postgres://user:pass@postgres:5432/app?sslmode=disable'

helm install my-migrations ./charts/db-migrator \
  --set image.repository=myregistry/myapp-migrations \
  --set image.tag=1.4.2 \
  --set migrator.dsn.existingSecret=db-migrator-dsn
```

## Use as a subchart dependency

```yaml
# Chart.yaml of your application chart
dependencies:
  - name: db-migrator
    version: "0.1.0"
    repository: "oci://<registry>/charts"   # or https://<repo>, or file://../charts/db-migrator
```

```yaml
# values.yaml of your application chart (values go under the "db-migrator" key)
db-migrator:
  image:
    repository: myregistry/myapp-migrations
    tag: "1.4.2"
  migrator:
    dsn:
      existingSecret: myapp-db
      secretKey: dsn
    path: /migrations
```

Then `helm dependency update && helm upgrade --install ...` — the `release` hook runs on
every install/upgrade before the app pods start.

## Rolling back

`rollback` is opt-in. Enable it explicitly when you actually want to revert the latest batch:

```bash
# werf or plain Helm
helm upgrade --install my-migrations ./charts/db-migrator \
  --reuse-values --set rollback.enabled=true
```

Under **plain Helm** you may instead wire rollback to `helm rollback` by turning the Job into
a hook (this does not fire under werf converge):

```bash
helm install ... --set 'rollback.hookTypes={pre-rollback}'
helm rollback my-migrations
```

## Providing migrations

The base image contains no migrations. Recommended: build your own image.

```dockerfile
FROM raoptimus/db-migrator:1.7.0
COPY ./migrations /migrations
```

Alternatively, mount migrations without rebuilding the image via passthrough values, e.g. a
ConfigMap:

```yaml
extraVolumes:
  - name: migrations
    configMap:
      name: myapp-migrations
extraVolumeMounts:
  - name: migrations
    mountPath: /migrations
migrator:
  path: /migrations
```

`initContainers` is also passed through (e.g. for a git-sync sidecar populating an
`emptyDir`).

## Values

| Key | Default | Description |
|-----|---------|-------------|
| `image.repository` | `raoptimus/db-migrator` | Image containing the migrations |
| `image.tag` | `""` (→ `.Chart.AppVersion`) | Image tag |
| `image.pullPolicy` | `IfNotPresent` | Image pull policy |
| `imagePullSecrets` | `[]` | Image pull secrets |
| `migrator.dsn.existingSecret` | `""` (**required**) | Secret holding the DSN |
| `migrator.dsn.secretKey` | `dsn` | Key inside the Secret |
| `migrator.path` | `/migrations` | `MIGRATION_PATH` |
| `migrator.table` | `migration` | `MIGRATION_TABLE` |
| `migrator.clusterName` | `""` | `MIGRATION_CLUSTER_NAME` (ClickHouse) |
| `migrator.replicated` | `false` | `MIGRATION_REPLICATED` (ClickHouse) |
| `migrator.maxConnAttempts` | `1` | `MAX_CONN_ATTEMPTS` |
| `migrator.compact` | `false` | `COMPACT` |
| `migrator.dryRun` | `false` | `DRY_RUN` |
| `migrator.placeholderCustom` | `""` | `PLACEHOLDER_CUSTOM` |
| `migrator.extraEnv` | `[]` | Extra env entries |
| `migrator.extraEnvFrom` | `[]` | Extra `envFrom` sources |
| `release.enabled` | `true` | Render the release hook Job |
| `release.command` | `release` | Command (`release` or `up`) |
| `release.hookTypes` | `[pre-install, pre-upgrade]` | Helm hook phases |
| `release.weight` | `"5"` | Hook weight |
| `release.deletePolicy` | `before-hook-creation` | Hook delete policy |
| `release.annotations` | `{werf.io/fail-mode: FailWholeDeployProcessImmediately}` | Extra Job annotations |
| `rollback.enabled` | `false` | Render the rollback Job (opt-in) |
| `rollback.command` | `rollback` | Command |
| `rollback.hookTypes` | `[]` | Empty = plain Job; `[pre-rollback]` = hook |
| `rollback.weight` | `"5"` | Hook weight (when hookTypes set) |
| `rollback.deletePolicy` | `before-hook-creation` | Hook delete policy |
| `rollback.annotations` | `{}` | Extra Job annotations |
| `backoffLimit` | `0` | Job `backoffLimit` |
| `activeDeadlineSeconds` | `3600` | Job `activeDeadlineSeconds` |
| `ttlSecondsAfterFinished` | `30` | Job TTL after completion |
| `resources` | `{}` | Container resources |
| `serviceAccount.create` | `false` | Create a ServiceAccount |
| `serviceAccount.name` | `""` | ServiceAccount name |
| `nodeSelector` / `tolerations` / `affinity` | `{}` / `[]` / `{}` | Scheduling |
| `podSecurityContext` / `securityContext` | `{}` | Security contexts |
| `initContainers` / `extraVolumes` / `extraVolumeMounts` | `[]` | Migration delivery passthrough |
