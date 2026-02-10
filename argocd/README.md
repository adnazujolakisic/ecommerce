# ArgoCD Configuration for MetalMart

This directory contains ArgoCD Application definitions for deploying MetalMart to different environments.

**Following the [MetalBear Playground](https://github.com/metalbear-co/playground) approach** - using Kustomize overlays directly.

## Applications

### `apps/metalmart.yaml`
- **Environment**: Minikube (Local Development)
- **Path**: `k8s/overlays/minikube`
- **Purpose**: Local development and testing
- **Image Source**: Local Docker images (`metalmart-*`)
- **Deploy**: `kubectl apply -k k8s/overlays/minikube` (or via ArgoCD)

### `apps/metalmart-gke.yaml`
- **Environment**: GKE (Google Kubernetes Engine)
- **Path**: `k8s/overlays/gke`
- **Purpose**: Production or staging deployment on GKE
- **Image Source**: Google Container Registry (`gcr.io/PROJECT_ID/metalmart/*`)
- **Deploy**: `kustomize build k8s/overlays/gke | kubectl apply -f -` (or via ArgoCD)

## Setup

### Prerequisites
1. ArgoCD installed in your cluster
2. Repository accessible to ArgoCD
3. For GKE: GCP project ID configured

### Installing Applications

**For Minikube:**
```bash
kubectl apply -f argocd/apps/metalmart.yaml
```

**For GKE:**
```bash
# First, replace PROJECT_ID in the kustomization file
export GCP_PROJECT_ID=your-project-id
sed -i.bak "s/PROJECT_ID/$GCP_PROJECT_ID/g" k8s/overlays/gke/kustomization.yaml

# Then apply the ArgoCD app
kubectl apply -f argocd/apps/metalmart-gke.yaml

# Or deploy directly (following playground approach):
kustomize build k8s/overlays/gke | kubectl apply -f -
```

## GKE PROJECT_ID Handling

The GKE overlay uses `PROJECT_ID` as a placeholder. You have three options:

### Option 1: Manual Replacement (Current Approach)
Replace `PROJECT_ID` in `k8s/overlays/gke/kustomization.yaml` before ArgoCD syncs:
```bash
sed -i.bak "s/PROJECT_ID/your-project-id/g" k8s/overlays/gke/kustomization.yaml
git add k8s/overlays/gke/kustomization.yaml
git commit -m "Set GCP project ID for GKE"
```

### Option 2: Environment-Specific Overlays (Recommended for Multiple Environments)
Create separate overlays for each environment:
- `k8s/overlays/gke-dev/` - Development GKE cluster
- `k8s/overlays/gke-prod/` - Production GKE cluster

Each overlay has the project ID hardcoded.

### Option 3: ArgoCD Parameters (Advanced)
Use ArgoCD's parameter substitution with ApplicationSets or plugins.

## Syncing

ArgoCD will automatically sync when:
- Git repository changes
- Manual sync is triggered via UI or CLI

**Manual sync via CLI:**
```bash
argocd app sync metalmart
argocd app sync metalmart-gke
```

**Manual sync via UI:**
1. Open ArgoCD UI
2. Select the application
3. Click "Sync"

## Troubleshooting

**Application stuck in "Unknown" or "OutOfSync":**
- Check if `PROJECT_ID` is replaced in GKE overlay
- Verify repository access
- Check ArgoCD logs: `kubectl logs -n argocd deployment/argocd-repo-server`

**Images not found:**
- For GKE: Ensure images are pushed to GCR
- For Minikube: Ensure images are built locally
- Check image pull secrets if using private registries

**Sync fails:**
- Check ArgoCD application events: `kubectl describe application metalmart-gke -n argocd`
- Verify kustomization.yaml is valid: `kubectl kustomize k8s/overlays/gke`
