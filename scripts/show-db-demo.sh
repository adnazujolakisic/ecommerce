#!/bin/bash
# Show cluster vs local (mirrord branch) inventory - before/after demo
# Usage: ./scripts/show-db-demo.sh [product_id]
#
# Run mirrord on port 18082 to avoid conflict with port-forward (use 8082 for cluster):
#   cd services/inventory && PORT=18082 mirrord exec -f ../../.mirrord/db-branching.json -- go run .

PRODUCT_ID="${1:-1}"
BRANCH_PORT="${BRANCH_PORT:-18082}"

echo "=========================================="
echo "  Inventory: CLUSTER vs YOUR DB BRANCH"
echo "  Product ID: $PRODUCT_ID"
echo "=========================================="
echo ""

echo "📦 CLUSTER (production database - kubectl exec):"
echo "-----------------------------------"
CLUSTER_OUT=$(kubectl exec -n metalmart deployment/inventory -- curl -s -i http://localhost:8082/api/inventory/$PRODUCT_ID 2>/dev/null)
if [ -n "$CLUSTER_OUT" ]; then
  echo "$CLUSTER_OUT" | grep -E "^X-Database-Source:" || true
  echo "$CLUSTER_OUT" | sed -n '/^{/p' | head -1
else
  echo "  (could not reach cluster - is it running?)"
fi
echo ""

echo "📦 YOUR BRANCH:"
echo "-----------------------------------"
BRANCH_OUT=$(curl -s -i --connect-timeout 2 http://localhost:$BRANCH_PORT/api/inventory/$PRODUCT_ID 2>/dev/null)
if [ -n "$BRANCH_OUT" ]; then
  echo "  (via local mirrord port $BRANCH_PORT)"
  echo "$BRANCH_OUT" | grep -E "^X-Database-Source:" || true
  echo "$BRANCH_OUT" | sed -n '/^{/p' | head -1
else
  # Fallback: query branch DB pod directly (when mirrord created one but local isn't reachable)
  BRANCH_POD=$(kubectl get pods -n metalmart --sort-by=.metadata.creationTimestamp -o name 2>/dev/null | grep "mirrord-postgres-branch-db" | tail -1 | cut -d/ -f2)
  if [ -n "$BRANCH_POD" ]; then
    echo "  (via kubectl - branch pod: $BRANCH_POD)"
    kubectl exec -n metalmart "$BRANCH_POD" -- psql -U postgres -d inventory -t -A -c "SELECT json_build_object('product_id',product_id,'stock_quantity',stock_quantity,'reserved_quantity',reserved_quantity) FROM inventory WHERE product_id='$PRODUCT_ID';" 2>/dev/null | head -1 || echo "  (query failed)"
  else
    echo "  (could not reach local or find branch pod)"
    echo ""
    echo "  → Start mirrord: cd services/inventory && mirrord exec -f ../../.mirrord/db-branching.json -- go run ."
echo "  → For frontend to use your branch: mirrord exec -f ../../.mirrord/db-branching-steal.json -- go run ."
    echo "  → Config uses PORT=18082. If local still fails, a branch pod may exist - run again."
  fi
fi
echo ""

echo "-----------------------------------"
echo "Tip: Place an order with mirrord running (frontend uses cluster), then run this"
echo "script again to see your branch change while cluster stays the same."
echo "=========================================="
