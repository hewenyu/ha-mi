package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/boringsoft/ha-mi/internal/auth"
	"github.com/boringsoft/ha-mi/internal/config"
	"github.com/boringsoft/ha-mi/internal/controllers"
	"github.com/boringsoft/ha-mi/internal/db"
)

// Server represents the API server
type Server struct {
	config          *config.Config
	router          *gin.Engine
	httpServer      *http.Server
	jwtService      *auth.JWTService
	nonceService    *auth.NonceService
	securityService *auth.SecurityService
	database        *db.DB
}

// NewServer creates a new API server
func NewServer(cfg *config.Config, database *db.DB) *Server {
	// Create services
	jwtService := auth.NewJWTService(
		cfg.Auth.SecretKey,
		cfg.Auth.AccessTokenExpiry,
		cfg.Auth.RefreshTokenExpiry,
	)
	nonceService := auth.NewNonceService(database.DB, cfg.Auth.NonceExpiry)
	securityService := auth.NewSecurityService(cfg.Auth.SecretKey, 60) // 60 seconds max diff

	// Create server
	server := &Server{
		config:          cfg,
		jwtService:      jwtService,
		nonceService:    nonceService,
		securityService: securityService,
		database:        database,
	}

	// Initialize router
	server.setupRouter()

	return server
}

// setupRouter sets up the HTTP router
func (s *Server) setupRouter() {
	// Create router
	router := gin.Default()

	// Add middleware
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(corsMiddleware())

	// API routes
	apiGroup := router.Group("/api/v1")

	// Add security middleware (except for certain routes)
	apiGroup.Use(SecurityMiddleware(s.nonceService, s.securityService))

	// Create controllers
	authController := controllers.NewAuthController(s.jwtService, s.nonceService, s.securityService, s.config)

	// Register auth routes (no auth middleware needed)
	authController.RegisterRoutes(apiGroup)

	// Protected routes (with auth middleware)
	protectedGroup := apiGroup.Group("")
	protectedGroup.Use(AuthMiddleware(s.jwtService))

	// Add health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	s.router = router
}

// Start starts the API server
func (s *Server) Start() error {
	// Create HTTP server
	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.Server.Host, s.config.Server.Port),
		Handler: s.router,
	}

	// Start server in a goroutine
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("HTTP server error: %s\n", err)
			os.Exit(1)
		}
	}()

	fmt.Printf("Server started on %s:%d\n", s.config.Server.Host, s.config.Server.Port)

	return nil
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(timeout time.Duration) error {
	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Shutdown HTTP server
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("error shutting down HTTP server: %w", err)
	}

	// Close database connection
	if err := s.database.Close(); err != nil {
		return fmt.Errorf("error closing database connection: %w", err)
	}

	return nil
}

// WaitForShutdown waits for shutdown signal
func (s *Server) WaitForShutdown() {
	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	fmt.Println("Shutting down server...")

	// Shutdown with 5s timeout
	if err := s.Shutdown(5 * time.Second); err != nil {
		fmt.Printf("Error during shutdown: %s\n", err)
	}

	fmt.Println("Server gracefully stopped")
}

// corsMiddleware handles CORS
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Timestamp, X-Nonce, X-Sign")
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
