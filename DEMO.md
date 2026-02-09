# MetalMart Enterprise Demo Guide

This is my comprehensive guide for demonstrating Mirrord's Kafka Queue Splitting feature. I use this for enterprise demos and presentations.

---

## Quick Start

**When to use the load generator:** I always start the load generator BEFORE my Mirrord debug session to create a steady stream of messages in Kafka.

### Step 1: Setup (~2 minutes)

```bash
# 1. Make sure everything is deployed
kubectl get pods -n metalmart

# 2. Enable demo mode (faster processing - I use this for demos)
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true

# 3. Verify Mirrord operator is running (if using Mirrord)
kubectl get pods -n mirrord
```

### Step 2: Start Load Generator (Do this BEFORE debugging)

**I open Terminal 1** and run:

```bash
cd /Users/adna/Desktop/metalmart

# For Kubernetes (what I usually use):
BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
go run scripts/load-generator.go 20 120

# OR for local Docker Compose:
go run scripts/load-generator.go 20 120
```

**What this does:**
- Creates 20 orders per second
- Runs for 2 minutes (120 seconds)
- Total: ~2,400 orders
- All orders have `@metalbear.com` emails (will match my filter)

**I leave this running** - it creates orders in the background.

### Step 3: Start Mirrord Debug Session (While load generator is running)

**I open Terminal 2** (or VSCode debugger):

```bash
cd /Users/adna/Desktop/metalmart/services/order-processor

# I set a breakpoint at line 127 in main.go
# Then start Mirrord:
mirrord exec -f ../../.mirrord/order-processor.json -- go run .
```

**What happens:**
1. Mirrord connects to the cluster
2. My local app starts waiting for messages
3. Load generator is creating messages with `@metalbear.com` emails
4. **My breakpoint hits** - messages are being routed to me!

### Step 4: Stop Load Generator

When I'm done with the demo:
- Press `Ctrl+C` in Terminal 1 (load generator)
- Stop my debug session

### Quick Reference

**Load Generator Settings I Use:**
- `20 120` = 20 orders/sec for 2 minutes (my recommended for demo)
- `10 60` = 10 orders/sec for 1 minute (lighter load)
- `50 30` = 50 orders/sec for 30 seconds (heavy load)

**Key Points:**
1. ✅ Start load generator FIRST
2. ✅ Then start Mirrord debug session
3. ✅ Breakpoint will hit within seconds
4. ✅ Messages keep flowing while I demo
5. ✅ Stop load generator when done

### Visual Timeline

```
Time 0:00 → Start load generator (Terminal 1)
           └─ Creates 20 orders/sec continuously

Time 0:30 → Start Mirrord debug session (Terminal 2/VSCode)
           └─ Set breakpoint at line 127
           └─ Run: mirrord exec ...

Time 0:35 → Breakpoint hits! 🎯
           └─ Show message data
           └─ Explain queue splitting
           └─ Show filter config

Time 2:00 → Load generator finishes (2400 orders created)
           └─ Stop load generator
           └─ Continue demo or stop
```

---

## Demo Goals

When I demo MetalMart, I want to show:

1. **Mirrord Queue Splitting** - Route specific Kafka messages to my local debugger
2. **Real-World Load** - Generate enough traffic to stress Kafka
3. **Isolation** - Debug locally without disrupting cluster operations
4. **Enterprise-Ready** - Show production-like message throughput

---

## The Problem I'm Solving

**The default setup is too slow for a compelling demo:**
- Orders process sequentially with 10-second delays
- Low message throughput (1-2 orders/minute)
- Not enough load to showcase queue splitting effectively

---

## My Solution: Load Generator + Demo Mode

### 1. Enable Demo Mode (Faster Processing)

I set `DEMO_MODE=true` to reduce processing delays from 10 seconds to 600ms:

```bash
# In Kubernetes
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true
```

**Processing times:**
- Normal: 2s → 3s → 5s = 10 seconds per order
- Demo: 200ms → 200ms → 200ms = 600ms per order

### 2. Use Load Generator

I generate orders at configurable rates:

```bash
# Using Go version (what I prefer)
cd scripts
go run load-generator.go [rate] [duration]

# Examples I use:
go run load-generator.go 10 60    # 10 orders/sec for 60 seconds = 600 orders
go run load-generator.go 20 120   # 20 orders/sec for 2 minutes = 2,400 orders (my go-to)
go run load-generator.go 50 30    # 50 orders/sec for 30 seconds = 1,500 orders

# Using shell script (alternative)
chmod +x load-generator.sh
./load-generator.sh 10 60
```

---

## Recommended Demo Flow

### Setup (5 minutes)

1. **Deploy to Kubernetes:**
   ```bash
   ./scripts/helpers/start-fresh.sh
   # Or manually: kubectl apply -k k8s/overlays/minikube
   ```

2. **Enable demo mode:**
   ```bash
   kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true
   ```

3. **Start load generator** (in background):
   ```bash
   BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
   go run scripts/load-generator.go 20 120
   ```
   This creates 20 orders/second for 2 minutes = 2,400 orders

### Demo Script (10 minutes)

1. **Show the problem** (2 min)
   - "Debugging Kafka consumers is hard - you can't easily debug in production"
   - "Running locally means setting up Kafka, Zookeeper, all dependencies..."
   - "We need a way to debug with real data, in real time, without disrupting others"

2. **Start Mirrord debug session** (1 min)
   - Open VSCode
   - Set breakpoint in `order-processor/main.go` line 127
   - Start debug with `.mirrord/order-processor.json`
   - Show: "Order processor started, waiting for messages..."

3. **Show queue splitting in action** (3 min)
   - Load generator is creating 20 orders/second
   - Cluster order-processor is handling most messages
   - **My breakpoint hits** - show the message came to my laptop
   - Inspect variables: real order ID, customer email, total amount
   - Explain: "This message matched my filter, so it came to me"

4. **Demonstrate filtering** (2 min)
   - Show `.mirrord/order-processor.json` filter:
     ```json
     "message_filter": {
       "customer_email": ".*@metalbear.com"
     }
     ```
   - Explain: Only messages with `@metalbear.com` emails come to me
   - Other messages go to cluster consumer
   - Show Kafka topic metrics to prove messages are being split

5. **Show isolation** (2 min)
   - Place order from frontend with different email
   - Show it goes to cluster (not my breakpoint)
   - Explain: "My teammate's orders don't interrupt my debugging"

### Demo Talking Points

**When my breakpoint hits, I explain:**

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
   - I open `.mirrord/order-processor.json`
   - Point to: `"customer_email": ".*@metalbear.com"`
   - Explain: "Only messages matching this pattern come to me"

4. **Show isolation:**
   - I place an order from frontend with email like `test@example.com`
   - Show it goes to cluster (doesn't hit my breakpoint)
   - Explain: "Other developers' orders don't interrupt my debugging"

---

## Metrics I Show

### Kafka Topic Metrics

```bash
# Show message rate
kubectl exec -n metalmart deployment/kafka -- \
  kafka-run-class kafka.tools.GetOffsetShell \
  --broker-list localhost:9092 \
  --topic order.created \
  --time -1

# Show consumer lag
kubectl exec -n metalmart deployment/kafka -- \
  kafka-consumer-groups --bootstrap-server localhost:9092 \
  --describe --group order-processor
```

### Order Processing Stats

```bash
# Count orders processed
kubectl logs -n metalmart deployment/order-processor --tail=100 | \
  grep "processing complete" | wc -l

# Show processing rate
kubectl logs -n metalmart deployment/order-processor --tail=1000 | \
  grep "status updated to: shipped" | wc -l
```

---

## Load Settings I Use

| Scenario | Rate | Duration | Total Orders | When I Use It |
|----------|------|----------|--------------|---------------|
| Light | 5/sec | 60s | 300 | Basic demo |
| Medium | 20/sec | 120s | 2,400 | **My go-to for demos** |
| Heavy | 50/sec | 60s | 3,000 | Stress test |
| Extreme | 100/sec | 30s | 3,000 | Maximum throughput |

---

## Troubleshooting

### My breakpoint doesn't hit
- Check load generator is running (Terminal 1)
- Verify filter matches: `.*@metalbear.com`
- Check Mirrord operator is running: `kubectl get pods -n mirrord`
- Verify Kafka resources: `kubectl get mirrordkafkatopicsconsumer -n metalmart`

### No messages in queue
- Increase load generator rate: `go run scripts/load-generator.go 50 60`
- Check order service is accessible
- Verify inventory is seeded

### Messages not matching filter
- Check load generator creates `@metalbear.com` emails (it does by default)
- Verify filter in `.mirrord/order-processor.json`

### Load generator fails
- Check order service is accessible
- Verify inventory is seeded
- Check network connectivity

### Not enough messages in queue
- Increase load generator rate
- Extend duration
- Enable demo mode for faster processing

### Breakpoint not hitting
- Verify Mirrord operator is running
- Check filter matches message headers
- Ensure debug session started before load generation

---

## Key Talking Points I Use

1. **Real Data**: "These aren't mocks - these are real orders from the cluster"
2. **No Disruption**: "My debugging doesn't affect anyone else"
3. **Selective Routing**: "I only get messages that match my filter"
4. **Production-Like**: "This is the same Kafka, same services, same data"
5. **Fast Iteration**: "I can debug, fix, and test without redeploying"

---

## See Also

- [MIRRORD-FEATURES-DEMO.md](MIRRORD-FEATURES-DEMO.md) - All Mirrord features demo
