# Mirrord Operator Monitoring

Loki + Promtail + Grafana setup for monitoring the mirrord operator with the official [mirrord Operator Dashboard](https://github.com/metalbear-co/docs/blob/main/docs/managing-mirrord/assets/Mirrord_grafana_Operator_Dashboard.json).

## Prerequisites

- Kubernetes cluster (minikube, GKE, etc.)
- Helm 3+
- Mirrord operator installed (see [.mirrord/MIRRORD-SETUP.md](../.mirrord/MIRRORD-SETUP.md))
- **Team or Enterprise mirrord plan** (JSON logging is a paid feature)

## Quick Setup

**Local development only** — The script refuses to deploy to remote clusters (GKE, EKS, AKS) by default.

```bash
# Set your mirrord license key (for operator jsonLog upgrade)
export MIRRORD_LICENSE_KEY="your-license-key"

# Run the setup script (minikube, kind, k3d, docker-desktop only)
./scripts/setup-mirrord-monitoring.sh
```

To override and deploy to a remote cluster: `MIRRORD_MONITORING_ALLOW_REMOTE=1 ./scripts/setup-mirrord-monitoring.sh`

Or with license as argument:

```bash
./scripts/setup-mirrord-monitoring.sh "your-license-key"
```

## What Gets Installed

| Component | Purpose |
|-----------|---------|
| **Loki** | Log aggregation (monolithic, single replica, MinIO storage) |
| **Promtail** | Collects pod logs, adds `service_name=mirrord-operator` for mirrord pods |
| **Grafana** | Visualization with pre-provisioned Loki datasource |

## After Setup

### 1. Access Grafana

```bash
# Minikube
minikube service grafana -n monitoring

# Or port-forward
kubectl port-forward svc/grafana 3000:80 -n monitoring
```

Open http://localhost:3000 — login: **admin** / **admin**

### 2. Import mirrord Operator Dashboard

1. **Dashboards** → **New** → **Import**
2. **Upload JSON file** → select `k8s/monitoring/mirrord-operator-dashboard.json`
3. Select datasource: **Loki (grafanacloud-logs)**
4. Click **Import**

### 3. Enable JSON Logging on Mirrord Operator

The dashboard needs the operator to emit JSON logs. Upgrade the operator:

```bash
helm upgrade mirrord-operator mirrord/mirrord-operator \
  --set license.key="<YOUR_LICENSE_KEY>" \
  --set operator.kafkaSplitting=true \
  --set operator.dbBranching=true \
  --set operator.jsonLog=true \
  -n mirrord
```

Restart the operator pod if it was already running:

```bash
kubectl rollout restart deployment/mirrord-operator -n mirrord
```

## Dashboard Panels

- **Client Usage** — Session count by client
- **User Count** — Unique mirrord users
- **Target Sessions** — Target name, namespace, type
- **Events** — Session start/end, port steal, etc.

Logs are filtered by `{namespace="mirrord", service_name="mirrord-operator"}`.

## Manual Installation

If you prefer to install components separately:

```bash
# Add repos
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Create namespace
kubectl create namespace monitoring

# Install Loki
helm install loki grafana/loki -n monitoring -f k8s/monitoring/loki-values.yaml

# Install Promtail (after Loki is ready)
helm install promtail grafana/promtail -n monitoring -f k8s/monitoring/promtail-values.yaml

# Install Grafana
helm install grafana grafana/grafana -n monitoring -f k8s/monitoring/grafana-values.yaml
```

## Troubleshooting

### No data in dashboard

1. **Operator JSON logging enabled?**
   ```bash
   kubectl get deployment mirrord-operator -n mirrord -o yaml | grep jsonLog
   ```

2. **Promtail running?**
   ```bash
   kubectl get pods -n monitoring -l app.kubernetes.io/name=promtail
   ```

3. **Check Loki has logs:**
   - Grafana → Explore → select Loki
   - Query: `{namespace="mirrord"}`
   - If empty, check Promtail can reach Loki and mirrord namespace pods

### Wrong datasource on import

When importing the dashboard, ensure you select the **Loki** datasource with UID `grafanacloud-logs` (pre-provisioned in `grafana-values.yaml`).

## Uninstall

```bash
helm uninstall grafana promtail loki -n monitoring
kubectl delete namespace monitoring
```
