package main

import (
	"fmt"
	"go-tweets/internal/config"
	"go-tweets/internal/handlers"
	"go-tweets/internal/repositories"
	"go-tweets/internal/services"
	"go-tweets/pkg/internalsql"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func main() {
	r := gin.Default()
	validate := validator.New()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatal(err)
	}

	db, err := internalsql.ConnectMySQL(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Go tweets API is running",
		})
	})

	userRepository := repositories.NewUserRepository(db)
	postRepository := repositories.NewPostRepository(db)
	commentRepository := repositories.NewCommentRepository(db)

	userService := services.NewUserService(cfg, userRepository)
	postService := services.NewPostService(cfg, postRepository, commentRepository)
	commentService := services.NewCommentService(cfg, commentRepository, postRepository)

	userHandler := handlers.NewUserHandler(r, validate, userService)
	postHandler := handlers.NewPostHandler(r, validate, postService)
	commentHandler := handlers.NewCommentHandler(r, validate, commentService)

	userHandler.RouteList(cfg.JWTSecret)
	postHandler.RouteList(cfg.JWTSecret)
	commentHandler.RouteList(cfg.JWTSecret)

	server := fmt.Sprintf("127.0.0.1:%s", cfg.Port)

	// Start the HTTP server on the specified port.
	if err := r.Run(server); err != nil {
		log.Fatalf("Failed to run the server: %v", err)
	}
}
