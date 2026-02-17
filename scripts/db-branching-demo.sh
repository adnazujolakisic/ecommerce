#!/bin/bash
# DB Branching Demo - "I can't break production" wow moment
#
# Setup:
#   Terminal 1: cd services/inventory && mirrord exec -f ../../.mirrord/db-branching-steal.json -- go run .
#   Terminal 2: Run this script
#
# Flow: Compare cluster vs your branch, then reduce stock in YOUR branch only.

PRODUCT_ID="${1:-1}"
BRANCH_PORT="${BRANCH_PORT:-18082}"

echo ""
echo "=========================================="
echo "  DB Branching Demo"
echo "=========================================="
echo ""

echo "1️⃣  BEFORE - Check both databases:"
echo ""
echo "   CLUSTER (production):"
kubectl exec -n metalmart deployment/inventory -- curl -s http://localhost:8082/api/inventory/$PRODUCT_ID 2>/dev/null | jq -c '{stock_quantity, reserved_quantity}' 2>/dev/null || echo "     (cluster unreachable)"
echo ""
echo "   YOUR BRANCH:"
curl -s http://localhost:$BRANCH_PORT/api/inventory/$PRODUCT_ID 2>/dev/null | jq -c '{stock_quantity, reserved_quantity}' 2>/dev/null || echo "     (start mirrord first! cd services/inventory && mirrord exec -f ../../.mirrord/db-branching-steal.json -- go run .)"
echo ""

read -p "2️⃣  Press Enter to reduce stock in YOUR branch only (reserve 95 units)..."

curl -s -X POST "http://localhost:$BRANCH_PORT/api/inventory/reserve" \
  -H "Content-Type: application/json" \
  -d "{\"items\":[{\"productId\":\"$PRODUCT_ID\",\"quantity\":95}]}" | jq . 2>/dev/null || echo "   (reserve failed - is mirrord running?)"

echo ""
echo "3️⃣  AFTER - Check both again:"
echo ""
echo "   CLUSTER (production):"
kubectl exec -n metalmart deployment/inventory -- curl -s http://localhost:8082/api/inventory/$PRODUCT_ID 2>/dev/null | jq -c '{stock_quantity, reserved_quantity}' 2>/dev/null || echo "     (unchanged)"
echo ""
echo "   YOUR BRANCH:"
curl -s http://localhost:$BRANCH_PORT/api/inventory/$PRODUCT_ID 2>/dev/null | jq -c '{stock_quantity, reserved_quantity}' 2>/dev/null || echo "     (check local)"
echo ""
echo "=========================================="
echo "  Cluster unchanged. Your branch only. That's the wow."
echo "=========================================="
