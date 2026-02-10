# Helper Scripts

Development and deployment utility scripts for MetalMart.

## Scripts

### `start-fresh.sh`
Complete Minikube deployment from scratch.
- Checks Docker Desktop
- Builds all images
- Deploys infrastructure (PostgreSQL, Kafka, Zookeeper)
- Deploys application services
- Seeds inventory data

**Usage:**
```bash
./scripts/helpers/start-fresh.sh
```

---

### `rebuild-frontend.sh`
Quick rebuild and restart of frontend in Kubernetes.

**Usage:**
```bash
./scripts/helpers/rebuild-frontend.sh
```

**What it does:**
1. Builds frontend with `npm run build`
2. Builds Docker image (handles Minikube docker-env automatically)
3. Tags image for Kubernetes
4. Restarts frontend deployment
5. Waits for rollout to complete
6. Verifies pod status

---

### `fix-order-status.sh`
Diagnoses and fixes order status update issues.

**Usage:**
```bash
./scripts/helpers/fix-order-status.sh
```

**What it does:**
1. Checks for Kafka connection errors
2. Restarts order service if needed
3. Restarts order-processor if needed
4. Provides troubleshooting commands

---

### `deploy-gke.sh`
Deploys MetalMart to Google Kubernetes Engine (GKE).

**Usage:**
```bash
export GCP_PROJECT_ID=your-project-id
./scripts/helpers/deploy-gke.sh
```

**What it does:**
1. Builds all Docker images
2. Tags images for GCR
3. Pushes images to GCR
4. Applies Kubernetes manifests with Kustomize

---

## When to Use

- **start-fresh.sh**: Starting from scratch, after major changes, or troubleshooting
- **rebuild-frontend.sh**: After frontend code changes
- **fix-order-status.sh**: When order status stops updating
- **deploy-gke.sh**: Deploying to production GKE cluster
