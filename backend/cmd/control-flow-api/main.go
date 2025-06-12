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

	"agent-connector/api/controlflow"
	"agent-connector/config"
	"agent-connector/internal"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	fmt.Printf("Starting Control Flow API Server...\n")
	fmt.Printf("Server: %s\n", cfg.GetServiceAddr("control"))
	fmt.Printf("Database: %s://%s:%d/%s\n", cfg.Database.Driver, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)
	fmt.Printf("Environment: %s\n", cfg.App.Environment)

	// Initialize database
	if err := internal.InitDatabase(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Set Gin mode
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// CORS configuration
	if cfg.API.EnableCORS {
		corsConfig := cors.DefaultConfig()
		corsConfig.AllowOrigins = []string{cfg.API.AllowedOrigins}
		corsConfig.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
		corsConfig.AllowHeaders = []string{"Origin", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization"}
		corsConfig.AllowCredentials = true
		router.Use(cors.New(corsConfig))
	}

	// Set routes
	controlflow.SetupControlFlowRoutes(router)

	// Root path
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":     cfg.App.Name + " Control Flow API",
			"version":     cfg.App.Version,
			"description": "Agent Connector Control Flow API",
			"status":      "running",
			"environment": cfg.App.Environment,
			"timestamp":   time.Now().Unix(),
			"endpoints":   "/api/v1/controlflow/",
		})
	})

	// Create HTTP server
	server := &http.Server{
		Addr:         cfg.GetServiceAddr("control"),
		Handler:      router,
		ReadTimeout:  cfg.Services.ControlFlowAPI.ReadTimeout,
		WriteTimeout: cfg.Services.ControlFlowAPI.WriteTimeout,
		IdleTimeout:  cfg.Services.ControlFlowAPI.IdleTimeout,
	}

	// Start server
	go func() {
		fmt.Printf("Control Flow API Server running on http://%s\n", cfg.GetServiceAddr("control"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down Control Flow API Server...")

	// Gracefully shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Control Flow API Server gracefully stopped")
	}
}
