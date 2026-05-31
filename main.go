package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

type Order struct {
	ID        int       `json:"id"`
	ProductID string    `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type OrderWithStock struct {
	Order
	ProductName    string `json:"product_name,omitempty"`
	StockAvailable int    `json:"stock_available,omitempty"`
	Warehouse      string `json:"warehouse,omitempty"`
}

var db *sql.DB

func initDB() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_HOST", "localhost"),
		getEnv("DB_PORT", "5432"),
		getEnv("DB_USER", "postgres"),
		getEnv("DB_PASSWORD", "postgres"),
		getEnv("DB_NAME", "orders"),
	)
	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err = db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	createTable()
	seedData()
	log.Println("database connected and seeded")
}

func createTable() {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			product_id VARCHAR(50) NOT NULL,
			quantity INT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Fatalf("failed to create table: %v", err)
	}
}

func seedData() {
	var count int
	db.QueryRow("SELECT COUNT(*) FROM orders").Scan(&count)
	if count > 0 {
		return
	}
	orders := []struct {
		productID string
		quantity  int
		status    string
	}{
		{"PROD-001", 5, "confirmed"},
		{"PROD-001", 3, "pending"},
		{"PROD-002", 10, "pending"},
		{"PROD-003", 1, "confirmed"},
	}
	for _, o := range orders {
		db.Exec("INSERT INTO orders (product_id, quantity, status) VALUES ($1, $2, $3)",
			o.productID, o.quantity, o.status)
	}
}

func handleGetOrders(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, product_id, quantity, status, created_at FROM orders ORDER BY id")
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error":"%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	inventoryURL := getEnv("INVENTORY_SERVICE_URL", "http://localhost:8081")
	var results []OrderWithStock
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.ProductID, &o.Quantity, &o.Status, &o.CreatedAt); err != nil {
			continue
		}
		enriched := OrderWithStock{Order: o}
		stock, err := fetchStock(inventoryURL, o.ProductID)
		if err == nil {
			enriched.ProductName = stock.ProductName
			enriched.StockAvailable = stock.Quantity
			enriched.Warehouse = stock.Warehouse
		}
		results = append(results, enriched)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

type StockInfo struct {
	ProductName string `json:"product_name"`
	Quantity    int    `json:"quantity"`
	Warehouse   string `json:"warehouse"`
}

func fetchStock(baseURL, productID string) (*StockInfo, error) {
	resp, err := http.Get(fmt.Sprintf("%s/stock?product_id=%s", baseURL, productID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("inventory service returned %d", resp.StatusCode)
	}
	body, _ := io.ReadAll(resp.Body)
	var stock StockInfo
	if err := json.Unmarshal(body, &stock); err != nil {
		return nil, err
	}
	return &stock, nil
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok","service":"order-service"}`))
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	initDB()
	port := getEnv("PORT", "8080")
	http.HandleFunc("/orders", handleGetOrders)
	http.HandleFunc("/health", handleHealth)
	log.Printf("order-service listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
