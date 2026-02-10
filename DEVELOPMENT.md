# MetalMart Development Guide

This is my development guide for MetalMart - a microservices e-commerce app I use for demonstrating Mirrord features. I keep this updated with all the commands I actually use.

## Quick Reference

**I use these most often:**
- `./scripts/helpers/start-fresh.sh` - Complete setup from scratch
- `./scripts/helpers/rebuild-frontend.sh` - Rebuild frontend after changes
- `./scripts/helpers/fix-order-status.sh` - Fix when order status stops updating
- `kubectl get pods -n metalmart` - Check what's running
- `kubectl logs -n metalmart deployment/order-processor -f` - Watch logs

---

## Project Overview

- Kafka queue splitting
- Database branching
- Filtering
- Steal mode
- Mirror mode

### Services (Go microservices)
- **catalogue** (8081) - Product catalog, seeded from DB
- **inventory** (8082) - Stock management
- **checkout** (8083) - Cart processing
- **order** (8084) - Order creation, publishes to Kafka
- **order-processor** - Kafka consumer, processes orders async

### Infrastructure
- **PostgreSQL** - Databases: catalogue, inventory, orders
- **Kafka + Zookeeper** - Async messaging for order processing
- **nginx** - Frontend reverse proxy

### Frontend
- React + Vite on port 3000 (Docker) or 80 (K8s)

---

## Local Development (Docker Compose)

I use Docker Compose for quick local testing when I don't need Kubernetes.

### Start Everything
```bash
docker-compose up --build -d
```

### Check Status
```bash
docker-compose ps
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f order-processor
```

### Stop Everything
```bash
docker-compose down
```

### URLs
- Frontend: http://localhost:3000
- Catalogue API: http://localhost:8081
- Inventory API: http://localhost:8082
- Checkout API: http://localhost:8083
- Order API: http://localhost:8084

---

## Minikube Deployment (Manual Steps)

I usually use `./scripts/helpers/start-fresh.sh` for this, but here are the manual steps if I need to do it step-by-step or troubleshoot.

### Step 1: Check Docker is Running
```bash
docker info
# If it fails, start Docker Desktop
```

### Step 2: Start Minikube
```bash
minikube start
eval $(minikube docker-env)
```

### Step 3: Build All Images
```bash
docker-compose build
```

### Step 4: Tag Images for Kubernetes
I need to tag them with the `metalmart/` prefix for K8s:
```bash
docker tag metalmart-catalogue:latest metalmart/catalogue:latest
docker tag metalmart-inventory:latest metalmart/inventory:latest
docker tag metalmart-checkout:latest metalmart/checkout:latest
docker tag metalmart-order:latest metalmart/order:latest
docker tag metalmart-order-processor:latest metalmart/order-processor:latest
docker tag metalmart-frontend:latest metalmart/frontend:latest
```

### Step 5: Deploy Infrastructure First
I deploy infrastructure before apps so they're ready:
```bash
# Create namespace
kubectl apply -f k8s/base/namespace.yaml

# Deploy secrets
kubectl apply -f k8s/base/infrastructure/secrets.yaml

# Deploy PostgreSQL
kubectl apply -f k8s/base/infrastructure/postgres.yaml

# Deploy Kafka and Zookeeper
kubectl apply -f k8s/base/infrastructure/kafka.yaml
```

### Step 6: Set Image Pull Policy for Infrastructure
Infrastructure images come from Docker Hub, so I set `IfNotPresent`:
```bash
kubectl patch deployment postgres -n metalmart -p '{"spec":{"template":{"spec":{"containers":[{"name":"postgres","imagePullPolicy":"IfNotPresent"}]}}}}'
kubectl patch deployment kafka -n metalmart -p '{"spec":{"template":{"spec":{"containers":[{"name":"kafka","imagePullPolicy":"IfNotPresent"}]}}}}'
kubectl patch deployment zookeeper -n metalmart -p '{"spec":{"template":{"spec":{"containers":[{"name":"zookeeper","imagePullPolicy":"IfNotPresent"}]}}}}'
```

### Step 7: Wait for Infrastructure to be Ready
I wait for infrastructure before deploying apps:
```bash
kubectl wait --for=condition=ready pod -l app=postgres -n metalmart --timeout=180s
kubectl wait --for=condition=ready pod -l app=zookeeper -n metalmart --timeout=180s
kubectl wait --for=condition=ready pod -l app=kafka -n metalmart --timeout=180s
```

### Step 8: Deploy Application Services
Now I deploy the app services:
```bash
kubectl apply -f k8s/base/catalogue/deployment.yaml
kubectl apply -f k8s/base/inventory/deployment.yaml
kubectl apply -f k8s/base/checkout/deployment.yaml
kubectl apply -f k8s/base/order/deployment.yaml
kubectl apply -f k8s/base/order-processor/deployment.yaml
kubectl apply -f k8s/base/frontend/deployment.yaml
```

### Step 9: Set Image Pull Policy for Local Images
I use `Never` for local images so K8s uses the ones I just built:
```bash
for svc in catalogue inventory checkout order order-processor frontend; do
  kubectl patch deployment $svc -n metalmart -p '{"spec":{"template":{"spec":{"containers":[{"name":"'$svc'","imagePullPolicy":"Never"}]}}}}'
done
```

### Step 10: Apply Mirrord Resources (Optional)
If I have Mirrord operator installed, I apply the Kafka resources:
```bash
# Check if operator is installed
kubectl get crd mirrordkafkaclientconfigs.queues.mirrord.metalbear.co

# If it exists, apply Kafka resources
kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml
```

### Step 11: Enable Demo Mode
I enable demo mode for faster order processing (200ms per step instead of 2-5 seconds):
```bash
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true
```

### Step 12: Wait for Pods to be Ready
I wait for key services to be ready:
```bash
kubectl wait --for=condition=ready pod -l app=catalogue -n metalmart --timeout=120s
kubectl wait --for=condition=ready pod -l app=inventory -n metalmart --timeout=120s
kubectl wait --for=condition=ready pod -l app=frontend -n metalmart --timeout=120s
```

### Step 13: Seed Inventory
Products are auto-seeded, but I need to seed inventory manually:
```bash
# Get the catalogue pod name
CATALOGUE_POD=$(kubectl get pod -n metalmart -l app=catalogue -o jsonpath='{.items[0].metadata.name}')

# Wait a bit for services to be ready
sleep 10

# Run the seed script inside the pod
kubectl exec -n metalmart $CATALOGUE_POD -- /bin/sh /app/seed-inventory.sh
```

### Step 14: Access Frontend
```bash
minikube service frontend -n metalmart
```

---

## Common Operations

### Check Pod Status
```bash
# All pods
kubectl get pods -n metalmart

# Watch for changes
kubectl get pods -n metalmart -w

# Filter by service
kubectl get pods -n metalmart -l app=order
```

### View Logs
```bash
# Specific deployment
kubectl logs -n metalmart deployment/order-processor -f

# By label
kubectl logs -n metalmart -l app=order-processor -f

# Last 50 lines
kubectl logs -n metalmart deployment/order --tail=50
```

### Restart a Service
```bash
kubectl rollout restart deployment/order -n metalmart

# Wait for restart
kubectl rollout status deployment/order -n metalmart
```

### Rebuild Frontend (Manual)
When I make frontend changes, I rebuild manually:
```bash
# Build new image
docker build -t metalmart/frontend:latest frontend/

# Restart deployment
kubectl rollout restart deployment/frontend -n metalmart

# Wait for rollout
kubectl rollout status deployment/frontend -n metalmart --timeout=60s
```

### Fix Order Status Issues (Manual)
When order status stops updating, I check and fix manually:
```bash
# 1. Check for Kafka errors in order service
kubectl logs -n metalmart deployment/order --tail=50 | grep -i "kafka.*fail\|failed.*kafka"

# 2. If errors found, restart order service
kubectl rollout restart deployment/order -n metalmart
kubectl rollout status deployment/order -n metalmart --timeout=30s

# 3. Check order-processor is running
kubectl get pods -n metalmart -l app=order-processor

# 4. If not running, restart it
kubectl rollout restart deployment/order-processor -n metalmart
kubectl rollout status deployment/order-processor -n metalmart --timeout=30s

# 5. Watch logs to verify it's working
kubectl logs -n metalmart deployment/order-processor -f
```

### Delete Everything
When I want to start completely fresh:
```bash
kubectl delete namespace metalmart
```

---

## Mirrord Development

I use Mirrord to debug services locally while connected to the K8s cluster's infrastructure (Kafka, Postgres, other services).

### Available Configs
I have Mirrord configs for each service:
- `.mirrord/order-processor.json` - For Kafka queue splitting demo
- `.mirrord/inventory-steal.json` - For steal mode demo
- `.mirrord/catalogue.json`
- `.mirrord/inventory.json`
- `.mirrord/checkout.json`
- `.mirrord/order.json`

### Basic Usage
```bash
# Run order service locally against cluster
cd services/order
mirrord exec -f ../../.mirrord/order.json -- go run .

# Or specify target inline
mirrord exec --target deployment/catalogue -n metalmart -- go run ./services/catalogue
```

### Mirrord Modes
- **mirror** (default) - Copy traffic to local, pod still receives it
- **steal** - Intercept all traffic, pod doesn't receive it

To use steal mode, I edit the config:
```json
"incoming": "steal"
```

### Known Issue: Kafka + Docker Desktop
Mirrord's outgoing network has issues with Docker Desktop on macOS. My workaround:
```bash
# Terminal 1: Port-forward Kafka
kubectl port-forward -n metalmart svc/kafka 9092:9092

# Terminal 2: Run with localhost
cd services/order-processor
KAFKA_BROKERS=localhost:9092 go run .
```

---

## Testing Kafka

### List Topics
```bash
kubectl exec -n metalmart deploy/kafka -- kafka-topics --bootstrap-server localhost:9092 --list
```

### Produce Test Message
```bash
kubectl exec -n metalmart -it deploy/kafka -- \
  kafka-console-producer --bootstrap-server localhost:9092 --topic order.created
```

### Consume Messages
```bash
kubectl exec -n metalmart -it deploy/kafka -- \
  kafka-console-consumer --bootstrap-server localhost:9092 --topic order.created --from-beginning
```

---

## Project Structure

```
metalmart/
├── services/
│   ├── catalogue/       # Product catalog service
│   ├── inventory/       # Stock management
│   ├── checkout/        # Cart processing
│   ├── order/           # Order creation + Kafka producer
│   └── order-processor/ # Kafka consumer
├── frontend/            # React + Vite
├── k8s/
│   ├── base/
│   │   ├── infrastructure/
│   │   │   ├── postgres.yaml
│   │   │   ├── kafka.yaml
│   │   │   └── secrets.yaml
│   │   ├── catalogue/
│   │   ├── inventory/
│   │   ├── checkout/
│   │   ├── order/
│   │   ├── order-processor/
│   │   ├── frontend/
│   │   └── kustomization.yaml
│   └── overlays/
│       ├── minikube/    # Local K8s
│       └── gke/         # Google Cloud
├── scripts/
│   ├── helpers/         # Helper scripts (start-fresh.sh, etc.)
│   └── load-generator.* # Load testing tools
├── .mirrord/            # Mirrord configs per service
├── docker-compose.yml
└── nginx.conf
```

---

## Secrets

### K8s Secrets (k8s/base/infrastructure/secrets.yaml)
- `db-secrets` - PostgreSQL connection strings

### Environment Variables
| Service | Key Variables |
|---------|--------------|
| catalogue | DATABASE_URL, SEED_DATA |
| inventory | DATABASE_URL |
| order | DATABASE_URL, KAFKA_BROKERS |
| checkout | INVENTORY_SERVICE_URL, ORDER_SERVICE_URL |
| order-processor | KAFKA_BROKERS, ORDER_SERVICE_URL, DEMO_MODE |

---

## Helper Scripts

I have helper scripts in `scripts/helpers/` that automate common tasks:

- `start-fresh.sh` - Complete setup from scratch (does all the manual steps above)
- `rebuild-frontend.sh` - Quick rebuild and restart of frontend
- `fix-order-status.sh` - Diagnose and fix order status update issues
- `deploy-gke.sh` - Deploy to Google Kubernetes Engine

See `scripts/helpers/README.md` for details.
