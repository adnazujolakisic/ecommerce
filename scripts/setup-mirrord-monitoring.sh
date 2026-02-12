#!/bin/bash
# Setup mirrord Operator monitoring: Loki + Promtail + Grafana + Operator Dashboard
#
# Prerequisites:
#   - kubectl, helm, minikube (or other k8s cluster)
#   - MIRRORD_LICENSE_KEY env var (or pass as first argument)
#   - mirrord-operator already installed (for upgrade step)
#
# Usage:
#   ./scripts/setup-mirrord-monitoring.sh [LICENSE_KEY]
#
# Local-only: Refuses to deploy to remote clusters (GKE, EKS, AKS).
# Override with: MIRRORD_MONITORING_ALLOW_REMOTE=1 ./scripts/setup-mirrord-monitoring.sh
#
# After setup:
#   - Grafana: minikube service grafana -n monitoring  (or kubectl port-forward)
#   - Import dashboard: Grafana → Dashboards → Import → upload k8s/monitoring/mirrord-operator-dashboard.json
#   - Select "grafanacloud-logs" datasource → choose your Loki (named "Loki" or "grafanacloud-logs")

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
MONITORING_DIR="$REPO_ROOT/k8s/monitoring"
NAMESPACE="monitoring"
LICENSE_KEY="${1:-${MIRRORD_LICENSE_KEY}}"

echo "=== Mirrord Operator Monitoring Setup ==="

# 0. Safety check: refuse to deploy to remote clusters
if [[ "${MIRRORD_MONITORING_ALLOW_REMOTE}" != "1" ]]; then
  CONTEXT="$(kubectl config current-context 2>/dev/null || true)"
  if echo "$CONTEXT" | grep -qE 'gke_|eks\.|aks\.|arn:aws:|azure'; then
    echo "ERROR: Detected remote cluster (context: $CONTEXT)"
    echo ""
    echo "Monitoring (Loki, Promtail, Grafana) is for local development only."
    echo "It will NOT be deployed to remote clusters."
    echo ""
    echo "To force deployment anyway: MIRRORD_MONITORING_ALLOW_REMOTE=1 $0 $*"
    exit 1
  fi
fi

# 1. Add Helm repos
echo ""
echo ">>> Adding Helm repositories..."
helm repo add grafana https://grafana.github.io/helm-charts 2>/dev/null || true
helm repo add mirrord https://helm.metalbear.co 2>/dev/null || true
helm repo update

# 2. Create namespace
echo ""
echo ">>> Creating namespace $NAMESPACE..."
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# 3. Install Loki (monolithic, single replica)
echo ""
echo ">>> Installing Loki..."
if helm list -n $NAMESPACE | grep -q '^loki\t'; then
  helm upgrade loki grafana/loki -n $NAMESPACE -f "$MONITORING_DIR/loki-values.yaml"
else
  helm install loki grafana/loki -n $NAMESPACE -f "$MONITORING_DIR/loki-values.yaml"
fi

# 4. Wait for Loki
echo ""
echo ">>> Waiting for Loki to be ready..."
sleep 15
for i in {1..24}; do
  if kubectl get pods -n $NAMESPACE -l app.kubernetes.io/name=loki --no-headers 2>/dev/null | grep -q Running; then
    break
  fi
  sleep 5
done

# 5. Install Promtail
echo ""
echo ">>> Installing Promtail..."
if helm list -n $NAMESPACE | grep -q '^promtail\t'; then
  helm upgrade promtail grafana/promtail -n $NAMESPACE -f "$MONITORING_DIR/promtail-values.yaml"
else
  helm install promtail grafana/promtail -n $NAMESPACE -f "$MONITORING_DIR/promtail-values.yaml"
fi

# 6. Install Grafana
echo ""
echo ">>> Installing Grafana..."
GRAFANA_VALUES="$MONITORING_DIR/grafana-values.yaml"
if [[ ! -f "$GRAFANA_VALUES" ]]; then
  echo "   Warning: $GRAFANA_VALUES not found. Create it from k8s/monitoring/README.md"
  exit 1
fi
if helm list -n $NAMESPACE | grep -q '^grafana\t'; then
  helm upgrade grafana grafana/grafana -n $NAMESPACE -f "$GRAFANA_VALUES"
else
  helm install grafana grafana/grafana -n $NAMESPACE -f "$GRAFANA_VALUES"
fi

# 7. Upgrade mirrord operator with jsonLog (if license provided)
if [[ -n "$LICENSE_KEY" ]]; then
  echo ""
  echo ">>> Upgrading mirrord operator with JSON logging..."
  helm upgrade mirrord-operator mirrord/mirrord-operator \
    --set license.key="$LICENSE_KEY" \
    --set operator.kafkaSplitting=true \
    --set operator.dbBranching=true \
    --set operator.jsonLog=true \
    -n mirrord 2>/dev/null || {
      echo "   (Operator upgrade skipped - run manually if needed:)"
      echo "   helm upgrade mirrord-operator mirrord/mirrord-operator --set license.key=<KEY> --set operator.jsonLog=true -n mirrord"
    }
else
  echo ""
  echo ">>> Skipping mirrord operator upgrade (no LICENSE_KEY)."
  echo "   To enable JSON logging, run:"
  echo "   helm upgrade mirrord-operator mirrord/mirrord-operator --set license.key=<KEY> --set operator.jsonLog=true -n mirrord"
fi

# 8. Summary
echo ""
echo "=== Setup Complete ==="
echo ""
echo "Access Grafana:"
echo "  minikube service grafana -n $NAMESPACE"
echo "  (or: kubectl port-forward svc/grafana 3000:80 -n $NAMESPACE)"
echo ""
echo "Default login: admin / admin"
echo ""
echo "Import mirrord Operator Dashboard:"
echo "  1. Grafana → Dashboards → New → Import"
echo "  2. Upload: $MONITORING_DIR/mirrord-operator-dashboard.json"
echo "  3. Select datasource: Loki (grafanacloud-logs)"
echo ""
echo "Ensure mirrord operator has jsonLog enabled (Team/Enterprise plan):"
echo "  helm upgrade mirrord-operator mirrord/mirrord-operator --set license.key=<KEY> --set operator.jsonLog=true -n mirrord"
echo ""
