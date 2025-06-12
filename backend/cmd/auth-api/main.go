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

	"agent-connector/api/auth"
	"agent-connector/config"
	"agent-connector/internal"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Starting Authentication API Server...\n")
	fmt.Printf("Server: %s\n", cfg.GetServiceAddr("auth"))
	fmt.Printf("Database: %s://%s:%d/%s\n", cfg.Database.Driver, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)
	fmt.Printf("Environment: %s\n", cfg.App.Environment)

	// Initialize database
	if err := internal.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin engine
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	if cfg.API.EnableCORS {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = []string{cfg.API.AllowedOrigins}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"*"}
		corsConfig.ExposeHeaders = []string{"*"}
		corsConfig.AllowCredentials = true
		router.Use(cors.New(corsConfig))
	}

	// Set up routes
	auth.SetupAuthRoutes(router)

	// Root path
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     cfg.App.Name + " Auth API",
			"version":     cfg.App.Version,
			"description": "Agent Connector Authentication API",
			"status":      "running",
			"environment": cfg.App.Environment,
			"timestamp":   time.Now().Unix(),
			"endpoints":   "/api/v1/auth/",
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServiceAddr("auth"),
		Handler:      router,
		ReadTimeout:  cfg.Services.AuthAPI.ReadTimeout,
		WriteTimeout: cfg.Services.AuthAPI.WriteTimeout,
		IdleTimeout:  cfg.Services.AuthAPI.IdleTimeout,
	}

	// Start server
	go func() {
		fmt.Printf("Authentication API Server running on http://%s\n", cfg.GetServiceAddr("auth"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down Authentication API Server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Authentication API Server gracefully stopped")
	}
}
