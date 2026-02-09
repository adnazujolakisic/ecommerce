#!/bin/bash
set -e

echo " Starting MetalMart from scratch..."

# Check if Docker is running
echo "0. Checking Docker..."
if ! docker info >/dev/null 2>&1; then
  echo "   ❌ Docker is not running!"
  echo "   Please start Docker Desktop and try again."
  exit 1
fi
echo "   ✅ Docker is running"

# 1. Start Minikube
echo "1. Starting Minikube..."
minikube start
eval $(minikube docker-env)

# 2. Build images
echo "2. Building Docker images..."
docker-compose build

# 3. Tag images
echo "3. Tagging images for K8s..."
docker tag metalmart-catalogue:latest metalmart/catalogue:latest
docker tag metalmart-inventory:latest metalmart/inventory:latest
docker tag metalmart-checkout:latest metalmart/checkout:latest
docker tag metalmart-order:latest metalmart/order:latest
docker tag metalmart-order-processor:latest metalmart/order-processor:latest
docker tag metalmart-frontend:latest metalmart/frontend:latest

# 4. Deploy infrastructure first
echo "4. Deploying infrastructure (secrets, postgres, kafka)..."
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/infrastructure/secrets.yaml
kubectl apply -f k8s/base/infrastructure/postgres.yaml
kubectl apply -f k8s/base/infrastructure/kafka.yaml

# Fix image pull policy for infrastructure (they need to pull from Docker Hub)
echo "   Setting image pull policy for infrastructure..."
kubectl patch deployment postgres -n metalmart -p '{"spec":{"template":{"spec":{"containers":[{"name":"postgres","imagePullPolicy":"IfNotPresent"}]}}}}' 2>/dev/null || true
kubectl patch deployment kafka -n metalmart -p '{"spec":{"template":{"spec":{"containers":[{"name":"kafka","imagePullPolicy":"IfNotPresent"}]}}}}' 2>/dev/null || true
kubectl patch deployment zookeeper -n metalmart -p '{"spec":{"template":{"spec":{"containers":[{"name":"zookeeper","imagePullPolicy":"IfNotPresent"}]}}}}' 2>/dev/null || true

# Wait for infrastructure to be ready
echo "   Waiting for infrastructure to be ready..."
kubectl wait --for=condition=ready pod -l app=postgres -n metalmart --timeout=180s || true
kubectl wait --for=condition=ready pod -l app=zookeeper -n metalmart --timeout=180s || true
kubectl wait --for=condition=ready pod -l app=kafka -n metalmart --timeout=180s || true

# 5. Deploy application services
echo "5. Deploying application services..."
kubectl apply -f k8s/base/catalogue/deployment.yaml
kubectl apply -f k8s/base/inventory/deployment.yaml
kubectl apply -f k8s/base/checkout/deployment.yaml
kubectl apply -f k8s/base/order/deployment.yaml
kubectl apply -f k8s/base/order-processor/deployment.yaml
kubectl apply -f k8s/base/frontend/deployment.yaml

# Set image pull policy for local images
echo "   Setting image pull policy for local images..."
for svc in catalogue inventory checkout order order-processor frontend; do
  kubectl patch deployment $svc -n metalmart -p '{"spec":{"template":{"spec":{"containers":[{"name":"'$svc'","imagePullPolicy":"Never"}]}}}}' 2>/dev/null || true
done

# 6. Apply Mirrord resources (optional - only if operator is installed)
echo "6. Applying Mirrord Kafka resources (if operator installed)..."
if kubectl get crd mirrordkafkaclientconfigs.queues.mirrord.metalbear.co &>/dev/null; then
  kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml
  echo "    Mirrord Kafka resources applied"
else
  echo "     Mirrord operator not installed - skipping Kafka resources"
  echo "     To install: helm install mirrord-operator mirrord/mirrord-operator --set license.key=<KEY> --set operator.kafkaSplitting=true -n mirrord --create-namespace"
fi

# 7. Enable demo mode
echo "7. Enabling demo mode..."
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true

echo ""
echo "Waiting for pods to be ready..."
kubectl wait --for=condition=ready pod -l app=catalogue -n metalmart --timeout=120s || true
kubectl wait --for=condition=ready pod -l app=inventory -n metalmart --timeout=120s || true
kubectl wait --for=condition=ready pod -l app=frontend -n metalmart --timeout=120s || true

# 8. Seed inventory
echo ""
echo "8. Seeding inventory..."
CATALOGUE_POD=$(kubectl get pod -n metalmart -l app=catalogue -o jsonpath='{.items[0].metadata.name}' 2>/dev/null)
if [ -n "$CATALOGUE_POD" ]; then
  echo "   Waiting for services to be ready..."
  sleep 10
  
  # Use the seed-inventory.sh script from the catalogue pod
  kubectl exec -n metalmart $CATALOGUE_POD -- /bin/sh /app/seed-inventory.sh 2>/dev/null && \
    echo "   Inventory seeded successfully" || \
    echo "   Warning: Inventory seeding may have failed (this is OK if already seeded)"
else
  echo "   Warning: Catalogue pod not found, skipping inventory seeding"
fi

echo ""
echo " Setup complete!"
echo ""
echo " Access frontend:"
echo "   minikube service frontend -n metalmart"
echo ""
echo " Check status:"
echo "   kubectl get pods -n metalmart"
