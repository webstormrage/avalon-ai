package main

import (
	"avalon/pkg/server"
	"avalon/pkg/store"
	"context"
	"database/sql"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

/*func main() {
	ctx := context.Background()
	_ = godotenv.Load()
	apiKey := os.Getenv("GEMINI_API_KEY")
	dataSource := os.Getenv("DATA_SOURCE_NAME")
}*/

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATA_SOURCE_NAME")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := store.RunInitMigration(ctx, db); err != nil {
		log.Fatal(err)
	}

	handler := &server.GameHandler{DB: db}

	mux := http.NewServeMux()
	mux.HandleFunc("/games/new", handler.CreateGame)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		/*ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  30 * time.Second,*/
	}

	go func() {
		log.Println("HTTP server started on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal(err)
	}

	log.Println("Server stopped")
}
