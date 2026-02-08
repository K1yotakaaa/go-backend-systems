package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"practice-2/internal/handlers"
	"practice-2/internal/middleware"
	"practice-2/internal/store"
)

func main() {
	st := store.New()

	tasks := &handlers.TasksHandler{Store: st}
	ext := &handlers.ExternalHandler{
		Client: &http.Client{Timeout: 5 * time.Second},
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			tasks.Get(w, r)
		case http.MethodPost:
			tasks.Post(w, r)
		case http.MethodPatch:
			tasks.Patch(w, r)
		case http.MethodDelete:
			tasks.Delete(w, r)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/external/todos", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		ext.Todos(w, r)
	})

	var h http.Handler = mux
	h = middleware.APIKey("secret12345")(h)
	h = middleware.Logging("tasks-api")(h)
	h = middleware.RequestID(h)
	h = middleware.JSONOnly(h)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: h,
	}

	go func() {
		log.Println("server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Println("shutting down...")
	_ = srv.Shutdown(ctx)
	log.Println("bye")
}
