# Order Queue Pipeline Demo

A relaxed walkthrough for showing queue messaging with Kafka and mirrord.

---

## Quick Overview: Kafka

Kafka is a message queue — think of it like a mailbox. Services **produce** messages (drop letters in) and **consume** them (pick them up) when they're ready. No one has to wait. The API can respond right away; the heavy work happens in the background.

In our flow: the Order Service publishes *"order created"* to Kafka and moves on. The Order Processor picks it up asynchronously and updates the status. Simple, reliable, scalable.

---

## The Demo: Let's Walk Through It

We'll go **frontend → order → Kafka → processor**. You'll see the message flow end-to-end.

### Before We Start

1. **Start the frontend:** `cd frontend && npm run dev`
2. **Run the Order Processor with mirrord** (queue-splitting.json) — that's the one that consumes from Kafka. Order Service stays in the cluster.
3. **Add these breakpoints** in VS Code:

### Breakpoints (and what each shows)

| # | Where | What you'll show the client |
|---|-------|----------------------------|
| 1 | `services/order/handlers/handlers.go` — **Line 29** `h.store.CreateOrder(req)` | *"The order just landed at our API and got saved to the database."* |
| 2 | `services/order/handlers/handlers.go` — **Line 44** `h.producer.PublishOrderCreated(event)` | *"We're publishing to Kafka. The API is done — the customer gets their confirmation immediately."* |
| 3 | `services/order-processor/main.go` — **Line 109** `log.Printf("Received message...")` | *"The processor picked up the message from Kafka. No one waited — it's async."* |
| 4 | `services/order-processor/main.go` — **Line 164** `p.updateOrderStatus(...)` | *"Each status transition: processing → confirmed → shipped. Watch it step through."* |

### The Walkthrough

1. **Go to the frontend** — browse, add to cart.
2. **Checkout** — use **`demo@metalbear.com`** for mirrord queue-splitting to work.
3. **Submit the order** → Breakpoint 1 hits. *"Order just arrived."*
4. **Continue** → Breakpoint 2 hits. *"Publishing to Kafka — API responds, we're done."*
5. **Frontend shows Order Confirmation** — customer sees order number and tracking link right away.
6. **Continue** → Breakpoint 3 hits. *"Processor got the message."*
7. **Continue** → Breakpoint 4 hits multiple times — each status step.
8. **Open the tracking link** — the status bar fills in live as the processor works.

### Nice touches to mention

- **Right after breakpoint 2:** The confirmation page is already there — the API didn't block on processing.
- **Stay on breakpoint 3:** The tracking page stays on "pending." The message is in Kafka, waiting.
- **Resume:** Status updates flow through without any refresh.

---

## Flow at a Glance

```
Customer → Frontend → Order Service → Kafka → Order Processor → Status updates
```

Order Service publishes; Order Processor consumes. Mirrord lets you run the processor locally while Kafka and the rest stay in the cluster.

---

## CLI: See mirrord's temp Kafka topics

```bash
kubectl exec -n metalmart deploy/kafka -- kafka-topics --bootstrap-server localhost:9092 --list | grep mirrord-tmp
```

Topics like `mirrord-tmp-xxx-order.created` are the extra topics mirrord creates for queue splitting.
