# MetalMart

A microservices e-commerce application designed as a demo platform for Mirrord features.

## Prerequisites

Before running the setup script, ensure you have the following installed:

1. **Docker Desktop** - [Install Docker Desktop](https://www.docker.com/products/docker-desktop)
   - Make sure Docker Desktop is running before starting

2. **docker-compose** - Usually included with Docker Desktop
   - Verify: `docker-compose --version`

3. **Minikube** - Local Kubernetes cluster
   - macOS: `brew install minikube`
   - Other platforms: [Minikube Installation Guide](https://minikube.sigs.k8s.io/docs/start/)

4. **kubectl** - Kubernetes command-line tool
   - macOS: `brew install kubectl`
   - Other platforms: [kubectl Installation Guide](https://kubernetes.io/docs/tasks/tools/)

## Quick Start

Once all prerequisites are installed, run:

```bash
./scripts/helpers/start-fresh.sh
```

This script will:
- ✅ Check all prerequisites
- ✅ Start Minikube
- ✅ Pre-pull infrastructure images (faster setup)
- ✅ Build all Docker images
- ✅ Deploy infrastructure (PostgreSQL, Kafka, Zookeeper)
- ✅ Deploy all application services
- ✅ Seed inventory data
- ✅ Verify deployment

## Access the Application

After setup completes:

```bash
# Access the frontend
minikube service frontend -n metalmart

# Check pod status
kubectl get pods -n metalmart
```

## Documentation

- **[DEVELOPMENT.md](DEVELOPMENT.md)** - Complete development guide
- **[PROJECT-CONTEXT.md](PROJECT-CONTEXT.md)** - Project overview and architecture
- **[MIRRORD-SETUP.md](MIRRORD-SETUP.md)** - Mirrord operator setup (optional)
- **[scripts/helpers/README.md](scripts/helpers/README.md)** - Helper scripts documentation

## Troubleshooting

If you encounter issues:

1. **Check prerequisites**: The script will verify all required tools are installed
2. **Check Docker**: Make sure Docker Desktop is running
3. **Check pods**: `kubectl get pods -n metalmart`
4. **View logs**: `kubectl logs -n metalmart deployment/<service-name>`
5. **Describe pod**: `kubectl describe pod <pod-name> -n metalmart`

For more help, see [DEVELOPMENT.md](DEVELOPMENT.md).
