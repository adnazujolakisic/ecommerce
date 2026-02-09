# Future Enhancements for Order Status & Delivery

## Current State
- Orders progress: `pending` → `processing` → `confirmed` → `shipped`
- "Delivered" status exists in UI but is never reached
- Order processor stops at "shipped"

---

## Third-Party Shipping Integration with AWS SQS

**Goal:** Integrate with a third-party shipping provider that uses AWS SQS for delivery notifications, similar to how many large companies handle shipping updates.

### Architecture

```
Order Service → Publishes "shipped" event to Kafka
                ↓
Shipping Provider → Receives order, ships package
                ↓
AWS SQS → Receives delivery notifications from shipping provider
                ↓
New Shipping Service → Consumes from SQS, updates order status
```

### Implementation Plan

**1. New Service: `shipping-service`**
- Deployed in AWS (separate from main cluster)
- Consumes messages from AWS SQS queue
- Receives delivery notifications from shipping provider
- Updates order status via Order Service API

**2. SQS Queue Setup**
- Create SQS queue: `metalmart-delivery-notifications`
- Configure shipping provider to send delivery events to SQS
- Message format:
  ```json
  {
    "order_id": "12345",
    "tracking_number": "1Z999AA10123456784",
    "status": "delivered",
    "delivered_at": "2026-02-05T14:30:00Z",
    "signature": "John Doe"
  }
  ```