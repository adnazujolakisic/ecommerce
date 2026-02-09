# MetalMart Project Context

## What is MetalMart?

MetalMart is a microservices e-commerce application designed as a **demo platform for Mirrord features**, particularly:
- Kafka queue splitting
- Database branching
- Filtering
- Steal mode
- Mirror mode

## Architecture

### Services (Go microservices)
- **catalogue** (8081) - Product catalog
- **inventory** (8082) - Stock management
- **checkout** (8083) - Cart processing (orchestrates checkout flow)
- **order** (8084) - Order creation, publishes to Kafka
- **order-processor** - Kafka consumer, updates order status

### Infrastructure
- **PostgreSQL** - 3 databases: catalogue, inventory, orders (see [DATA-MODEL.md](DATA-MODEL.md))
- **Kafka + Zookeeper** - Async messaging for order processing
- **nginx** - Frontend reverse proxy

### Frontend
- React + TypeScript + Vite
- Port 3000 (Docker) or 80 (K8s)

## Key Design Decisions

1. **No Stripe** - Removed payment processing (was mock anyway)
2. **No Account Service** - Removed for demo simplicity
3. **Synchronous Inventory Commit** - Happens during checkout (not async)
4. **Order Processor** - Only updates order status, doesn't touch inventory

## Checkout Flow (Synchronous)

1. Reserve Inventory → locks stock
2. Create Order → saves order + publishes to Kafka
3. Confirm Inventory → commits stock (reduces stock_quantity)
4. Return Response → user sees confirmation

## Order Processing (Asynchronous)

1. Kafka → Order Processor (consumes event)
2. Order Processor → Order Service (updates status)
   - Status: `pending` → `processing` → `confirmed` → `shipped`

## Important Files

- `DEVELOPMENT.md` - Development guide
- `MIRRORD-SETUP.md` - Mirrord setup instructions
- `QUICK-START.md` - Quick start guide
- `scripts/helpers/` - Deployment helper scripts
- `.mirrord/` - Mirrord configuration files

## Common Issues & Fixes

- **Order status not updating**: Run `./scripts/helpers/fix-order-status.sh`
- **Frontend changes**: Run `./scripts/helpers/rebuild-frontend.sh`
- **Kafka connection errors**: Restart order service

## Deployment

- **Local**: Docker Compose (`docker-compose up`)
- **Minikube**: `./scripts/helpers/start-fresh.sh`
- **GKE**: `./scripts/helpers/deploy-gke.sh`

## Demo Email Filter

For Mirrord Kafka queue splitting demo, orders with `@metalbear.com` emails are filtered to local debugger.
