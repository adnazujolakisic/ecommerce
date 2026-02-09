#!/bin/bash

# Quick fix script for when order status stops updating
# This usually means the order service lost connection to Kafka

set -e

echo "🔍 Diagnosing order status update issue..."
echo ""

# Check if order service can connect to Kafka
echo "1. Checking order service logs for Kafka errors..."
ORDER_KAFKA_ERROR=$(kubectl logs -n metalmart deployment/order --tail=50 | grep -i "kafka.*fail\|failed.*kafka" || echo "")

if [ -n "$ORDER_KAFKA_ERROR" ]; then
    echo "   ⚠️  Found Kafka connection errors in order service"
    echo "   Error: $ORDER_KAFKA_ERROR"
    echo ""
    echo "2. Restarting order service to reconnect to Kafka..."
    kubectl rollout restart deployment/order -n metalmart
    echo "   ✅ Order service restarted"
    echo ""
    echo "3. Waiting for rollout..."
    kubectl rollout status deployment/order -n metalmart --timeout=30s
    echo ""
    echo "4. Checking order-processor is running..."
    ORDER_PROCESSOR_STATUS=$(kubectl get pods -n metalmart -l app=order-processor -o jsonpath='{.items[0].status.phase}' 2>/dev/null || echo "NotFound")
    if [ "$ORDER_PROCESSOR_STATUS" != "Running" ]; then
        echo "   ⚠️  Order processor is not running! Restarting..."
        kubectl rollout restart deployment/order-processor -n metalmart
        kubectl rollout status deployment/order-processor -n metalmart --timeout=30s
    else
        echo "   ✅ Order processor is running"
    fi
    echo ""
    echo "✅ Fix applied! Try placing a new order and check if status updates."
else
    echo "   ✅ No Kafka errors found in order service"
    echo ""
    echo "2. Checking if order-processor is receiving messages..."
    ORDER_PROCESSOR_LOGS=$(kubectl logs -n metalmart deployment/order-processor --tail=20 2>/dev/null | grep -i "received\|processing" || echo "")
    if [ -z "$ORDER_PROCESSOR_LOGS" ]; then
        echo "   ⚠️  Order processor not receiving messages"
        echo "   Restarting order-processor..."
        kubectl rollout restart deployment/order-processor -n metalmart
        kubectl rollout status deployment/order-processor -n metalmart --timeout=30s
        echo "   ✅ Order processor restarted"
    else
        echo "   ✅ Order processor is active"
        echo "   Recent activity: $ORDER_PROCESSOR_LOGS"
    fi
fi

echo ""
echo "📋 Quick test commands:"
echo "  # Watch order service logs:"
echo "  kubectl logs -n metalmart deployment/order -f"
echo ""
echo "  # Watch order processor logs:"
echo "  kubectl logs -n metalmart deployment/order-processor -f"
echo ""
echo "  # Check all pods status:"
echo "  kubectl get pods -n metalmart | grep -E 'order|kafka'"
