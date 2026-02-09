# MetalMart Mirrord Demo Guide

Complete guide for demonstrating Mirrord features with MetalMart - for customer presentations and internal testing.

---

## Quick Start (Customer Demo)

**The 3-Step Demo:**

1. **Start Load Generator** (Terminal 1)
2. **Start Mirrord** (Terminal 2)  
3. **Show Breakpoint Hit** 🎯

**Commands:**
```bash
# Terminal 1: Generate load
cd /Users/adna/Desktop/ecommerce
BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
go run scripts/load-generator.go 20 120

# Terminal 2: Start Mirrord
cd /Users/adna/Desktop/ecommerce/services/order-processor
mirrord exec -f ../../.mirrord/order-processor.json -- go run .
```

**What happens:**
- Load generator creates 20 orders/sec (all with `@metalbear.com` emails)
- Mirrord routes matching messages to your local debugger
- Breakpoint hits within 5-10 seconds
- You see real order data from the cluster's Kafka

---

## Prerequisites Check

Before starting, verify everything is set up:

```bash
# 1. Check Mirrord CLI
mirrord --version

# 2. Check operator is running
kubectl get pods -n mirrord
mirrord operator status

# 3. Check Kafka resources (note: plural names and correct namespaces)
kubectl get mirrordkafkaclientconfigs -n mirrord
kubectl get mirrordkafkatopicsconsumers -n metalmart

# 4. Verify MetalMart is running
kubectl get pods -n metalmart

# 5. Check DEMO_MODE is enabled (optional but recommended)
kubectl describe deployment order-processor -n metalmart | grep DEMO_MODE
```

---

## Demo 1: Kafka Queue Splitting (Main Demo)

This is the primary demo for customers - shows how Mirrord routes specific Kafka messages to your local debugger.

### Setup (~2 minutes)

```bash
# 1. Make sure everything is deployed
kubectl get pods -n metalmart

# 2. Enable demo mode (faster processing - recommended for demos)
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true

# Verify it's set:
kubectl describe deployment order-processor -n metalmart | grep DEMO_MODE
# Should show: DEMO_MODE: true

# 3. Verify Mirrord operator is running
kubectl get pods -n mirrord
```

### Demo Steps

**Step 1: Start Load Generator (Terminal 1)**

```bash
cd /Users/adna/Desktop/ecommerce

# For Kubernetes:
BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
go run scripts/load-generator.go 20 120

# OR for local Docker Compose:
go run scripts/load-generator.go 20 120
```

**What this does:**
- Creates 20 orders per second
- Runs for 2 minutes (120 seconds)
- Total: ~2,400 orders
- All orders have `@metalbear.com` emails (matches your filter)

**Leave this running** - it creates orders in the background.

**Step 2: Start Mirrord Debug Session (Terminal 2)**

```bash
cd /Users/adna/Desktop/ecommerce/services/order-processor

# Option A: Command line
mirrord exec -f ../../.mirrord/order-processor.json -- go run .

# Option B: VSCode Debugger
# 1. Set breakpoint at line 127 in main.go
# 2. Use launch config with mirrord
# 3. Start debugging
```

**What happens:**
1. Mirrord connects to the cluster
2. Your local app starts: "Order processor started, waiting for messages..."
3. Load generator is creating messages with `@metalbear.com` emails
4. **Within 5-10 seconds: Breakpoint hits!** 🎯
5. Messages with `@metalbear.com` emails come to YOUR local instance
6. Other messages go to cluster order-processor

**Step 3: Verify It's Working**

```bash
# Check cluster order-processor logs (should show non-@metalbear.com orders)
kubectl logs -n metalmart deployment/order-processor --tail=20

# Check your local terminal (should show @metalbear.com orders)
```

### Customer Talking Points

**When breakpoint hits, explain:**

1. **Show the message data:**
   ```
   "Look at this - real order ID, real customer email, real total amount.
   This isn't a mock - this came from the cluster's Kafka."
   ```

2. **Explain the flow:**
   ```
   "The load generator created this order → Order Service published to Kafka →
   Mirrord operator intercepted it → Checked the email header →
   Matched my filter (@metalbear.com) → Routed it to my laptop →
   Breakpoint hit!"
   ```

3. **Show filtering:**
   - Open `.mirrord/order-processor.json`
   - Point to: `"customer_email": ".*@metalbear.com"`
   - Explain: "Only messages matching this pattern come to me"

4. **Show isolation:**
   - Place an order from frontend with email like `test@example.com`
   - Show it goes to cluster (doesn't hit your breakpoint)
   - Explain: "Other developers' orders don't interrupt my debugging"

### Load Generator Settings

| Scenario | Rate | Duration | Total Orders | When to Use |
|----------|------|----------|--------------|-------------|
| Light | 5/sec | 60s | 300 | Basic demo |
| Medium | 20/sec | 120s | 2,400 | **Recommended for demos** |
| Heavy | 50/sec | 60s | 3,000 | Stress test |
| Extreme | 100/sec | 30s | 3,000 | Maximum throughput |

**Key Points:**
1. ✅ Start load generator FIRST
2. ✅ Then start Mirrord debug session
3. ✅ Breakpoint will hit within seconds
4. ✅ Messages keep flowing while you demo
5. ✅ Stop load generator when done (Ctrl+C)

---

## Demo 2: Database Branching (Inventory)

Shows how Mirrord creates isolated database copies so you can debug without affecting the shared database.

### Setup

```bash
# Check current inventory in cluster
kubectl exec -n metalmart deployment/inventory -- \
  curl -s http://localhost:8082/api/inventory/1
```

### Demo Steps

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

---

## Demo 3: Steal Mode (Inventory)

Shows how to intercept ALL traffic to a service for testing.

### Demo Steps

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

---

## Demo 4: Mirror Mode (Catalogue)

Shows how to copy traffic to your local instance while cluster still serves requests.

### Demo Steps

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

---

## Troubleshooting

### My breakpoint doesn't hit

**Check 1: Load generator is running**
```bash
# Should see orders being created
# If not, restart: BASE_URL=http://order.metalmart.svc.cluster.local:8084 go run scripts/load-generator.go 20 120
```

**Check 2: Mirrord operator is running**
```bash
kubectl get pods -n mirrord
mirrord operator status
```

**Check 3: Kafka resources exist**
```bash
kubectl get mirrordkafkaclientconfigs -n mirrord
kubectl get mirrordkafkatopicsconsumers -n metalmart
```

**Check 4: Filter matches**
- Verify `.mirrord/order-processor.json` has `".*@metalbear.com"`
- Load generator creates `@metalbear.com` emails by default

**Check 5: Messages in Kafka**
```bash
kubectl exec -n metalmart deployment/kafka -- \
  kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic order.created --from-beginning --max-messages 5
```

### "operator not found"

```bash
# Check operator
kubectl get pods -n mirrord
mirrord operator status

# Reinstall if needed
helm install mirrord-operator mirrord/mirrord-operator \
  --set license.key=<YOUR_KEY> \
  --set operator.kafkaSplitting=true \
  --set operator.dbBranching=true \
  -n mirrord --create-namespace
```

### "no applicable MirrordKafkaTopicsConsumer found"

```bash
# Apply Kafka resources
kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml

# Verify
kubectl get mirrordkafkatopicsconsumers -n metalmart
```

### Messages not matching filter

- Check filter in `.mirrord/order-processor.json`
- Verify load generator creates `@metalbear.com` emails (it does by default)
- Check message headers in Kafka

### Database branching not working

- Verify operator has `dbBranching=true`
- Check config has `db_branches` section
- Check operator logs: `kubectl logs -n mirrord deployment/mirrord-operator`

### No messages in queue

- Increase load generator rate: `go run scripts/load-generator.go 50 60`
- Check order service is accessible
- Verify inventory is seeded

### Load generator fails

- Check order service is accessible
- Verify inventory is seeded
- Check network connectivity

---

## Quick Commands Reference

```bash
# Check operator status
mirrord operator status

# Check DEMO_MODE
kubectl describe deployment order-processor -n metalmart | grep DEMO_MODE

# Test queue splitting (main demo)
cd services/order-processor
mirrord exec -f ../../.mirrord/order-processor.json -- go run .

# Test database branching
cd services/inventory
mirrord exec -f ../../.mirrord/inventory-db-branch.json -- go run .

# Test steal mode
cd services/inventory
mirrord exec -f ../../.mirrord/inventory-steal.json -- go run .

# Test mirror mode
cd services/catalogue
mirrord exec -f ../../.mirrord/catalogue.json -- go run .

# Generate load
BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
go run scripts/load-generator.go 20 120
```

---

## Key Talking Points

1. **Real Data**: "These aren't mocks - these are real orders from the cluster"
2. **No Disruption**: "My debugging doesn't affect anyone else"
3. **Selective Routing**: "I only get messages that match my filter"
4. **Production-Like**: "This is the same Kafka, same services, same data"
5. **Fast Iteration**: "I can debug, fix, and test without redeploying"

---

## Demo Mode Explained

**What is DEMO_MODE?**
- Optional setting that speeds up order processing
- Without it: Each order takes 10 seconds (2s + 3s + 5s delays)
- With it: Each order takes 600ms (200ms + 200ms + 200ms delays)

**Do you need it?**
- For demos: **Yes** - makes demos faster and more compelling
- For real testing: **Optional** - Mirrord works fine without it

**Check if enabled:**
```bash
kubectl describe deployment order-processor -n metalmart | grep DEMO_MODE
```

**Enable it:**
```bash
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true
```
