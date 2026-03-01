package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/K1yotakaaa/go-backend-systems/practice-4/internal/db"
	"github.com/K1yotakaaa/go-backend-systems/practice-4/internal/movies"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Waiting for database...")

	pool, err := db.Connect(context.Background(), db.FromEnv())
	if err != nil {
		log.Fatalf("DB connect error: %v", err)
	}
	defer pool.Close()

	log.Println("Starting the Server.")

	repo := movies.Repo{DB: pool}
	mh := movies.Handlers{Repo: repo}

	r := chi.NewRouter()
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	r.Get("/whoami", func(w http.ResponseWriter, r *http.Request) {
		host, _ := os.Hostname()
		json.NewEncoder(w).Encode(map[string]any{
			"hostname": host,
			"time":     time.Now().Format(time.RFC3339),
		})
	})
	r.Mount("/movies", mh.Routes())

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("Listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}