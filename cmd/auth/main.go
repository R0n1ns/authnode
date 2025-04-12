package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"authmicro/configs"
	"authmicro/internal/api/grpc/server"
	"authmicro/internal/api/rest/router"
	"authmicro/internal/repository/postgres"
	"authmicro/internal/service"
	"authmicro/pkg/logger"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Initialize logger
	l := logger.NewLogger()
	l.Info("Starting auth microservice")

	// Load configuration
	cfg := configs.NewConfig()
	l.Info("Configuration loaded successfully")

	// Initialize database connection
	db, err := postgres.NewPostgresDB(cfg.DB)
	if err != nil {
		l.Fatalf("Failed to initialize database connection: %v", err)
	}
	defer db.Close()
	l.Info("Database connection established")

	// Initialize database schema
	if err := postgres.InitSchema(db); err != nil {
		l.Fatalf("Failed to initialize database schema: %v", err)
	}

	// Initialize repositories
	userRepo := postgres.NewUserRepository(db)
	sessionRepo := postgres.NewSessionRepository(db)
	roleRepo := postgres.NewRoleRepository(db)

	// Initialize services
	tokenService := service.NewTokenService(cfg.JWT, sessionRepo)
	emailService := service.NewEmailService(cfg.SMTP)
	authService := service.NewAuthService(userRepo, roleRepo, sessionRepo, tokenService, emailService, l)

	// Initialize REST router
	r := router.NewRouter(authService, tokenService, l)

	// Start REST server
	go func() {
		l.Infof("Starting REST server on %s", cfg.HTTPServerAddress)
		if err := r.Start(cfg.HTTPServerAddress); err != nil {
			l.Fatalf("Failed to start REST server: %v", err)
		}
	}()

	// Initialize and start gRPC server
	grpcServer := server.NewGRPCServer(cfg.GRPCServerAddress, authService, tokenService, l)
	go func() {
		l.Infof("Starting gRPC server on %s", cfg.GRPCServerAddress)
		if err := grpcServer.Start(); err != nil {
			l.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	l.Info("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	grpcServer.Stop()

	if err := r.Shutdown(ctx); err != nil {
		l.Errorf("Error during server shutdown: %v", err)
	}

	l.Info("Server exited properly")
}
