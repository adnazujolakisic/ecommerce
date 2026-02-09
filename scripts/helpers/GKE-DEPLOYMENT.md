# GKE Deployment Guide

## Prerequisites

1. **Google Cloud SDK (gcloud)** installed and authenticated
   ```bash
   gcloud auth login
   gcloud config set project YOUR_PROJECT_ID
   ```

2. **kubectl** configured for your GKE cluster
   ```bash
   gcloud container clusters get-credentials CLUSTER_NAME --zone ZONE --project YOUR_PROJECT_ID
   ```

3. **Docker** configured to push to GCR
   ```bash
   gcloud auth configure-docker
   ```

4. **GKE Cluster** created and running
   ```bash
   # For demo: 1 node is enough
   gcloud container clusters create metalmart-cluster \
     --num-nodes=1 \
     --zone=us-central1-a \
     --machine-type=e2-small
   
   # Or for slightly more headroom:
   # --machine-type=e2-medium
   ```

## Quick Deploy

```bash
export GCP_PROJECT_ID=your-project-id
./scripts/helpers/deploy-gke.sh
```

## Manual Steps

### 1. Set Project ID
```bash
export GCP_PROJECT_ID=your-project-id
```

### 2. Build and Push Images
```bash
# Build images
docker-compose build

# Tag for GCR
docker tag metalmart-catalogue:latest gcr.io/$GCP_PROJECT_ID/metalmart/catalogue:latest
docker tag metalmart-inventory:latest gcr.io/$GCP_PROJECT_ID/metalmart/inventory:latest
docker tag metalmart-checkout:latest gcr.io/$GCP_PROJECT_ID/metalmart/checkout:latest
docker tag metalmart-order:latest gcr.io/$GCP_PROJECT_ID/metalmart/order:latest
docker tag metalmart-order-processor:latest gcr.io/$GCP_PROJECT_ID/metalmart/order-processor:latest
docker tag metalmart-frontend:latest gcr.io/$GCP_PROJECT_ID/metalmart/frontend:latest

# Push to GCR
docker push gcr.io/$GCP_PROJECT_ID/metalmart/catalogue:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/inventory:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/checkout:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/order:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/order-processor:latest
docker push gcr.io/$GCP_PROJECT_ID/metalmart/frontend:latest
```

### 3. Update Kustomization
```bash
sed -i.bak "s/PROJECT_ID/$GCP_PROJECT_ID/g" k8s/overlays/gke/kustomization.yaml
```

### 4. Deploy Infrastructure First
```bash
kubectl apply -f k8s/base/namespace.yaml
kubectl apply -f k8s/base/infrastructure/secrets.yaml
kubectl apply -f k8s/base/infrastructure/postgres.yaml
kubectl apply -f k8s/base/infrastructure/kafka.yaml

# Wait for infrastructure
kubectl wait --for=condition=ready pod -l app=postgres -n metalmart --timeout=180s
kubectl wait --for=condition=ready pod -l app=zookeeper -n metalmart --timeout=180s
kubectl wait --for=condition=ready pod -l app=kafka -n metalmart --timeout=180s
```

### 5. Deploy Application Services
```bash
kubectl apply -k k8s/overlays/gke
```

### 6. Seed Inventory
```bash
CATALOGUE_POD=$(kubectl get pod -n metalmart -l app=catalogue -o jsonpath='{.items[0].metadata.name}')
kubectl exec -n metalmart $CATALOGUE_POD -- /bin/sh /app/seed-inventory.sh
```

### 7. Get Frontend URL
```bash
kubectl get svc frontend -n metalmart
# Wait for EXTERNAL-IP to be assigned, then access via that IP
```

## Differences from Minikube

1. **Images**: Pushed to GCR instead of local Docker
2. **Image Pull Policy**: Uses `IfNotPresent` (default) - GKE pulls from GCR
3. **Frontend Service**: Uses `LoadBalancer` type (already configured in base)
4. **Storage**: Uses GKE persistent volumes (automatically provisioned)

## Production Considerations

For production GKE deployments, consider:

1. **Managed Services**:
   - Use **Cloud SQL** instead of in-cluster PostgreSQL
   - Use **Confluent Cloud** or **Google Pub/Sub** instead of in-cluster Kafka

2. **Security**:
   - Use **Workload Identity** for service accounts
   - Enable **Pod Security Standards**
   - Use **Secret Manager** instead of Kubernetes secrets

3. **Scaling**:
   - Configure **Horizontal Pod Autoscaler (HPA)**
   - Use **Cluster Autoscaler** for node scaling

4. **Monitoring**:
   - Enable **Cloud Monitoring** and **Cloud Logging**
   - Set up **Alerting** for critical services

5. **Networking**:
   - Configure **Ingress** with SSL certificates
   - Use **Cloud Armor** for DDoS protection

## Troubleshooting

```bash
# Check pod status
kubectl get pods -n metalmart

# View logs
kubectl logs -n metalmart deployment/catalogue --tail=50

# Check service endpoints
kubectl get endpoints -n metalmart

# Describe pod for events
kubectl describe pod -n metalmart <pod-name>
```
What you need to do Set your GCP project ID: export GCP_PROJECT_ID=your-project-id Run the deployment script: ./scripts/helpers/deploy-gke.sh Key differences from Minikube Images: pushed to GCR instead of local Docker Image pull policy: uses IfNotPresent (GKE pulls from GCR automatically) Frontend access: uses LoadBalancer with external IP (no minikube service needed) Files created scripts/deploy-gke.sh — automated GKE deployment script GKE-DEPLOYMENT.md — deployment guide with manual steps and production considerations The script handles: Building and pushing images to GCR Updating the kustomization with your project ID Deploying infrastructure first, then apps Seeding inventory automatically Ready to deploy. Set your GCP_PROJECT_ID and run the script.