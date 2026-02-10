#!/bin/bash

# Quick script to rebuild and restart frontend in Kubernetes
# This script:
# 1. Builds the frontend with npm
# 2. Builds the Docker image
# 3. Restarts the frontend deployment

set -e

echo "🔨 Rebuilding frontend..."

# Step 1: Build frontend with npm
echo "1. Building frontend (npm run build)..."
cd frontend
npm run build
cd ..

# Step 2: Build Docker image
echo ""
echo "2. Building Docker image..."

# Check if we're using Minikube
if kubectl config current-context | grep -q minikube; then
  echo "   Detected Minikube - setting docker-env..."
  eval $(minikube docker-env)
  docker-compose build frontend
  
  # Tag for Kubernetes
  PROJECT_NAME=$(basename $(pwd) | tr '[:upper:]' '[:lower:]')
  if docker image inspect "${PROJECT_NAME}-frontend:latest" &>/dev/null; then
    docker tag "${PROJECT_NAME}-frontend:latest" "metalmart/frontend:latest"
    echo "   Tagged ${PROJECT_NAME}-frontend -> metalmart/frontend"
  elif docker image inspect "metalmart-frontend:latest" &>/dev/null; then
    docker tag "metalmart-frontend:latest" "metalmart/frontend:latest"
    echo "   Tagged metalmart-frontend -> metalmart/frontend"
  fi
else
  echo "   Building for non-Minikube environment..."
  docker-compose build frontend
fi

# Step 3: Restart deployment
echo ""
echo "3. Restarting frontend deployment..."
kubectl rollout restart deployment/frontend -n metalmart

# Step 4: Wait for rollout
echo ""
echo "4. Waiting for rollout to complete..."
kubectl rollout status deployment/frontend -n metalmart --timeout=120s || echo "   ⚠️  Rollout status check timed out, but deployment may still be in progress"

# Step 5: Verify pod is running
echo ""
echo "5. Verifying frontend pod..."
sleep 5
kubectl get pods -n metalmart -l app=frontend

echo ""
echo "✅ Frontend rebuild complete!"
echo ""
echo "💡 If you don't see changes, try:"
echo "   - Hard refresh your browser (Cmd+Shift+R / Ctrl+Shift+R)"
echo "   - Check pod logs: kubectl logs -n metalmart -l app=frontend --tail=20"
