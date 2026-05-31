# order-service

HTTP service that manages orders. Reads from DB, enriches with stock data from inventory-service.

## Architecture
```
Client → GET /orders → order-service (DB) → inventory-service (GET /stock)
```

## Build & Run

### SQLite mode (no external DB needed)
```bash
# Start inventory-service on :8081 first
export DB_DRIVER=sqlite3
export INVENTORY_SERVICE_URL=http://localhost:8081
go run main.go
```

### Postgres mode
```bash
export DB_HOST=localhost DB_PORT=5432 DB_USER=postgres DB_PASSWORD=postgres DB_NAME=orders
export INVENTORY_SERVICE_URL=http://localhost:8081
go run main.go
```

Listens on `:8080` (override with `PORT` env var).

## Endpoints
- `GET /orders` — returns all orders enriched with stock info from inventory-service
- `GET /health` — health check

## Environment Variables
| Variable | Default | Description |
|----------|---------|-------------|
| DB_DRIVER | postgres | Database driver: `postgres` or `sqlite3` |
| DB_PATH | /tmp/orders.db | SQLite file path (only when DB_DRIVER=sqlite3) |
| DB_HOST | localhost | Postgres host |
| DB_PORT | 5432 | Postgres port |
| DB_USER | postgres | Postgres user |
| DB_PASSWORD | postgres | Postgres password |
| DB_NAME | orders | Database name |
| INVENTORY_SERVICE_URL | http://localhost:8081 | inventory-service base URL |
| PORT | 8080 | HTTP listen port |

## Database
Auto-creates `orders` table and seeds 4 rows on startup if empty.

## Testing with order-integration-tests
```bash
# 1. Start inventory-service: go run main.go (in inventory-service dir)
# 2. Start order-service: DB_DRIVER=sqlite3 INVENTORY_SERVICE_URL=http://localhost:8081 go run main.go
# 3. Run tests: ORDER_SERVICE_URL=http://localhost:8080 go test -v ./... (in order-integration-tests dir)
```
