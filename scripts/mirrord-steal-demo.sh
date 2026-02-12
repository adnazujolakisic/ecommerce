#!/bin/bash
# Mirrord steal demo: frontend uses your DB branch for inventory display AND checkout
# Run this, then place orders - production DB stays unchanged

set -e
cd "$(dirname "$0")/.."

echo "=== Mirrord Steal Demo Setup ==="
echo ""

# 1. Scale inventory to 1 replica (steal works best with single pod)
echo "1. Scaling inventory to 1 replica..."
kubectl scale deployment/inventory -n metalmart --replicas=1 2>/dev/null || true
kubectl rollout status deployment/inventory -n metalmart --timeout=60s 2>/dev/null || true

# 2. Restart checkout (clears any cached connections to inventory)
echo "2. Restarting checkout to clear connections..."
kubectl rollout restart deployment/checkout -n metalmart
kubectl rollout status deployment/checkout -n metalmart --timeout=60s 2>/dev/null || true

echo ""
echo "=== Ready! Run these in separate terminals ==="
echo ""
echo "Terminal 1 - Start mirrord (inventory with DB branch + steal):"
echo "  cd services/inventory && mirrord exec -f ../../.mirrord/db-branching-steal.json -- go run ."
echo ""
echo "Terminal 2 - Start minikube tunnel for frontend:"
echo "  minikube service frontend -n metalmart"
echo "  (Note the URL, e.g. http://127.0.0.1:55587)"
echo ""
echo "Terminal 3 - Start frontend with branch inventory display:"
echo "  FRONTEND_URL=\$(minikube service frontend -n metalmart --url 2>/dev/null | head -1)"
echo "  cd frontend && VITE_PROXY_TARGET=\$FRONTEND_URL VITE_INVENTORY_API=http://localhost:18082 npm run dev"
echo "  Then open http://localhost:5173"
echo ""
echo "Or use minikube frontend directly (steal must capture checkout traffic):"
echo "  Open the URL from Terminal 2 in browser"
echo ""
