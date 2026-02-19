package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	httpSwagger "github.com/swaggo/http-swagger"

	h "practice-3/internal/handler/http"
	"practice-3/internal/middleware"
	"practice-3/internal/repository/_postgres"
	"practice-3/internal/usecase"
	"practice-3/pkg/modules"

	_ "practice-3/docs"
)

func Run() error {
	_ = godotenv.Load()

	cfg := &modules.PostgreConfig{
		Host:     getenv("DB_HOST", "localhost"),
		Port:     getenv("DB_PORT", "5432"),
		Username: getenv("DB_USER", "ak1ra"),
		Password: getenv("DB_PASS", "Ak1raWC3"),
		DBName:   getenv("DB_NAME", "godb"),
		SSLMode:  getenv("DB_SSLMODE", "disable"),
	}

	apiKey := getenv("API_KEY", "supersecret")

	ctx := context.Background()
	pg := _postgres.NewPGDialect(ctx, cfg)
	repos := _postgres.NewRepositories(pg)
	uc := usecase.NewUserUsecase(repos.UserRepository)
	handler := h.NewUserHandler(uc)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger)

	r.Get("/swagger/*", httpSwagger.WrapHandler)
	r.Get("/health", handler.Health)

	r.Route("/", func(pr chi.Router) {
		pr.Use(middleware.APIKeyAuth(apiKey))

		pr.Get("/users", handler.GetUsers)
		pr.Get("/users/{id}", handler.GetUserByID)
		pr.Post("/users", handler.CreateUser)
		pr.Put("/users/{id}", handler.UpdateUser)
		pr.Delete("/users/{id}", handler.DeleteUser)
	})

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		log.Println("server started on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_ = pg.DB.Close()
	return srv.Shutdown(shCtx)
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
