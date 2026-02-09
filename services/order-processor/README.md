# Order Processor Service

A Kafka consumer service that processes orders from the `order.created` topic.

## Overview

This service:
- Consumes order events from Kafka
- Updates order status through the order service API (pending → processing → confirmed → shipped)
- Runs as a deployment in Kubernetes

## Local Development with mirrord

mirrord allows you to debug this Kafka consumer locally while connected to your Kubernetes cluster, using **queue splitting** to route specific messages to your local machine.

### Prerequisites

1. **mirrord operator** installed with Kafka splitting enabled:
   ```bash
   helm install mirrord-operator mirrord/mirrord-operator \
     --set license.key=<YOUR_LICENSE_KEY> \
     --set operator.kafkaSplitting=true \
     -n mirrord --create-namespace
   ```

2. **Kubernetes CRDs** applied:
   ```bash
   kubectl apply -f k8s/base/infrastructure/mirrord-kafka.yaml
   ```

### How Queue Splitting Works

1. **Operator intercepts Kafka** - The mirrord operator consumes from the original `order.created` topic
2. **Creates temporary topics** - For each debug session, mirrord creates:
   - `mirrord-tmp-<session>-order.created` - Messages for your local app
   - `mirrord-tmp-<session>-fallback-order.created` - Messages for the cluster
3. **Filters on headers** - Messages are routed based on **Kafka headers** (not body)
4. **Exclusive routing** - Each message goes to exactly one consumer

### Configuration Files

#### `.mirrord/order-processor.json`
```json
{
  "target": {
    "path": { "deployment": "order-processor" },
    "namespace": "metalmart"
  },
  "feature": {
    "network": {
      "incoming": "mirror",
      "outgoing": { "tcp": true, "udp": true },
      "dns": true
    },
    "fs": "local",
    "env": true,
    "split_queues": {
      "order-created": {
        "queue_type": "Kafka",
        "message_filter": {
          "customer_email": ".*"
        }
      }
    }
  },
  "operator": true
}
```

**Key settings:**
- `split_queues.order-created` - References the topic ID from `MirrordKafkaTopicsConsumer`
- `message_filter` - Regex patterns matched against **Kafka headers**
- `"customer_email": ".*"` - Matches any message with a `customer_email` header

#### `k8s/base/infrastructure/mirrord-kafka.yaml`
```yaml
apiVersion: queues.mirrord.metalbear.co/v1alpha
kind: MirrordKafkaClientConfig
metadata:
  name: kafka-client-config
  namespace: mirrord
spec:
  properties:
    - name: bootstrap.servers
      value: kafka.metalmart.svc.cluster.local:9092
---
apiVersion: queues.mirrord.metalbear.co/v1alpha
kind: MirrordKafkaTopicsConsumer
metadata:
  name: order-processor-consumer
  namespace: metalmart
spec:
  consumerApiVersion: apps/v1
  consumerKind: Deployment
  consumerName: order-processor
  topics:
    - id: order-created
      clientConfig: kafka-client-config
      nameSources:
        - directEnvVar:
            container: order-processor
            variable: KAFKA_TOPIC
      groupIdSources:
        - directEnvVar:
            container: order-processor
            variable: KAFKA_GROUP_ID
```

### Important: Kafka Headers

**mirrord filters on Kafka headers, NOT message body.**

The order service must include routing headers when producing messages:

```go
// services/order/kafka/producer.go
msg := &sarama.ProducerMessage{
    Topic: "order.created",
    Key:   sarama.StringEncoder(event.OrderID),
    Value: sarama.ByteEncoder(data),
    Headers: []sarama.RecordHeader{
        {
            Key:   []byte("customer_email"),
            Value: []byte(event.CustomerEmail),
        },
    },
}
```

### Debugging Steps

1. **Start minikube tunnel** (in a separate terminal):
   ```bash
   minikube tunnel
   ```

2. **Port forward frontend** (optional, for placing orders):
   ```bash
   kubectl port-forward -n metalmart svc/frontend 3000:80
   ```

3. **Open VSCode** and set a breakpoint at line 127 in `main.go`:
   ```go
   func (p *OrderProcessor) ProcessOrder(event OrderCreatedEvent) error {
   ```

4. **Start debug session** with mirrord:
   - Use the VSCode debug configuration
   - mirrord will use `.mirrord/order-processor.json`

5. **Wait for connection**:
   ```
   Order processor started, waiting for messages...
   ```

6. **Place an order** at http://localhost:3000

7. **Breakpoint hits** - You can now:
   - Inspect the `event` variable
   - Step through code (F10/F11)
   - Examine order processing logic

### Filter Examples

**Match specific email:**
```json
"message_filter": {
  "customer_email": "test@example.com"
}
```

**Match email pattern:**
```json
"message_filter": {
  "customer_email": ".*@metalbear.com"
}
```

**Match any message with header:**
```json
"message_filter": {
  "customer_email": ".*"
}
```

### Troubleshooting

**Messages going to fallback instead of local session:**
- Ensure Kafka headers are being set by the producer
- Check header name matches filter key exactly
- Verify debug session is running before placing order

**Connection refused to localhost:9092:**
- mirrord routes traffic through the cluster
- Ensure `env: true` is set to get cluster env vars
- Check Kafka advertised listener uses FQDN: `kafka.metalmart.svc.cluster.local:9092`

**View Kafka messages:**
```bash
kubectl exec -n metalmart deployment/kafka -- \
  kafka-console-consumer --bootstrap-server localhost:9092 \
  --topic order.created --from-beginning --property print.headers=true
```

**Check consumer groups:**
```bash
kubectl exec -n metalmart deployment/kafka -- \
  kafka-consumer-groups --bootstrap-server localhost:9092 \
  --describe --group order-processor
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| KAFKA_BROKERS | localhost:9092 | Kafka broker addresses |
| KAFKA_TOPIC | order.created | Topic to consume from |
| KAFKA_GROUP_ID | order-processor | Consumer group ID |
| ORDER_SERVICE_URL | http://localhost:8084 | Order service API URL |

## Architecture

```
┌─────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   Frontend  │────▶│  Order Service  │────▶│     Kafka       │
└─────────────┘     └─────────────────┘     └────────┬────────┘
                                                     │
                                            order.created topic
                                                     │
                    ┌────────────────────────────────┼────────────────────────────────┐
                    │                                │                                │
                    ▼                                ▼                                ▼
            ┌───────────────┐              ┌─────────────────┐              ┌─────────────────┐
            │ mirrord       │              │ tmp-session     │              │ tmp-fallback    │
            │ operator      │─────────────▶│ topic           │              │ topic           │
            │ (intercepts)  │              └────────┬────────┘              └────────┬────────┘
            └───────────────┘                       │                                │
                                                    ▼                                ▼
                                            ┌───────────────┐              ┌─────────────────┐
                                            │ Local Debug   │              │ Cluster         │
                                            │ Session       │              │ order-processor │
                                            └───────────────┘              └─────────────────┘
```

---

## Demo Script

### The Problem

*"Debugging Kafka consumers in a microservices environment is notoriously difficult. Traditionally, you have three bad options:*

1. **Add print statements** - Deploy, wait, check logs, repeat. Slow and painful.
2. **Run everything locally** - Spin up Kafka, Zookeeper, all dependent services... complex and doesn't match production.
3. **Use a shared dev environment** - But then your debugging interferes with your teammates' work."

### The Solution

*"With mirrord and queue splitting, we can debug our Kafka consumer locally while connected to the real Kubernetes cluster - without disrupting anyone else."*

### Live Demo Flow

#### 1. Show the Architecture
*"Here's our e-commerce app. When a customer places an order, the Order Service publishes a message to Kafka. The Order Processor consumes that message and updates the order status."*

#### 2. Start the Debug Session
*"I'll start my local debugger with mirrord. Notice I'm setting a breakpoint right where we process the order."*

```
[Start VSCode debug session with mirrord]
[Show: "Order processor started, waiting for messages..."]
```

*"My local app is now connected to the cluster's Kafka - but here's the magic: mirrord created a temporary topic just for my session."*

#### 3. Place an Order
*"Let me place an order on the frontend..."*

```
[Go to localhost:3000, add item to cart, checkout]
```

#### 4. Hit the Breakpoint
*"Boom! The breakpoint hit."*

**[PAUSE - This is the key moment]**

*"Let me explain what just happened. I placed an order on the frontend, which runs in the cluster. That order went through the Order Service in the cluster, which published a Kafka message to the cluster's Kafka. And somehow... that message ended up here, in my local debugger, on my laptop."*

*"Look at this data - this is real. Real order ID, real customer email, real total amount. This isn't a mock. This isn't a test fixture. This is a live message from the production-like environment."*

```
[Show Variables panel with event data]
[Point to: CustomerEmail, OrderID, TotalAmount]
```

#### 5. Explain How the Message Got Here

*"So how did a message from the cluster's Kafka end up in my local debugger? This is where mirrord's queue splitting comes in."*

**[Draw or show this flow]**

```
Frontend (cluster)
    ↓
Order Service (cluster)
    ↓
Kafka "order.created" topic (cluster)
    ↓
┌─────────────────────────────────┐
│   mirrord operator (cluster)    │  ← Intercepts ALL messages
│                                 │
│   Checks message HEADERS:       │
│   "customer_email: adna@..."    │
│                                 │
│   Matches my filter? ─────YES───┼──→ Route to MY temporary topic
│                       │         │         ↓
│                       NO        │    My laptop (via mirrord)
│                       ↓         │         ↓
│              Fallback topic     │    BREAKPOINT HITS! 🎯
│                   ↓             │
│         Cluster order-processor │
└─────────────────────────────────┘
```

*"The mirrord operator sits between Kafka and all consumers. When I started my debug session, it created a temporary topic just for me. Every message that comes in, it checks the headers against my filter. If it matches - it goes to my laptop. If not - it goes to the normal cluster consumer."*

*"This means:*
- *My orders come to me*
- *Everyone else's orders go to the cluster*
- *No one is disrupted*
- *I get real data to debug with"*

#### 6. Show the Filter
*"The filter is simple - I'm matching on the `customer_email` header. I could filter by user ID, tenant, feature flag - whatever makes sense for your use case."*

```json
"message_filter": {
  "customer_email": ".*@metalbear.com"
}
```

#### 7. Continue Execution
*"I can step through the code, inspect variables, even modify values. When I'm done debugging, I just continue..."*

```
[Press F5 to continue]
[Show order status updating: processing → confirmed → shipped]
```

*"The order processed successfully, hitting the real Order Service API in the cluster."*

### Key Takeaways

*"Let me recap what just happened:*

1. **I placed an order in the cluster** - Frontend, Order Service, Kafka - all running in Kubernetes
2. **The message found its way to my laptop** - mirrord intercepted it based on my filter
3. **I hit a breakpoint with real data** - Not mocks, not test fixtures, a real order
4. **My local code talked back to the cluster** - Status updates hit the real Order API
5. **Nobody else was affected** - The cluster kept processing other orders normally

*The message traveled: Cluster → Kafka → mirrord operator → my laptop → breakpoint*

*This is what cloud-native development should feel like - debug anywhere, disrupt no one."*

### Q&A Prompts

- *"What if two developers filter on the same criteria?"* → First session wins, second gets fallback
- *"Does this work with other message queues?"* → Yes, SQS is also supported
- *"What about production?"* → You can connect to any cluster you have access to, with appropriate RBAC
- *"How do messages know where to go?"* → Kafka headers - lightweight metadata that doesn't require parsing the message body
