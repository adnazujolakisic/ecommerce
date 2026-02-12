#!/bin/bash
# Debug why mirrord Grafana dashboard shows "No data"
# Run this, then fix any issues and optionally: helm upgrade promtail grafana/promtail -n monitoring -f k8s/monitoring/promtail-values.yaml
set -e

echo "=== Mirrord Monitoring Diagnostics ==="
echo ""

echo "1. Mirrord operator pod (check name matches 'mirrord-operator-*'):"
kubectl get pods -n mirrord -o wide 2>/dev/null || echo "   (no mirrord namespace or pods)"
echo ""

echo "2. Operator logs (last 3 lines - should be JSON if jsonLog enabled):"
kubectl logs -n mirrord deployment/mirrord-operator --tail=3 2>/dev/null || echo "   (could not get logs)"
echo ""

echo "3. Promtail pods:"
kubectl get pods -n monitoring -l app.kubernetes.io/name=promtail 2>/dev/null || echo "   (no promtail)"
echo ""

echo "4. Promtail logs (recent errors):"
kubectl logs -n monitoring -l app.kubernetes.io/name=promtail --tail=20 2>&1 | grep -iE "error|warn|failed|loki" || echo "   (no obvious errors)"
echo ""

echo "5. Loki services:"
kubectl get svc -n monitoring | grep -E "loki|NAME" || true
echo ""

echo "6. Test: Run a mirrord session to generate logs, then in Grafana Explore try:"
echo "   {namespace=\"mirrord\"}"
echo "   {namespace=\"mirrord\", service_name=\"mirrord-operator\"}"
echo ""
