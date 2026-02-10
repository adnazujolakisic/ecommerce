# GKE Deployment Prerequisites Checklist

Use this checklist before deploying to your own GKE account.

## ✅ Prerequisites

### 1. Google Cloud Account Setup
- [ ] **GCP Account** - You have a Google Cloud Platform account
- [ ] **Project Created** - You have a GCP project (or create one: `gcloud projects create PROJECT_ID`)
- [ ] **Billing Enabled** - Billing is enabled for your project
- [ ] **gcloud CLI Installed** - `gcloud --version` works
- [ ] **gcloud Authenticated** - `gcloud auth login` completed
- [ ] **Project Set** - `gcloud config set project YOUR_PROJECT_ID`

### 2. GKE Cluster
- [ ] **Cluster Created** - GKE cluster exists (or create with command below)
- [ ] **kubectl Configured** - `kubectl get nodes` works

**Create cluster (if needed):**
```bash
gcloud container clusters create metalmart-cluster \
  --num-nodes=1 \
  --zone=us-central1-a \
  --machine-type=e2-small \
  --project=YOUR_PROJECT_ID
```

**Get credentials:**
```bash
gcloud container clusters get-credentials metalmart-cluster \
  --zone=us-central1-a \
  --project=YOUR_PROJECT_ID
```

### 3. Docker & Container Registry
- [ ] **Docker Running** - `docker ps` works
- [ ] **GCR Access Configured** - `gcloud auth configure-docker` completed
- [ ] **GCR API Enabled** - Container Registry API enabled in your project

**Enable GCR API (if needed):**
```bash
gcloud services enable containerregistry.googleapis.com --project=YOUR_PROJECT_ID
```

### 4. Required Tools
- [ ] **kubectl** - `kubectl version --client` works
- [ ] **kustomize** - `kustomize version` works (or use `kubectl kustomize`)
- [ ] **docker-compose** - `docker-compose --version` works

### 5. Permissions
- [ ] **Kubernetes Engine Admin** - You have permissions to create/manage GKE clusters
- [ ] **Storage Admin** - You have permissions to push images to GCR
- [ ] **Service Account User** - For GKE operations

## 📋 What You Need

### Environment Variable
```bash
export GCP_PROJECT_ID=your-actual-project-id
```

### Files You Have (Already in Repo)
- ✅ `scripts/helpers/deploy-gke.sh` - Automated deployment script
- ✅ `k8s/overlays/gke/kustomization.yaml` - GKE overlay configuration
- ✅ `k8s/base/` - Base Kubernetes manifests
- ✅ `k8s/base/infrastructure/` - Infrastructure resources (Postgres, Kafka, secrets)
- ✅ All service Dockerfiles and code

## 🚀 Quick Start

Once all prerequisites are met:

```bash
# 1. Set your project ID
export GCP_PROJECT_ID=your-project-id

# 2. Run the deployment script
./scripts/helpers/deploy-gke.sh
```

## ⚠️ Important Notes

1. **Costs**: Running a GKE cluster costs money (~$25-50/month for minimal setup)
2. **NodePort vs LoadBalancer**: The script uses NodePort for cost savings. For production, consider LoadBalancer or Ingress
3. **Image Pull**: First deployment will take time to pull images from GCR
4. **Storage**: Persistent volumes will be created automatically by GKE

## 🔍 Verify Everything Works

After deployment:

```bash
# Check all pods are running
kubectl get pods -n metalmart

# Check services
kubectl get svc -n metalmart

# Access frontend
kubectl port-forward -n metalmart svc/frontend 3000:80
# Then open: http://localhost:3000
```

## 📚 Additional Resources

- [GKE Documentation](https://cloud.google.com/kubernetes-engine/docs)
- [GCR Documentation](https://cloud.google.com/container-registry/docs)
- Full deployment guide: `scripts/helpers/GKE-DEPLOYMENT.md`
