# Order Queue Pipeline Demo

## What It Does

Orders flow through a real-time Kafka message queue:

```
Customer places order → Order Service → Kafka → Order Processor → Status updates
```

1. **Order Service** receives the order via REST API and publishes an event to Kafka
2. **Kafka** holds the message in the `order.created` topic
3. **Order Processor** picks up the event and moves the order through stages:
   `pending → processing → confirmed → shipped`

## Why It Matters

- **Decoupled services** — the order API responds instantly, processing happens in the background
- **Reliability** — if the processor goes down, messages queue up in Kafka and get processed when it's back
- **Scalability** — spin up multiple processors to handle high order volume without touching the API
- **Real-time visibility** — order status updates live as the processor works through each stage

## Live Demo (Frontend → Backend → Queue)

### Setup

**Option A: Local (Minikube / docker-compose)** — services talk to local Postgres + Kafka  
**Option B: Remote GKE** — services use mirrord to connect to cluster Postgres + Kafka

1. **Launch from VS Code debug dropdown:**
   - **Local:** `Full Queue Pipeline (Producer + Consumer)`
   - **GKE (mirrord):** `Full Queue Pipeline (with mirrord)`
2. **Start the frontend:**
   ```bash
   cd frontend && npm run dev
   ```
   - **GKE:** Point to cluster frontend first:
     ```bash
     kubectl port-forward -n metalmart svc/frontend 55587:80
     export VITE_PROXY_TARGET=http://127.0.0.1:55587
     cd frontend && npm run dev
     ```
3. **Set these breakpoints** in VS Code before the demo:

### Recommended Breakpoints

| # | File | Line | What You'll See |
|---|------|------|-----------------|
| 1 | `services/order/handlers/handlers.go` | **Line 29** — `h.store.CreateOrder(req)` | Order being saved to the database |
| 2 | `services/order/handlers/handlers.go` | **Line 44** — `h.producer.PublishOrderCreated(event)` | The message being published to Kafka |
| 3 | `services/order-processor/main.go` | **Line 109** — `log.Printf("Received message...")` | Kafka message arriving at the consumer |
| 4 | `services/order-processor/main.go` | **Line 164** — `p.updateOrderStatus(...)` | Each status transition: processing → confirmed → shipped |

### Walkthrough (2 min)

1. **Open the frontend** in the browser — browse products, add to cart
2. **Go to Checkout** — fill in name, **email (use `demo@metalbear.com` for mirrord queue-splitting)**, shipping address, submit
3. **Hit breakpoint 1** — pause here, show the client: *"The order just arrived at our API"*
4. **Continue** → **Hit breakpoint 2** — *"Now we're publishing it to Kafka — the API is done, customer gets instant response"*
5. **Frontend redirects to Order Confirmation** — show the order number and tracking link
6. **Hit breakpoint 3** — *"The processor picked up the message from Kafka asynchronously"*
7. **Continue** → **Hit breakpoint 4 multiple times** — *"Watch the status update in real-time"*
8. **Click the tracking link** on the confirmation page — the tracking page polls every 2s and the status bar fills up live: `pending → processing → confirmed → shipped`

### Show mirrord-Created Topics (CLI)

When queue splitting is active, mirrord creates temporary Kafka topics. List them:

```bash
# All topics
kubectl exec -n metalmart deploy/kafka -- kafka-topics --bootstrap-server localhost:9092 --list

# Only mirrord temp topics
./scripts/list-kafka-topics.sh
```

You'll see topics like `mirrord-tmp-<id>-order.created` — those are the extra topics mirrord creates.

### Key Moments to Highlight

- **After breakpoint 2**: The frontend already shows confirmation — the API didn't wait for processing
- **Pause the processor** (stay on breakpoint 3): Show that the frontend stays on "pending" — the queue is holding the message
- **Resume**: Watch the tracking page update live without a page refresh

## Architecture Highlights

| Component | Role | Tech |
|-----------|------|------|
| Order Service | REST API + event producer | Go, PostgreSQL, Kafka |
| Kafka | Message broker | Apache Kafka |
| Order Processor | Async consumer + status updater | Go, Sarama |

## Scaling Story

> "What happens when we get 10x more orders?"

- Kafka partitions the `order.created` topic
- We add more processor instances to the `order-processor` consumer group
- Kafka automatically distributes messages across processors
- Zero changes to the Order Service — it just keeps publishing
