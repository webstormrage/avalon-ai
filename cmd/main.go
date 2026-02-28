package main

import (
	"avalon/pkg/operouter"
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

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		log.Printf(
			"➡️  %s %s from %s",
			r.Method,
			r.URL.Path,
			r.RemoteAddr,
		)

		lrw := &loggingResponseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK,
		}

		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		log.Printf(
			"⬅️  %s %s | status=%d | duration=%s",
			r.Method,
			r.URL.Path,
			lrw.statusCode,
			duration,
		)
	})
}

func main() {
	_ = godotenv.Load()
	dsn := os.Getenv("DATA_SOURCE_NAME")
	// apiKey := os.Getenv("GEMINI_API_KEY")
	mediaDir := os.Getenv("MEDIA_DIR")
	openRouterUri := os.Getenv("OPEN_ROUTER_URI")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	agent, err := operouter.NewAgent(ctx, openRouterUri)
	if err != nil {
		log.Fatal(err)
	}

	if err := store.RunInitMigration(ctx, db); err != nil {
		log.Fatal(err)
	}

	handler := &server.GameHandler{
		DB:       db,
		Agent:    agent,
		Ctx:      ctx,
		MediaDir: mediaDir,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/games/new", handler.CreateGame)
	mux.HandleFunc("/games/state", handler.GetGameState)
	mux.HandleFunc("/games/next-tick", handler.NextTick)

	loggedMux := LoggingMiddleware(mux)

	server := &http.Server{
		Addr:    ":8080",
		Handler: loggedMux,
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
