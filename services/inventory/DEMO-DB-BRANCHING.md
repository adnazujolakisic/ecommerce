# Database Branching Demo with mirrord

## The Problem

When debugging locally with mirrord, you're connected to the **shared cluster database**. Any changes you make affect everyone.

## The Solution

With **database branching**, mirrord creates an isolated copy of your database. You can break things freely - the shared database stays untouched.

---

## Demo: Toggle Between Shared vs Branched Database

### Config 1: Shared Database (No Branching)

Use `.mirrord/inventory.json`:

```json
{
  "target": {
    "path": { "deployment": "inventory" },
    "namespace": "metalmart"
  },
  "feature": {
    "network": { "incoming": "mirror", "outgoing": true },
    "fs": "local",
    "env": true
  }
}
```

**Run locally → Check inventory:**
```bash
curl http://localhost:8082/api/inventory/1
# Returns: {"product_id":"1", "stock_quantity": 100, ...}
```

**Make a change:**
```bash
curl -X POST http://localhost:8082/api/inventory/reserve \
  -H "Content-Type: application/json" \
  -d '{"items": [{"productId": "1", "quantity": 50}]}'
```

**Result:** Stock is now 50 - **and so is everyone else's!** You modified the shared database.

---

### Config 2: Branched Database (Isolated)

Switch to `.mirrord/inventory-db-branch.json`:

```json
{
  "target": {
    "path": { "deployment": "inventory" },
    "namespace": "metalmart"
  },
  "feature": {
    "network": { "incoming": "mirror", "outgoing": true },
    "fs": "local",
    "env": true
  },
  "operator": true,
  "db_branches": [
    {
      "id": "my-debug-session",
      "type": "pg",
      "version": "16",
      "name": "inventory",
      "ttl_secs": 300,
      "connection": {
        "url": { "type": "env", "variable": "DATABASE_URL" }
      },
      "copy": { "mode": "all" }
    }
  ]
}
```

**Run locally → Check inventory:**
```bash
curl http://localhost:8082/api/inventory/1
# Returns: {"product_id":"1", "stock_quantity": 100, ...}
# (Fresh copy from the branched database)
```

**Make a destructive change:**
```bash
curl -X POST http://localhost:8082/api/inventory/reserve \
  -H "Content-Type: application/json" \
  -d '{"items": [{"productId": "1", "quantity": 99}]}'
```

**Check your branch:**
```bash
curl http://localhost:8082/api/inventory/1
# Returns: {"stock_quantity": 1, "reserved_quantity": 99, ...}
```

**Check shared database (via cluster):**
```bash
# Shared database still has stock_quantity: 100
# Your changes are isolated to YOUR branch
```

---

## Key Difference

| Config | Database | Changes affect |
|--------|----------|----------------|
| `inventory.json` | Shared cluster DB | Everyone |
| `inventory-db-branch.json` | Isolated branch | Only you |

---

## What the `db_branches` Config Does

| Field | Purpose |
|-------|---------|
| `id` | Branch identifier (reuse to reconnect) |
| `type: "pg"` | PostgreSQL |
| `name` | Database name to branch |
| `ttl_secs` | Auto-delete after N seconds (max 900) |
| `copy.mode` | `"all"` = full copy, `"schema"` = structure only, `"empty"` = blank |

---

## Cleanup

Branches auto-expire after `ttl_secs`. Or manually:

```bash
mirrord db-branches -n metalmart destroy my-debug-session
```

---

## Config Files

| File | Use Case |
|------|----------|
| `.mirrord/inventory.json` | Normal development (shared DB) |
| `.mirrord/inventory-db-branch.json` | Debugging/testing (isolated DB) |
