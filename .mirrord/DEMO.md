# MetalMart Mirrord Demo Guide

Complete guide for demonstrating Mirrord features with MetalMart - for customer presentations and internal testing.

---

## Quick Start (Customer Demo)

**Two Ways to Demo:**

### Option A: Manual Order (Recommended - Better for customer demos)
1. **Start Mirrord** (Terminal 1)
2. **Open Frontend** and place order manually
3. **Breakpoint hits immediately** 🎯

### Option B: Load Generator (For high volume demos)
1. **Start Load Generator** (Terminal 1) - creates 20 orders/sec
2. **Start Mirrord** (Terminal 2)  
3. **Breakpoint hits within 5-10 seconds** 🎯

**Quick Commands:**

**Option A - Manual:**
```bash
# Terminal 1: Start Mirrord
cd /Users/adna/Desktop/ecommerce/services/order-processor
mirrord exec -f ../../.mirrord/queue-splitting.json -- go run .

# Then open frontend and place order (email pre-filled with demo@metalbear.com)
minikube service frontend -n metalmart
```

**Option B - Load Generator:**
```bash
# Terminal 1: Generate load
cd /Users/adna/Desktop/ecommerce
BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
go run scripts/load-generator.go 20 120

# Terminal 2: Start Mirrord
cd /Users/adna/Desktop/ecommerce/services/order-processor
mirrord exec -f ../../.mirrord/queue-splitting.json -- go run .
```

**What happens:**
- Messages with `@metalbear.com` emails route to your local debugger
- Breakpoint hits (immediately for manual, 5-10 sec for load generator)
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

**Option 1: Load Generator (Automated - for high volume demos)**

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

**Where to set breakpoints (to show status progression):**

Set breakpoints at these lines in `services/order-processor/main.go`:

1. **Line 127** - Entry point (message received)
   ```go
   func (p *OrderProcessor) ProcessOrder(event OrderCreatedEvent) error {
   ```

2. **Line 164** - Status update (will hit 3 times: processing → confirmed → shipped)
   ```go
   if err := p.updateOrderStatus(event.OrderID, step.status); err != nil {
   ```

**How to set breakpoints:**
- Click to the left of the line number (red dot appears)
- Or press F9 on that line
- Set both breakpoints to see the full flow

**Option A: Command line (no breakpoint needed - just watch logs)**
```bash
cd /Users/adna/Desktop/ecommerce/services/order-processor
mirrord exec -f ../../.mirrord/queue-splitting.json -- go run .
```

**Option B: VSCode Debugger (with breakpoints - recommended for demos)**

**Set these breakpoints to show status progression:**

1. **Line 127** - Entry point (message received)
   ```go
   func (p *OrderProcessor) ProcessOrder(event OrderCreatedEvent) error {
   ```
   - Click left of line number (red dot) or press F9

2. **Line 164** - Status update (hits 3 times: processing → confirmed → shipped)
   ```go
   if err := p.updateOrderStatus(event.OrderID, step.status); err != nil {
   ```
   - Click left of line number (red dot) or press F9

**Demo flow with breakpoints:**
1. Start debugging - breakpoint at line 127 hits first
2. Show customer the order data in Variables panel
3. Continue (F5) - breakpoint at line 164 hits (status = "processing")
4. Continue (F5) - breakpoint at line 164 hits again (status = "confirmed")
5. Continue (F5) - breakpoint at line 164 hits again (status = "shipped")
6. Continue (F5) - order processing complete

This shows the full lifecycle: Message → Processing → Confirmed → Shipped

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

---

**Option 2: Manual Order Placement (Recommended for customer demos)**

**Better for demos:** More controlled, customer sees the full flow, easier to explain.

**Step 1: Start Mirrord Debug Session (Terminal 1)**

**Where to set breakpoints (to show status progression):**

Set breakpoints at these lines in `services/order-processor/main.go`:

1. **Line 127** - Entry point (message received)
   ```go
   func (p *OrderProcessor) ProcessOrder(event OrderCreatedEvent) error {
   ```

2. **Line 164** - Status update (will hit 3 times: processing → confirmed → shipped)
   ```go
   if err := p.updateOrderStatus(event.OrderID, step.status); err != nil {
   ```

**How to set breakpoints:**
- Click to the left of the line number (red dot appears)
- Or press F9 on that line
- Set both breakpoints to see the full flow

**Option A: Command line**
```bash
cd /Users/adna/Desktop/ecommerce/services/order-processor
mirrord exec -f ../../.mirrord/queue-splitting.json -- go run .
```

**Option B: VSCode Debugger (with breakpoints - recommended for demos)**

**Set these breakpoints to show status progression:**

1. **Line 127** - Entry point (message received)
   ```go
   func (p *OrderProcessor) ProcessOrder(event OrderCreatedEvent) error {
   ```
   - Click left of line number (red dot) or press F9

2. **Line 164** - Status update (hits 3 times: processing → confirmed → shipped)
   ```go
   if err := p.updateOrderStatus(event.OrderID, step.status); err != nil {
   ```
   - Click left of line number (red dot) or press F9

**Demo flow with breakpoints:**
1. Start debugging - breakpoint at line 127 hits first
2. Show customer the order data in Variables panel
3. Continue (F5) - breakpoint at line 164 hits (status = "processing")
4. Continue (F5) - breakpoint at line 164 hits again (status = "confirmed")
5. Continue (F5) - breakpoint at line 164 hits again (status = "shipped")
6. Continue (F5) - order processing complete

This shows the full lifecycle: Message → Processing → Confirmed → Shipped

Wait for: "Order processor started, waiting for messages..."

**Step 2: Open Frontend and Place Order**

```bash
# Get frontend URL
minikube service frontend -n metalmart
# Or if using port-forward:
kubectl port-forward -n metalmart svc/frontend 3000:80
```

**In the browser:**
1. Open the frontend (http://localhost:3000 or the minikube service URL)
2. Add items to cart
3. Go to checkout
4. **Email is pre-filled with `demo@metalbear.com`** (matches your filter)
5. Click "Place Order"

**Step 3: Breakpoint Hits!**

- Within 1-2 seconds, your breakpoint at line 127 should hit
- Show the customer the real order data in Variables panel
- Explain how the message came from the cluster's Kafka

**Step 4: Show Status Progression**

- Continue execution (F5 or click Continue button)
- Breakpoint at line 164 will hit 3 times:
  1. **First hit:** Check Variables panel → `step.status = "processing"`
     - Point out: "Now updating status to processing"
     - Continue (F5)
  2. **Second hit:** Check Variables panel → `step.status = "confirmed"`
     - Point out: "Status updated to confirmed"
     - Continue (F5)
  3. **Third hit:** Check Variables panel → `step.status = "shipped"`
     - Point out: "Final status - shipped"
     - Continue (F5)
- **Final:** See "Order processing complete" in debug console
- This shows the full order lifecycle happening on your laptop, with real status updates to the cluster

**Advantages:**
- ✅ More controlled - one order at a time
- ✅ Customer sees the full user flow
- ✅ Easier to explain what's happening
- ✅ No load generator needed
- ✅ Can place multiple orders manually to show it working repeatedly

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
   - Open `.mirrord/queue-splitting.json`
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

**Option A: Interactive (Using Frontend - Recommended for demos)**

**Terminal 1: Run with Database Branching + Mirror Mode**

**Where to set breakpoints (to show database operations):**

Set breakpoints at these lines in `services/inventory/`:

1. **`handlers/handlers.go` line 24** - Reading from database
   ```go
   inv, err := h.store.GetInventory(productID)
   ```
   - Shows when inventory is fetched from your branched database

2. **`handlers/handlers.go` line 41** - Reserving inventory (key operation)
   ```go
   reservationID, err := h.store.Reserve(req.Items)
   ```
   - Shows when inventory reservation happens in your branch

3. **`store/postgres.go` line 112** - Database transaction starts
   ```go
   func (s *PostgresStore) Reserve(items []models.ReserveItem) (string, error) {
   ```
   - Shows the database operation beginning

4. **`store/postgres.go` line 137** - Database UPDATE (the actual write)
   ```go
   _, err = tx.Exec(`
       UPDATE inventory
       SET reserved_quantity = reserved_quantity + $1, last_updated = NOW()
       WHERE product_id = $2
   `, item.Quantity, item.ProductID)
   ```
   - Shows the actual database write to your branch

**How to set breakpoints:**
- Click to the left of the line number (red dot appears)
- Or press F9 on that line

**How to run:**

**Option 1: VSCode Extension (Recommended - if installed)**
1. Install the **Mirrord VSCode Extension** from the marketplace
2. Set your breakpoints in VSCode
3. Use the extension's "Run with Mirrord" command
4. Select your config file: `.mirrord/db-branching.json`
5. Breakpoints work automatically - no launch.json needed

**Option 2: Terminal (Works without extension)**
1. **Set your breakpoints in VSCode** (click left of line numbers in the files mentioned above)
2. **Open terminal in VSCode** (Ctrl+` or View → Terminal)
3. **Run from terminal:**
   ```bash
   cd /Users/adna/Desktop/ecommerce/services/inventory
   mirrord exec -f ../../.mirrord/db-branching.json -- go run .
   ```
4. **Breakpoints work automatically** - VSCode attaches to the debugger when you run `go run`
5. **Don't click "Run and Debug" button** - just use the terminal command above

**Note:** Both methods work. The extension is more convenient (no terminal needed), but terminal works fine too. You don't need launch.json with either method.

**Terminal 2: Open Frontend**
```bash
# Get frontend URL
minikube service frontend -n metalmart
# Or port-forward:
kubectl port-forward -n metalmart svc/frontend 3000:80
```

**In the browser:**
1. Open frontend (http://localhost:3000)
2. **Show initial stock** - Click on a product (e.g., Product 1)
   - **Breakpoint at line 24 hits** - Show in Variables panel: `productID = "1"`
   - Point out: "This is reading from my branched database"
   - Continue (F5) - Stock shown comes from your BRANCHED database (same as cluster initially)

3. **Make a change** - Go through checkout and place an order
   - **Breakpoint at line 41 hits** - Show in Variables panel: `req.Items` with product and quantity
   - Point out: "This reservation is happening in my branch"
   - Continue (F5) - **Breakpoint at line 112 hits** - Database transaction starting
   - Continue (F5) - **Breakpoint at line 137 hits** - Show the UPDATE query
   - Point out: "This UPDATE is writing to my branched database, not production"
   - Continue (F5) - Reservation complete

4. **Refresh product page** - Stock should be reduced
   - **Breakpoint at line 24 hits again** - Show the reduced stock
   - This shows YOUR branch has changed

5. **Verify cluster is unchanged** - In another terminal:
   ```bash
   kubectl exec -n metalmart deployment/inventory -- \
     curl -s http://localhost:8082/api/inventory/1
   ```
   - Cluster still shows original stock!

**What to show:**
- Frontend shows stock from your branched database
- Changes you make are visible in the frontend
- Cluster database stays untouched (verify with kubectl)

---

**Option B: Command Line (For testing/verification)**

**Terminal 1: Run with Database Branching**
```bash
cd /Users/adna/Desktop/ecommerce/services/inventory

# This creates an isolated database branch
mirrord exec -f ../../.mirrord/db-branching.json -- go run .
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
- Verify `.mirrord/queue-splitting.json` has `".*@metalbear.com"`
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

- Check filter in `.mirrord/queue-splitting.json`
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
mirrord exec -f ../../.mirrord/queue-splitting.json -- go run .

# Test database branching
cd services/inventory
mirrord exec -f ../../.mirrord/db-branching.json -- go run .

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
