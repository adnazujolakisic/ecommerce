# Deployment Comparison: Minikube vs GKE

## Key Differences

### 1. **Image Storage & Pull Policy**

**Minikube:**
- Images built locally with Docker
- Images stored in local Docker daemon
- Uses `imagePullPolicy: Never` (assumes images exist locally)
- Image names: `metalmart-catalogue`, `metalmart-inventory`, etc.

**GKE:**
- Images built locally, then pushed to **Google Container Registry (GCR)**
- Images stored in `gcr.io/PROJECT_ID/metalmart/*`
- Uses `imagePullPolicy: IfNotPresent` (default - pulls from GCR)
- Image names: `gcr.io/PROJECT_ID/metalmart/catalogue`, etc.

### 2. **Image Paths**

**Minikube (`k8s/overlays/minikube/kustomization.yaml`):**
```yaml
images:
  - name: metalmart/catalogue
    newName: metalmart-catalogue  # Local Docker image
    newTag: latest
```

**GKE (`k8s/overlays/gke/kustomization.yaml`):**
```yaml
images:
  - name: metalmart/catalogue
    newName: gcr.io/PROJECT_ID/metalmart/catalogue  # GCR path
    newTag: latest
```

### 3. **Frontend Service Type**

**Minikube:**
- Service type: `NodePort` (patched in minikube overlay)
- Access via: `minikube service frontend -n metalmart`
- Gets a random port on the Minikube VM

**GKE:**
- Service type: `LoadBalancer` (from base config)
- Access via: External IP assigned by GCP
- Gets a public IP address automatically

### 4. **Deployment Process**

**Minikube:**
```bash
# 1. Build images locally
docker-compose build

# 2. Deploy (images already in local Docker)
./scripts/helpers/start-fresh.sh
# OR
kubectl apply -k k8s/overlays/minikube
```

**GKE:**
```bash
# 1. Build images locally
docker-compose build

# 2. Tag and push to GCR
export GCP_PROJECT_ID=your-project-id
./scripts/helpers/deploy-gke.sh
# OR manually tag/push, then:
kubectl apply -k k8s/overlays/gke
```

### 5. **Storage**

**Minikube:**
- Uses local persistent volumes (stored on Minikube VM)
- Data persists in `~/.minikube/`

**GKE:**
- Uses GKE persistent volumes (automatically provisioned)
- Data stored in GCP's persistent disk
- More reliable and scalable

### 6. **Network Access**

**Minikube:**
- Local development environment
- Access via port-forwarding or `minikube service`
- No external access by default

**GKE:**
- Cloud environment
- LoadBalancer provides external IP
- Accessible from internet (if configured)

---

## ArgoCD Configuration

### Current Setup

**File:** `argocd/apps/metalmart.yaml`

```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: metalmart
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/metalbear-co/metalmart.git
    targetRevision: HEAD
    path: k8s/overlays/minikube  # ⚠️ Points to Minikube, not GKE
  destination:
    server: https://kubernetes.default.svc
    namespace: metalmart
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
      - CreateNamespace=true
```

### ArgoCD Status

**Current State:**
- ✅ ArgoCD Application manifest exists
- ⚠️ **Points to Minikube overlay** (not GKE)
- ⚠️ **Points to GitHub repo** (`metalbear-co/metalmart`) - may need updating
- ✅ Configured for automated sync with self-healing

**What This Means:**
- If ArgoCD is installed in your cluster, it will:
  1. Watch the GitHub repo
  2. Sync changes from `k8s/overlays/minikube` path
  3. Automatically apply changes when code is pushed
  4. Self-heal if someone manually changes resources

**For GKE Deployment:**
- You would need a separate ArgoCD Application pointing to `k8s/overlays/gke`
- Or modify the existing one to point to GKE overlay

---

## Recommendations

### For Minikube (Local Development)
- Use `start-fresh.sh` script
- Images stay local
- Fast iteration
- Good for Mirrord debugging

### For GKE (Production/Staging)
- Use `deploy-gke.sh` script
- Images in GCR
- External access via LoadBalancer
- Better for demos and testing

### For ArgoCD Integration
1. **If using Minikube:** Current config is fine (if repo URL is correct)
2. **If using GKE:** Create a new ArgoCD app or update path:
   ```yaml
   source:
     path: k8s/overlays/gke  # Change this
   ```
3. **Update repo URL** if your repo is different from `metalbear-co/metalmart`

---

## Quick Reference

| Feature | Minikube | GKE |
|---------|----------|-----|
| Image Storage | Local Docker | GCR |
| Image Pull Policy | `Never` | `IfNotPresent` |
| Frontend Access | `minikube service` | LoadBalancer IP |
| Deployment Script | `start-fresh.sh` | `deploy-gke.sh` |
| Overlay Path | `k8s/overlays/minikube` | `k8s/overlays/gke` |
| Storage | Local VM | GCP Persistent Disk |
| Network | Local only | Public (if configured) |
