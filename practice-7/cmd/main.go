package main

import (
	"log"
	v1 "practice-7/internal/controller/http/v1"
	"practice-7/internal/entity"
	"practice-7/internal/usecase"
	"practice-7/internal/usecase/repo"
	"practice-7/pkg/postgres"

	"github.com/gin-gonic/gin"
)

func main() {
	pg, err := postgres.New()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	pg.Conn.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"")

	err = pg.Conn.AutoMigrate(&entity.User{})
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	userRepo := repo.NewUserRepo(pg)
	userUsecase := usecase.New(userRepo)

	r := gin.Default()
	
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	v1.New(r, userUsecase)

	log.Println("Server starting on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}