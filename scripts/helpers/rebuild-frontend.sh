#!/bin/bash

# Quick script to rebuild and restart frontend in Kubernetes

set -e

echo "🔨 Rebuilding frontend image..."

# Check if we're using Minikube
if kubectl config current-context | grep -q minikube; then
  echo "   Detected Minikube - building and loading image..."
  docker build -t metalmart/frontend:latest frontend/
  minikube image load metalmart/frontend:latest
else
  echo "   Building image..."
  docker build -t metalmart/frontend:latest frontend/
fi

echo ""
echo "🔄 Restarting frontend deployment..."
kubectl rollout restart deployment/frontend -n metalmart

echo ""
echo "⏳ Waiting for rollout to complete..."
kubectl rollout status deployment/frontend -n metalmart --timeout=60s

echo ""
echo "✅ Frontend updated successfully!"
