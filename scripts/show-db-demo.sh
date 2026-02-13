#!/bin/bash
# Show cluster vs local (mirrord branch) inventory - before/after demo
# Usage: ./scripts/show-db-demo.sh [product_id]
#
# Run mirrord: cd services/inventory && mirrord exec -f ../../.mirrord/db-branching-steal.json -- go run .
# With PORT override in config, local listens on 18082. Without it, use BRANCH_PORT=8082.

PRODUCT_ID="${1:-1}"
BRANCH_PORT="${BRANCH_PORT:-18082}"

echo "=========================================="
echo "  Inventory: CLUSTER vs YOUR DB BRANCH"
echo "  Product ID: $PRODUCT_ID"
echo "=========================================="
echo ""

echo "📦 CLUSTER (production database):"
echo "-----------------------------------"
CLUSTER_OUT=$(kubectl exec -n metalmart deployment/inventory -- curl -s http://localhost:8082/api/inventory/$PRODUCT_ID 2>/dev/null)
if echo "$CLUSTER_OUT" | grep -q "product_id"; then
  echo "  (via inventory pod)"
  echo "  $CLUSTER_OUT"
else
  # Fallback: query main Postgres directly (works when inventory pod scaled down or unreachable)
  CLUSTER_DB=$(kubectl exec -n metalmart deployment/postgres -- psql -U postgres -d inventory -t -A -c "SELECT json_build_object('product_id',product_id,'stock_quantity',stock_quantity,'reserved_quantity',reserved_quantity) FROM inventory WHERE product_id='$PRODUCT_ID';" 2>/dev/null | tr -d '\n' | head -1)
  if [ -n "$CLUSTER_DB" ]; then
    echo "  (via postgres)"
    echo "  $CLUSTER_DB"
  else
    echo "  (could not reach cluster - is postgres running?)"
  fi
fi
echo ""

echo "📦 YOUR BRANCH:"
echo "-----------------------------------"
BRANCH_OUT=$(curl -s --connect-timeout 2 http://localhost:$BRANCH_PORT/api/inventory/$PRODUCT_ID 2>/dev/null)
if echo "$BRANCH_OUT" | grep -q "product_id"; then
  echo "  (via local mirrord port $BRANCH_PORT)"
  echo "  $BRANCH_OUT"
else
  # Fallback: query branch DB pod directly (when mirrord created one but local isn't reachable)
  BRANCH_POD=$(kubectl get pods -n metalmart --sort-by=.metadata.creationTimestamp -o name 2>/dev/null | grep "mirrord-postgres-branch-db" | tail -1 | cut -d/ -f2)
  if [ -n "$BRANCH_POD" ]; then
    echo "  (via kubectl - branch pod: $BRANCH_POD)"
    kubectl exec -n metalmart "$BRANCH_POD" -- psql -U postgres -d inventory -t -A -c "SELECT json_build_object('product_id',product_id,'stock_quantity',stock_quantity,'reserved_quantity',reserved_quantity) FROM inventory WHERE product_id='$PRODUCT_ID';" 2>/dev/null | head -1 || echo "  (query failed)"
  else
    echo "  (could not reach local or find branch pod)"
    echo ""
    echo "  → Start mirrord: cd services/inventory && mirrord exec -f ../../.mirrord/db-branching-steal.json -- go run ."
    echo "  → Use BRANCH_PORT=8082 if running without PORT override in config."
  fi
fi
echo ""

echo "-----------------------------------"
echo "Tip: Place an order with mirrord steal running, then run this script again to"
echo "see your branch change while cluster stays the same."
echo "=========================================="
