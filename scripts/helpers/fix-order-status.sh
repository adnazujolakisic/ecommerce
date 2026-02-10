#!/bin/bash
# Quick fix for order status updates not working
# This restarts order and order-processor services to reconnect to Kafka

set -e

NAMESPACE="metalmart"

echo "🔧 Fixing order status updates..."
echo ""

# Check if we're connected to the right cluster
echo "1. Checking cluster connection..."
kubectl cluster-info &>/dev/null || {
    echo "❌ Not connected to cluster. Run: kubectl get nodes"
    exit 1
}

echo "✅ Connected to cluster"
echo ""

# Check Kafka status
echo "2. Checking Kafka status..."
KAFKA_READY=$(kubectl get pod -n $NAMESPACE -l app=kafka -o jsonpath='{.items[0].status.containerStatuses[0].ready}' 2>/dev/null || echo "false")
if [ "$KAFKA_READY" != "true" ]; then
    echo "⚠️  Kafka is not ready. Waiting 10 seconds..."
    sleep 10
    KAFKA_READY=$(kubectl get pod -n $NAMESPACE -l app=kafka -o jsonpath='{.items[0].status.containerStatuses[0].ready}' 2>/dev/null || echo "false")
    if [ "$KAFKA_READY" != "true" ]; then
        echo "❌ Kafka is still not ready. Check: kubectl get pods -n $NAMESPACE | grep kafka"
        exit 1
    fi
fi
echo "✅ Kafka is ready"
echo ""

# Restart order service to reconnect to Kafka
echo "3. Restarting order service to reconnect to Kafka..."
kubectl rollout restart deployment/order -n $NAMESPACE
echo "⏳ Waiting for order service to restart..."
kubectl rollout status deployment/order -n $NAMESPACE --timeout=60s || echo "⚠️  Order service restart taking longer than expected"
echo "✅ Order service restarted"
echo ""

# Restart order-processor to reconnect to Kafka
echo "4. Restarting order-processor to reconnect to Kafka..."
kubectl rollout restart deployment/order-processor -n $NAMESPACE
echo "⏳ Waiting for order-processor to restart..."
kubectl rollout status deployment/order-processor -n $NAMESPACE --timeout=60s || echo "⚠️  Order-processor restart taking longer than expected"
echo "✅ Order-processor restarted"
echo ""

# Wait a bit for services to connect
echo "5. Waiting for services to connect to Kafka..."
sleep 10

# Check logs
echo ""
echo "6. Checking order service logs (last 5 lines)..."
kubectl logs -n $NAMESPACE -l app=order --tail=5 2>&1 | tail -5
echo ""

echo "7. Checking order-processor logs (last 5 lines)..."
kubectl logs -n $NAMESPACE -l app=order-processor --tail=5 2>&1 | tail -5
echo ""

echo "✅ Done! Order status updates should work now."
echo ""
echo "📝 Next steps:"
echo "   - Create a new order to test"
echo "   - Check order status page - it should update every 2 seconds"
echo "   - If still not working, check logs: kubectl logs -n $NAMESPACE -l app=order-processor"
