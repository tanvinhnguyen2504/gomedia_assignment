package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"property-viewings-service/configs"
	"property-viewings-service/internal"
)

func main() {
	db := connectDB()
	defer db.Close()

	viewingRepo := internal.NewPostgresRepository(db)
	viewingService := internal.NewService(viewingRepo)
	srv := newHTTPServer(viewingService)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go runJobScheduler(ctx, viewingService)
	go runHTTPServer(srv)

	waitForShutdown()

	log.Println("Shutting down...")
	cancel()
	drainHTTPServer(srv)
	log.Println("Server stopped.")
}

func connectDB() *sqlx.DB {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		configs.Getenv("DB_HOST", "localhost"),
		configs.Getenv("DB_PORT", "5432"),
		configs.Getenv("DB_USER", "postgres"),
		configs.Getenv("DB_PASSWORD", "postgres"),
		configs.Getenv("DB_NAME", "viewings_db"),
		configs.Getenv("DB_SSLMODE", "disable"),
	)
	db, err := sqlx.Connect("postgres", dsn)
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(30 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	return db
}

func newHTTPServer(svc internal.Service) *http.Server {
	port := configs.Getenv("SERVER_PORT", "9999")
	return &http.Server{
		Addr:    ":" + port,
		Handler: internal.NewHandler(svc).Routes(),
	}
}

func runHTTPServer(srv *http.Server) {
	log.Printf("listening on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func runJobScheduler(ctx context.Context, svc internal.Service) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			svc.MarkMissedViewings(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func waitForShutdown() {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
}

func drainHTTPServer(srv *http.Server) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
}
