package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/robinLudan/go-auth-example/internal/api"
	"github.com/robinLudan/go-auth-example/internal/storage"
)

func main() {
	loadEnvs()

	db, err := sql.Open("sqlite3", "user.db")
	if err != nil {
		log.Fatal(err)
	}

	store := storage.NewSQLite(db)
	store.CreateTables()

	handler := api.NewApiServer(store)
	svr := &http.Server{
		Addr:    ":" + os.Getenv("PORT"),
		Handler: handler,
	}

	// graceful shutdown the server
	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		signal.Notify(sigint, syscall.SIGTERM)
		<-sigint

		log.Println("Service interrupt received")
		if err := svr.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown error: %v", err)
		}
		close(idleConnsClosed)
	}()

	log.Println("Starting server on port", svr.Addr)
	if err := svr.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server failed to start: %v", err)
	}

	<-idleConnsClosed
	log.Println("Service stopped")
}

func loadEnvs() {
	// checks if required envs have been set
	if err := godotenv.Load(); err != nil {
		log.Fatal(err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT env not set")
	}

	key := os.Getenv("JWT_KEY")
	if key == "" {
		log.Fatal("JWT_KEY env not set")
	}
}
