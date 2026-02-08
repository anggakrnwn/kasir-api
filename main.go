package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/anggakrnwn/kasir-api/database"
	"github.com/anggakrnwn/kasir-api/handlers"
	"github.com/anggakrnwn/kasir-api/repositories"
	"github.com/anggakrnwn/kasir-api/services"
	"github.com/spf13/viper"
)

type Config struct {
	Port   string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
}

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}
	config := Config{
		Port:   viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}
	if config.Port == "" {
		config.Port = "8080"
	}
	// Setup database
	db, err := database.InitDB(config.DBConn)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)
	// setup routes
	http.HandleFunc("/api/product", productHandler.HandleProduct)
	http.HandleFunc("/api/product/", productHandler.HandleProductByID)
	transactionRepo := repositories.NewTransactionRepository(db)
	transactionService := services.NewTransactionService(transactionRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	http.HandleFunc("/api/checkout", transactionHandler.HandleCheckout)
	http.HandleFunc("/api/report/hari-ini", func(w http.ResponseWriter, r *http.Request) {
		transactionHandler.GetReport(w, r)
	})
	http.HandleFunc("/api/report", transactionHandler.HandleReport)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OK",
			"message": "API Running",
		})
	})
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintln(w, "=================================================")
		fmt.Fprintln(w, "               KASIR API ENDPOINTS              ")
		fmt.Fprintln(w, "=================================================")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "GENERAL")
		fmt.Fprintln(w, "GET   /health           - Cek status API")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "PRODUCT MANAGEMENT")
		fmt.Fprintln(w, "GET   /api/product      - Get semua produk (filter: ?name=)")
		fmt.Fprintln(w, "POST  /api/product      - Buat produk baru")
		fmt.Fprintln(w, "GET   /api/product/{id} - Get produk by ID")
		fmt.Fprintln(w, "PUT   /api/product/{id} - Update produk")
		fmt.Fprintln(w, "DELETE /api/product/{id} - Delete produk")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "TRANSACTION")
		fmt.Fprintln(w, "POST  /api/checkout     - Checkout transaksi (multi-item)")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "REPORT")
		fmt.Fprintln(w, "GET   /api/report/hari-ini - Laporan penjualan hari ini")
		fmt.Fprintln(w, "GET   /api/report?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD")
		fmt.Fprintln(w, "                          - Laporan berdasarkan rentang tanggal")
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "=================================================")
		fmt.Fprintln(w, "Contoh Request Checkout:")
		fmt.Fprintln(w, `POST /api/checkout`)
		fmt.Fprintln(w, `Body: {"items": [{"product_id": 1, "quantity": 2}]}`)
		fmt.Fprintln(w, "")
		fmt.Fprintln(w, "Contoh Request Produk:")
		fmt.Fprintln(w, `POST /api/product`)
		fmt.Fprintln(w, `Body: {"name": "Indomie", "price": 3000, "stock": 50}`)
		fmt.Fprintln(w, "=================================================")
	})
	addr := "0.0.0.0:" + config.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      http.DefaultServeMux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	// channel untuk signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	// server di goroutine
	go func() {
		fmt.Println("Server running on " + addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server error:", err)
		}
	}()
	// tunggu signal
	<-sigChan
	fmt.Println("\nShutdown signal received, gracefully shutting down...")
	// graceful shutdown dengan timeout 30 detik
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
	fmt.Println("Server stopped")
}
