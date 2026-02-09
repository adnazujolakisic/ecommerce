# MetalMart

A microservices e-commerce application designed as a demo platform for Mirrord features (Kafka queue splitting, database branching, filtering, steal mode, and mirror mode).

## Architecture

**Services (Go microservices):**
- `catalogue` (8081) - Product catalog
- `inventory` (8082) - Stock management
- `checkout` (8083) - Cart processing
- `order` (8084) - Order creation, publishes to Kafka
- `order-processor` - Kafka consumer, processes orders asynchronously

**Infrastructure:**
- PostgreSQL - Databases: catalogue, inventory, orders
- Kafka + Zookeeper - Async messaging for order processing
- Frontend - React + TypeScript + Vite

## Prerequisites

Before running the setup script, ensure you have the following installed:

1. **Docker Desktop** - [Install Docker Desktop](https://www.docker.com/products/docker-desktop)
   - Make sure Docker Desktop is running before starting

2. **docker-compose** - Usually included with Docker Desktop
   - Verify: `docker-compose --version`

3. **Minikube** - Local Kubernetes cluster (for Kubernetes deployment)
   - macOS: `brew install minikube`
   - Other platforms: [Minikube Installation Guide](https://minikube.sigs.k8s.io/docs/start/)

4. **kubectl** - Kubernetes command-line tool (for Kubernetes deployment)
   - macOS: `brew install kubectl`
   - Other platforms: [kubectl Installation Guide](https://kubernetes.io/docs/tasks/tools/)

## Quick Start

### Option 1: Kubernetes (Minikube) - Recommended

Once all prerequisites are installed, run:

```bash
./scripts/helpers/start-fresh.sh
```

This script will:
- Check all prerequisites
- Start Minikube
- Pre-pull infrastructure images (faster setup)
- Build all Docker images
- Deploy infrastructure (PostgreSQL, Kafka, Zookeeper)
- Deploy all application services
- Seed inventory data
- Verify deployment

**Access the application:**
```bash
# Access the frontend
minikube service frontend -n metalmart

# Check pod status
kubectl get pods -n metalmart
```

