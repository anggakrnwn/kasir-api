package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/anggakrnwn/kasir-api/config"
	"github.com/anggakrnwn/kasir-api/database"
	"github.com/anggakrnwn/kasir-api/handlers"
	"github.com/anggakrnwn/kasir-api/middlewares"
	"github.com/anggakrnwn/kasir-api/repositories"
	"github.com/anggakrnwn/kasir-api/services"
)

func main() {

	cfg, err := config.Init()
	if err != nil {
		log.Fatal("failed to load config:", err)
	}

	// Setup database
	db, err := database.InitDB(cfg.Database.ConnectionString)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	apiKeyMiddleware := middlewares.APIKey(cfg.Auth.APIKey)

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService)
	transactionRepo := repositories.NewTransactionRepository(db)
	transactionService := services.NewTransactionService(transactionRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	// setup routes
	mux := http.NewServeMux()

	mux.HandleFunc("/health", healthHandler(cfg))
	mux.HandleFunc("/", homeHandler(cfg))

	mux.HandleFunc("/api/product", apiKeyMiddleware(productHandler.HandleProduct))
	mux.HandleFunc("/api/product/", apiKeyMiddleware(productHandler.HandleProductByID))
	mux.HandleFunc("/api/checkout", apiKeyMiddleware(transactionHandler.HandleCheckout))
	mux.HandleFunc("/api/report/hari-ini", apiKeyMiddleware(transactionHandler.GetReport))
	mux.HandleFunc("/api/report", apiKeyMiddleware(transactionHandler.HandleReport))

	addr := "0.0.0.0:" + cfg.Server.Port
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("server running on %s (%s mode)", addr, cfg.Env)
		log.Printf("api documentation: http://localhost:%s/", cfg.Server.Port)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("server error:", err)
		}
	}()

	<-ctx.Done()
	log.Println("\n shutting down server..")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("server forced to shutdown:", err)
	}

	log.Println("server stopped gracefully")
}

func healthHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"service":   cfg.App.Name,
			"version":   cfg.App.Version,
			"env":       cfg.Env,
			"timestamp": time.Now().Format(time.RFC3339),
		})
	}
}

func homeHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		fmt.Fprintf(w, "=================================================\n")
		fmt.Fprintf(w, "      %s v%s (%s)\n", cfg.App.Name, cfg.App.Version, cfg.Env)
		fmt.Fprintf(w, "=================================================\n\n")
		fmt.Fprintf(w, "ENDPOINTS:\n")
		fmt.Fprintf(w, "  GET    /health              Health check\n")
		fmt.Fprintf(w, "  GET    /api/product         List products\n")
		fmt.Fprintf(w, "  POST   /api/product         Create product\n")
		fmt.Fprintf(w, "  GET    /api/product/{id}    Get product by ID\n")
		fmt.Fprintf(w, "  PUT    /api/product/{id}    Update product\n")
		fmt.Fprintf(w, "  DELETE /api/product/{id}    Delete product\n")
		fmt.Fprintf(w, "  POST   /api/checkout        Checkout transaction\n")
		fmt.Fprintf(w, "  GET    /api/report/hari-ini Today's sales report\n")
		fmt.Fprintf(w, "  GET    /api/report          Sales report by date\n\n")
		fmt.Fprintf(w, "=================================================\n")
	}
}
