# Mirrord Operator Setup for Minikube

Complete guide to install and configure Mirrord operator for Kafka queue splitting and database branching.

## Prerequisites

1. **Mirrord CLI installed:**
   ```bash
   # macOS
   brew install metalbear-co/mirrord/mirrord

2. **Minikube running:**
   ```bash
   minikube start
   ```

3. **Helm installed:**
   ```bash
   # macOS
   brew install helm
   
   # Verify
   helm version
   ```

4. **Mirrord License Key:**
   - Get from: https://metalbear.com/mirrord
   - Or use trial license

---

## Step 1: Add/Update Mirrord Helm Repository

```bash
helm repo add mirrord https://helm.metalbear.co
helm repo update mirrord
```

**Verify:**
```bash
helm search repo mirrord
# Should show: mirrord/mirrord-operator
```

---

## Step 2: Install Mirrord Operator

### For Kafka Queue Splitting + Database Branching

```bash
helm install mirrord-operator mirrord/mirrord-operator \
  --set license.key=<YOUR_LICENSE_KEY> \
  --set operator.kafkaSplitting=true \
  --set operator.dbBranching=true \
  -n mirrord --create-namespace
```


## Step 3: Wait for Operator to Start

```bash
# Watch operator pod
kubectl get pods -n mirrord -w

# Wait until STATUS is "Running" (usually 30-60 seconds)
# Press Ctrl+C when ready
```


## Step 4: Verify Operator Installation

```bash
# Check operator status
mirrord operator status

# Should show:
# ✅ Operator Status (X.X.X)
#   ✅ operator found
#   ✅ operator version matches CLI
```

**If you see errors:**
```bash
# Check operator pod logs
kubectl logs -n mirrord deployment/mirrord-operator --tail=50

# Check if operator is running
kubectl get pods -n mirrord
```

---

## Step 5: Apply Kafka CRDs (for Queue Splitting)

```bash
kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml
```

**Verify CRDs are created:**
```bash
kubectl get crd | grep mirrord
# Should show:
# mirrordkafkaclientconfigs.queues.mirrord.metalbear.co
# mirrordkafkatopicsconsumers.queues.mirrord.metalbear.co
```

---

## Step 6: Verify Kafka Resources

```bash
# Check Kafka client config (note: plural name and mirrord namespace)
kubectl get mirrordkafkaclientconfigs -n mirrord

# Check Kafka topics consumer (note: plural name)
kubectl get mirrordkafkatopicsconsumers -n metalmart
```

---

## Complete Setup Checklist

- [ ] Mirrord CLI installed (`mirrord --version`)
- [ ] Helm installed (`helm version`)
- [ ] Mirrord Helm repo added (`helm repo list | grep mirrord`)
- [ ] Operator installed (`kubectl get pods -n mirrord`)
- [ ] Operator running (`mirrord operator status` shows )
- [ ] Kafka CRDs applied (`kubectl get crd | grep mirrord`)
- [ ] Kafka resources created (`kubectl get mirrordkafkaclientconfigs -n mirrord`)

---

## Testing the Setup

### Test Kafka Queue Splitting

```bash
cd services/order-processor
mirrord exec --config-file ../../.mirrord/queue-splitting.json -- go run main.go
```

**Expected:** Should connect without "operator not found" error.

### Test Database Branching

```bash
cd services/inventory
mirrord exec --config-file ../../.mirrord/db-branching.json -- go run main.go
```

**Expected:** Should create database branch and connect.

---

## Troubleshooting

### Error: "operator not found"

**Check 1: Is operator installed?**
```bash
kubectl get pods -n mirrord
# Should show mirrord-operator pod
```

**Check 2: Is operator running?**
```bash
kubectl get pods -n mirrord
# STATUS should be "Running"
```

**Check 3: Check operator logs:**
```bash
kubectl logs -n mirrord deployment/mirrord-operator --tail=50
```

**Fix: Reinstall operator:**
```bash
helm uninstall mirrord-operator -n mirrord
helm install mirrord-operator mirrord/mirrord-operator \
  --set license.key=<YOUR_LICENSE_KEY> \
  --set operator.kafkaSplitting=true \
  --set operator.dbBranching=true \
  -n mirrord --create-namespace
```

### Error: "Kafka queue splitting not supported"

**Problem:** Operator version too old.

**Check operator version:**
```bash
kubectl get deployment mirrord-operator -n mirrord -o jsonpath='{.spec.template.spec.containers[0].image}'
```

**Fix: Upgrade operator:**
```bash
helm upgrade mirrord-operator mirrord/mirrord-operator \
  --set license.key=<YOUR_LICENSE_KEY> \
  --set operator.kafkaSplitting=true \
  --set operator.dbBranching=true \
  -n mirrord
```

### Error: "no applicable MirrordKafkaTopicsConsumer found"

**Problem:** Kafka CRDs not applied.

**Fix:**
```bash
kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml
kubectl get mirrordkafkatopicsconsumer -n metalmart
```

### Error: "release not found" when upgrading

**Problem:** Operator installed in different namespace.

**Check:**
```bash
helm list -A | grep mirrord
# Note the namespace
```

**Fix:** Use correct namespace:
```bash
helm upgrade mirrord-operator mirrord/mirrord-operator \
  --set license.key=<YOUR_LICENSE_KEY> \
  --set operator.kafkaSplitting=true \
  -n <CORRECT_NAMESPACE>
```

---

## Uninstall (if needed)

```bash
# Remove operator
helm uninstall mirrord-operator -n mirrord

# Remove namespace
kubectl delete namespace mirrord

# Remove Kafka CRDs
kubectl delete -f k8s/base/infrastructure/mirrord-kafka.yaml
```

---

## Quick Reference

### Install Operator
```bash
helm install mirrord-operator mirrord/mirrord-operator \
  --set license.key=<KEY> \
  --set operator.kafkaSplitting=true \
  --set operator.dbBranching=true \
  -n mirrord --create-namespace
```

### Check Status
```bash
mirrord operator status
kubectl get pods -n mirrord
```

### Apply Kafka Resources
```bash
kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml
```

### Test Queue Splitting
```bash
cd services/order-processor
mirrord exec --config-file ../../.mirrord/queue-splitting.json -- go run main.go
```

### Test Database Branching
```bash
cd services/inventory
mirrord exec --config-file ../../.mirrord/db-branching.json -- go run main.go
```

---

## Next Steps

Once operator is installed and verified:

1. **Test Kafka Queue Splitting** - See `MIRRORD-FEATURES-DEMO.md`
2. **Test Database Branching** - See `MIRRORD-FEATURES-DEMO.md`
3. **Run Load Generator** - See `DEMO.md`
4. **Start Demo** - See `DEMO.md`
