# Mirrord Testing Guide

Quick guide to test Mirrord features with MetalMart.

## Prerequisites Check

```bash
# 1. Check Mirrord CLI
mirrord --version

# 2. Check operator is running
kubectl get pods -n mirrord
mirrord operator status

# 3. Check Kafka resources
kubectl get mirrordkafkaclientconfig -n mirrord
kubectl get mirrordkafkatopicsconsumer -n metalmart

# 4. Verify MetalMart is running
kubectl get pods -n metalmart
```

## Test 1: Kafka Queue Splitting (Order Processor)

### Setup

```bash
# Enable demo mode for faster processing
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true

# Verify order-processor is running in cluster
kubectl logs -n metalmart deployment/order-processor -f
# (Press Ctrl+C after seeing it's working)
```

### Test Steps

**Terminal 1: Start Load Generator**
```bash
cd /Users/adna/Desktop/ecommerce

# Generate 20 orders/sec for 2 minutes (all with @metalbear.com emails)
BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
go run scripts/load-generator.go 20 120
```

**Terminal 2: Start Mirrord Debug Session**
```bash
cd /Users/adna/Desktop/ecommerce/services/order-processor

# Start with Mirrord (will route @metalbear.com orders to you)
mirrord exec -f ../../.mirrord/order-processor.json -- go run .
```

**What to expect:**
- Your local order-processor starts
- You should see: "Waiting for messages..."
- Within 5-10 seconds, you should see messages being processed
- Messages with `@metalbear.com` emails come to YOUR local instance
- Other messages go to cluster order-processor

**Verify it's working:**
```bash
# Check cluster order-processor logs (should show non-@metalbear.com orders)
kubectl logs -n metalmart deployment/order-processor --tail=20

# Check your local terminal (should show @metalbear.com orders)
```

### Troubleshooting

**No messages coming to local:**
- Check operator: `kubectl get pods -n mirrord`
- Check filter matches: `.mirrord/order-processor.json` has `".*@metalbear.com"`
- Verify load generator is running
- Check Kafka topic: `kubectl exec -n metalmart deployment/kafka -- kafka-console-consumer --bootstrap-server localhost:9092 --topic order.created --from-beginning`

**Operator not found:**
```bash
# Reinstall operator
helm uninstall mirrord-operator -n mirrord
helm install mirrord-operator mirrord/mirrord-operator \
  --set license.key=<YOUR_KEY> \
  --set operator.kafkaSplitting=true \
  -n mirrord --create-namespace
```

## Test 2: Database Branching (Inventory)

### Setup

```bash
# Check current inventory in cluster
kubectl exec -n metalmart deployment/inventory -- \
  curl -s http://localhost:8082/api/inventory/1
```

### Test Steps

**Terminal 1: Run with Database Branching**
```bash
cd /Users/adna/Desktop/ecommerce/services/inventory

# This creates an isolated database branch
mirrord exec -f ../../.mirrord/inventory-db-branch.json -- go run .
```

**Terminal 2: Make Changes**
```bash
# Reserve inventory (this affects YOUR branch, not cluster)
curl -X POST http://localhost:8082/api/inventory/reserve \
  -H "Content-Type: application/json" \
  -d '{"items": [{"productId": "1", "quantity": 50}]}'

# Check YOUR branch
curl http://localhost:8082/api/inventory/1
# Should show reduced stock

# Check CLUSTER database (should be unchanged)
kubectl exec -n metalmart deployment/inventory -- \
  curl -s http://localhost:8082/api/inventory/1
# Should show original stock
```

**What to expect:**
- Your local service connects to a BRANCHED database
- Changes you make don't affect the cluster
- Cluster database stays untouched

## Test 3: Steal Mode (Inventory)

### Test Steps

```bash
cd /Users/adna/Desktop/ecommerce/services/inventory

# Use steal mode - intercepts ALL traffic
mirrord exec -f ../../.mirrord/inventory-steal.json -- go run .
```

**What to expect:**
- ALL requests to inventory service come to YOUR local instance
- Cluster inventory service receives NO traffic
- Useful for testing changes without affecting cluster

**Test it:**
```bash
# Make request (goes to YOUR local instance)
curl http://inventory.metalmart.svc.cluster.local:8082/api/inventory/1

# Check cluster logs (should be empty - no requests)
kubectl logs -n metalmart deployment/inventory --tail=10
```

## Test 4: Mirror Mode (Catalogue)

### Test Steps

```bash
cd /Users/adna/Desktop/ecommerce/services/catalogue

# Mirror mode - copies traffic to you, cluster still receives it
mirrord exec -f ../../.mirrord/catalogue.json -- go run .
```

**What to expect:**
- Requests go to BOTH your local instance AND cluster
- You can debug without disrupting users
- Cluster continues serving requests normally

**Test it:**
```bash
# Make request (goes to BOTH)
curl http://catalogue.metalmart.svc.cluster.local:8081/api/products

# Check YOUR logs (should see request)
# Check CLUSTER logs (should also see request)
kubectl logs -n metalmart deployment/catalogue --tail=10
```

## Quick Test Checklist

- [ ] Operator installed and running
- [ ] Kafka resources applied
- [ ] Load generator creates orders
- [ ] Queue splitting routes messages correctly
- [ ] Database branching isolates changes
- [ ] Steal mode intercepts all traffic
- [ ] Mirror mode copies traffic

## Common Issues

### "operator not found"
```bash
# Check operator
kubectl get pods -n mirrord
mirrord operator status

# Reinstall if needed
helm install mirrord-operator mirrord/mirrord-operator \
  --set license.key=<KEY> \
  --set operator.kafkaSplitting=true \
  -n mirrord --create-namespace
```

### "no applicable MirrordKafkaTopicsConsumer found"
```bash
# Apply Kafka resources
kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml

# Verify
kubectl get mirrordkafkatopicsconsumer -n metalmart
```

### Messages not matching filter
- Check filter in `.mirrord/order-processor.json`
- Verify load generator creates `@metalbear.com` emails (it does by default)
- Check message headers in Kafka

### Database branching not working
- Verify operator has `dbBranching=true`
- Check config has `db_branches` section
- Check operator logs: `kubectl logs -n mirrord deployment/mirrord-operator`

## Quick Commands Reference

```bash
# Check operator status
mirrord operator status

# Test queue splitting
cd services/order-processor
mirrord exec -f ../../.mirrord/order-processor.json -- go run .

# Test database branching
cd services/inventory
mirrord exec -f ../../.mirrord/inventory-db-branch.json -- go run .

# Test steal mode
cd services/inventory
mirrord exec -f ../../.mirrord/inventory-steal.json -- go run .

# Generate load
BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
go run scripts/load-generator.go 20 120
```
