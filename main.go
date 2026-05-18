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
	_ "property-viewings-service/docs"
	"property-viewings-service/internal"
)

func scheduleJob(ctx context.Context, s internal.Service) {
	s.MarkMissedViewings(ctx)
}

func main() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

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
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	repo := internal.NewPostgresRepository(db)
	svc := internal.NewService(repo)
	h := internal.NewHandler(svc)

	port := configs.Getenv("SERVER_PORT", "9999")
	log.Printf("listening on :%s", port)
	log.Printf("swagger UI at http://localhost:%s/swagger/index.html", port)

	go func() {
		if err := http.ListenAndServe(":"+port, h.Routes()); err != nil {
			log.Fatal(err)
		}
	}()

	ctx, cancel := context.WithCancel(context.Background())

	for {
		select {
		case <-ticker.C:
			scheduleJob(ctx, svc)
		case <-quit:
			cancel()
			log.Println("Shutting down server...")
			return
		}
	}
}
