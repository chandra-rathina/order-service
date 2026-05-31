# order-service

HTTP service that manages orders. Reads from Postgres, enriches with stock data from inventory-service.

## Architecture
```
Client → GET /orders → order-service (Postgres) → inventory-service (GET /stock)
```

## Build & Run
```bash
# Start dependencies first:
# 1. Postgres on :5432 (auto-creates tables and seeds data)
# 2. inventory-service on :8081

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
| DB_HOST | localhost | Postgres host |
| DB_PORT | 5432 | Postgres port |
| DB_USER | postgres | Postgres user |
| DB_PASSWORD | postgres | Postgres password |
| DB_NAME | orders | Database name |
| INVENTORY_SERVICE_URL | http://localhost:8081 | inventory-service base URL |
| PORT | 8080 | HTTP listen port |

## Database
Auto-creates `orders` table and seeds 4 rows on startup if empty.
