#!/bin/bash
set -e

echo "🚀 Starting MetalMart from scratch..."
echo ""

# Prerequisites check
echo "📋 Checking prerequisites..."

# Check Docker
if ! command -v docker &> /dev/null; then
  echo "   ❌ Docker is not installed!"
  echo "   Install from: https://www.docker.com/products/docker-desktop"
  exit 1
fi
if ! docker info >/dev/null 2>&1; then
  echo "   ❌ Docker is not running!"
  echo "   Please start Docker Desktop and try again."
  exit 1
fi
echo "   ✅ Docker is installed and running"

# Check docker-compose
if ! command -v docker-compose &> /dev/null; then
  echo "   ❌ docker-compose is not installed!"
  echo "   Install from: https://docs.docker.com/compose/install/"
  exit 1
fi
echo "   ✅ docker-compose is installed"

# Check Minikube
if ! command -v minikube &> /dev/null; then
  echo "   ❌ Minikube is not installed!"
  echo "   Install with: brew install minikube"
  echo "   Or from: https://minikube.sigs.k8s.io/docs/start/"
  exit 1
fi
echo "   ✅ Minikube is installed"

# Check kubectl
if ! command -v kubectl &> /dev/null; then
  echo "   ❌ kubectl is not installed!"
  echo "   Install with: brew install kubectl"
  echo "   Or from: https://kubernetes.io/docs/tasks/tools/"
  exit 1
fi
echo "   ✅ kubectl is installed"

echo ""
echo "✅ All prerequisites met!"
echo ""

# 1. Start Minikube
echo "1. Starting Minikube..."
# Check if Minikube is actually working by verifying kubectl can connect
if minikube status &>/dev/null && kubectl cluster-info &>/dev/null 2>&1; then
  echo "   Minikube is already running and healthy"
else
  # If Minikube reports as running but kubectl can't connect, it's in a broken state
  if minikube status &>/dev/null; then
    echo "   ⚠️  Minikube appears to be in an inconsistent state"
    echo "   Skipping cleanup - minikube start will handle it"
    # Don't try to delete/stop as it can hang - let minikube start handle it
  fi
  echo "   Starting Minikube (this may take a minute)..."
  # minikube start should detect and fix broken states automatically
  minikube start
fi

# Verify Minikube is actually working before setting docker-env
if ! kubectl cluster-info &>/dev/null 2>&1; then
  echo "   ❌ Failed to connect to Minikube cluster"
  echo "   Try running: minikube delete && minikube start"
  exit 1
fi

eval $(minikube docker-env)
echo "   ✅ Minikube is ready"

# 1.5. Pre-pull infrastructure images to avoid slow pulls during deployment
echo "1.5. Pre-pulling infrastructure images..."
echo "   Pulling postgres:15-alpine..."
docker pull postgres:15-alpine 2>/dev/null || echo "   (using cached image)"
echo "   Pulling confluentinc/cp-zookeeper:7.5.0..."
docker pull confluentinc/cp-zookeeper:7.5.0 2>/dev/null || echo "   (using cached image)"
echo "   Pulling confluentinc/cp-kafka:7.5.0..."
docker pull confluentinc/cp-kafka:7.5.0 2>/dev/null || echo "   (using cached image)"
echo "   ✅ Infrastructure images ready"

# 2. Build images
echo "2. Building Docker images..."
docker-compose build

# 3. Tag images
echo "3. Tagging images for K8s..."
# Docker Compose names images as <directory-name>-<service-name>
# Get the project name (directory name) from docker-compose
PROJECT_NAME=$(basename $(pwd) | tr '[:upper:]' '[:lower:]')

# Tag images - try both ecommerce and metalmart prefixes for compatibility
for service in catalogue inventory checkout order order-processor frontend; do
  # Try ecommerce prefix first (current directory name)
  if docker image inspect "${PROJECT_NAME}-${service}:latest" &>/dev/null; then
    docker tag "${PROJECT_NAME}-${service}:latest" "metalmart/${service}:latest"
    echo "   Tagged ${PROJECT_NAME}-${service} -> metalmart/${service}"
  # Fallback to metalmart prefix (old naming)
  elif docker image inspect "metalmart-${service}:latest" &>/dev/null; then
    docker tag "metalmart-${service}:latest" "metalmart/${service}:latest"
    echo "   Tagged metalmart-${service} -> metalmart/${service}"
  else
    echo "   ⚠️  Warning: Image for ${service} not found"
  fi
done

# 4. Deploy infrastructure first
echo "4. Deploying infrastructure (secrets, postgres, kafka)..."
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/infrastructure/secrets.yaml
kubectl apply -f k8s/base/infrastructure/postgres.yaml
kubectl apply -f k8s/base/infrastructure/kafka.yaml

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
echo "✅ Setup complete!"
echo ""

# Final verification
echo "🔍 Verifying deployment..."
READY_PODS=$(kubectl get pods -n metalmart --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l | tr -d ' ')
TOTAL_PODS=$(kubectl get pods -n metalmart --no-headers 2>/dev/null | wc -l | tr -d ' ')

if [ "$READY_PODS" -gt 0 ]; then
  echo "   ✅ $READY_PODS/$TOTAL_PODS pods are running"
else
  echo "   ⚠️  No pods are running yet. Check status with: kubectl get pods -n metalmart"
fi

echo ""
echo "📝 Next steps:"
echo ""
echo "   Access frontend:"
echo "     minikube service frontend -n metalmart"
echo ""
echo "   Check pod status:"
echo "     kubectl get pods -n metalmart"
echo ""
echo "   View logs:"
echo "     kubectl logs -n metalmart deployment/<service-name> -f"
echo ""
echo "   Troubleshooting:"
echo "     kubectl describe pod <pod-name> -n metalmart"
echo ""
