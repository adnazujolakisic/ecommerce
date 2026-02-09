#!/bin/bash
set -e

# GKE Deployment Script
# Prerequisites:
# 1. gcloud CLI installed and authenticated
# 2. kubectl configured for your GKE cluster
# 3. Docker configured to push to GCR
# 4. Set GCP_PROJECT_ID environment variable

if [ -z "$GCP_PROJECT_ID" ]; then
  echo "Error: GCP_PROJECT_ID environment variable not set"
  echo "Usage: export GCP_PROJECT_ID=your-project-id && ./scripts/deploy-gke.sh"
  exit 1
fi

echo "🚀 Deploying MetalMart to GKE..."
echo "Project ID: $GCP_PROJECT_ID"

# 1. Configure Docker for GCR
echo "1. Configuring Docker for GCR..."
gcloud auth configure-docker --quiet

# 2. Build and push images
echo "2. Building Docker images..."
docker-compose build

echo "3. Tagging and pushing images to GCR..."
# Tag and push each service
docker tag metalmart-catalogue:latest gcr.io/$GCP_PROJECT_ID/metalmart/catalogue:latest
docker tag metalmart-inventory:latest gcr.io/$GCP_PROJECT_ID/metalmart/inventory:latest
docker tag metalmart-checkout:latest gcr.io/$GCP_PROJECT_ID/metalmart/checkout:latest
docker tag metalmart-order:latest gcr.io/$GCP_PROJECT_ID/metalmart/order:latest
docker tag metalmart-order-processor:latest gcr.io/$GCP_PROJECT_ID/metalmart/order-processor:latest
docker tag metalmart-frontend:latest gcr.io/$GCP_PROJECT_ID/metalmart/frontend:latest

docker push gcr.io/$GCP_PROJECT_ID/metalmart/catalogue:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/inventory:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/checkout:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/order:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/order-processor:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/frontend:latest

# 3. Update GKE kustomization with project ID
echo "4. Updating GKE kustomization with project ID..."
sed -i.bak "s/PROJECT_ID/$GCP_PROJECT_ID/g" k8s/overlays/gke/kustomization.yaml

# 4. Deploy infrastructure first
echo "5. Deploying infrastructure (secrets, postgres, kafka)..."
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/infrastructure/secrets.yaml
kubectl apply -f k8s/base/infrastructure/postgres.yaml
kubectl apply -f k8s/base/infrastructure/kafka.yaml

# Wait for infrastructure to be ready
echo "   Waiting for infrastructure to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n metalmart --timeout=180s || true
kubectl wait --for=condition=ready pod -l app=zookeeper -n metalmart --timeout=180s || true
kubectl wait --for=condition=ready pod -l app=kafka -n metalmart --timeout=180s || true

# 5. Deploy application services using GKE overlay
echo "6. Deploying application services..."
kubectl apply -k k8s/overlays/gke

# 6. Apply Mirrord resources (optional)
echo "7. Applying Mirrord Kafka resources (if operator installed)..."
if kubectl get crd mirrordkafkaclientconfigs.queues.mirrord.metalbear.co &>/dev/null; then
  kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml
  echo "   ✅ Mirrord Kafka resources applied"
else
  echo "   ⚠️  Mirrord operator not installed - skipping Kafka resources"
fi

# 7. Enable demo mode
echo "8. Enabling demo mode..."
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true

# 8. Wait for pods
echo ""
echo "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=catalogue -n metalmart --timeout=120s || true
kubectl wait --for=condition=ready pod -l app=inventory -n metalmart --timeout=120s || true
kubectl wait --for=condition=ready pod -l app=frontend -n metalmart --timeout=120s || true

# 9. Seed inventory
echo ""
echo "9. Seeding inventory..."
CATALOGUE_POD=$(kubectl get pod -n metalmart -l app=catalogue -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -n "$CATALOGUE_POD" ]; then
  echo "   Waiting for services to be ready..."
  sleep 10
  
  kubectl exec -n metalmart $CATALOGUE_POD -- /bin/sh /app/seed-inventory.sh 2>/dev/null && \
    echo "   ✅ Inventory seeded successfully" || \
    echo "   ⚠️  Warning: Inventory seeding may have failed (this is OK if already seeded)"
else
  echo "   ⚠️  Warning: Catalogue pod not found, skipping inventory seeding"
fi

# 10. Get frontend LoadBalancer IP
echo ""
echo "✅ Setup complete!"
echo ""
echo "Getting frontend LoadBalancer IP..."
kubectl get svc frontend -n metalmart

echo ""
echo "📊 Check status:"
echo "   kubectl get pods -n metalmart"
echo ""
echo "🌐 Access frontend (wait for EXTERNAL-IP to be assigned):"
echo "   kubectl get svc frontend -n metalmart"
