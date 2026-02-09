#!/bin/bash

# Quick script to rebuild and restart frontend in Kubernetes

set -e

echo "🔨 Rebuilding frontend image..."
docker build -t metalmart/frontend:latest frontend/

echo ""
echo "🔄 Restarting frontend deployment..."
kubectl rollout restart deployment/frontend -n metalmart

echo ""
echo "⏳ Waiting for rollout to complete..."
kubectl rollout status deployment/frontend -n metalmart --timeout=60s

echo ""
echo "✅ Frontend updated successfully!"
