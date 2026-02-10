#!/bin/bash
# Fix Kafka connectivity issues
# Restarts Kafka and Zookeeper, then verifies connectivity

set -e

NAMESPACE="metalmart"

echo "🔧 Fixing Kafka connectivity..."
echo ""

# Check cluster connection
echo "1. Checking cluster connection..."
kubectl cluster-info &>/dev/null || {
    echo "❌ Not connected to cluster"
    exit 1
}
echo "✅ Connected to cluster"
echo ""

# Restart Zookeeper first (Kafka depends on it)
echo "2. Restarting Zookeeper..."
kubectl rollout restart deployment/zookeeper -n $NAMESPACE
echo "⏳ Waiting for Zookeeper to restart..."
kubectl rollout status deployment/zookeeper -n $NAMESPACE --timeout=120s || echo "⚠️  Zookeeper restart taking longer"
echo "✅ Zookeeper restarted"
echo ""

# Wait for Zookeeper to be ready
echo "3. Waiting for Zookeeper to be ready..."
kubectl wait --for=condition=ready pod -l app=zookeeper -n $NAMESPACE --timeout=120s || {
    echo "⚠️  Zookeeper not ready, but continuing..."
}
echo ""

# Restart Kafka
echo "4. Restarting Kafka..."
kubectl rollout restart deployment/kafka -n $NAMESPACE
echo "⏳ Waiting for Kafka to restart..."
kubectl rollout status deployment/kafka -n $NAMESPACE --timeout=120s || echo "⚠️  Kafka restart taking longer"
echo "✅ Kafka restarted"
echo ""

# Wait for Kafka to be ready
echo "5. Waiting for Kafka to be ready..."
sleep 15
KAFKA_READY=$(kubectl get pod -n $NAMESPACE -l app=kafka -o jsonpath='{.items[0].status.containerStatuses[0].ready}' 2>/dev/null || echo "false")
if [ "$KAFKA_READY" != "true" ]; then
    echo "⚠️  Kafka not ready yet, waiting 30 more seconds..."
    sleep 30
fi
echo ""

# Check Kafka logs
echo "6. Checking Kafka logs..."
kubectl logs -n $NAMESPACE -l app=kafka --tail=10 2>&1 | tail -10
echo ""

# Restart order and order-processor to reconnect
echo "7. Restarting order services to reconnect to Kafka..."
kubectl rollout restart deployment/order -n $NAMESPACE
kubectl rollout restart deployment/order-processor -n $NAMESPACE
echo "⏳ Waiting 20 seconds for services to reconnect..."
sleep 20
echo ""

# Check order service logs
echo "8. Checking order service connection..."
kubectl logs -n $NAMESPACE -l app=order --tail=5 2>&1 | grep -i kafka | tail -3 || echo "No Kafka connection messages in recent logs"
echo ""

# Check order-processor logs
echo "9. Checking order-processor connection..."
kubectl logs -n $NAMESPACE -l app=order-processor --tail=5 2>&1 | tail -5
echo ""

echo "✅ Kafka fix complete!"
echo ""
echo "📝 Next steps:"
echo "   - Create a new order to test"
echo "   - Check if order status updates work"
echo "   - Monitor logs: kubectl logs -n $NAMESPACE -l app=order-processor -f"
