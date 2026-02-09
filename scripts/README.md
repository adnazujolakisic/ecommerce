# Scripts Directory

This directory contains application-dependent scripts and helper utilities.

## Structure

- **`load-generator.go` / `load-generator.sh`** - Load testing tools (app-dependent)
- **`helpers/`** - Development and deployment utility scripts
  - See [helpers/README.md](helpers/README.md) for details

---

# Load Generator Scripts

Tools to generate load for MetalMart demos and testing.

## Quick Start

### Go Version (Recommended)

```bash
cd scripts
go run load-generator.go [rate] [duration]

# Examples:
go run load-generator.go 10 60    # 10 orders/sec for 60 seconds
go run load-generator.go 50 30    # 50 orders/sec for 30 seconds
```

### Shell Script Version

```bash
chmod +x load-generator.sh
./load-generator.sh [rate] [duration]
```

## Environment Variables

- `BASE_URL`: Order service URL (default: `http://localhost:8084`)
  ```bash
  BASE_URL=http://order.metalmart.svc.cluster.local:8084 \
  go run load-generator.go 20 120
  ```

## Recommended Settings for Mirrord Demo

**Medium Load (Recommended):**
```bash
go run load-generator.go 20 120
```
- 20 orders/second
- 2 minutes duration
- ~2,400 total orders
- Good balance of load without overwhelming

**Heavy Load:**
```bash
go run load-generator.go 50 60
```
- 50 orders/second
- 1 minute duration
- ~3,000 total orders
- Stress test scenario

## Enable Demo Mode

For faster processing (600ms vs 10s per order), set `DEMO_MODE=true`:

**Docker Compose:**
```bash
DEMO_MODE=true docker-compose up -d order-processor
```

**Kubernetes:**
```bash
kubectl set env deployment/order-processor -n metalmart DEMO_MODE=true
```

## What It Does

1. Generates orders at specified rate
2. Uses random product IDs from catalogue
3. Creates unique customer emails (filterable for Mirrord)
4. Sends orders directly to order service API
5. Each order triggers Kafka message

## Output

The script shows:
- Configuration (rate, duration, total orders)
- Progress dots (every 10 orders)
- Final statistics (actual rate achieved)
