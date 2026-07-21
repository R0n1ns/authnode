package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"authmicro/configs"
	"authmicro/internal/api/rest/router"
	"authmicro/internal/repository/postgres"
	"authmicro/internal/service"
	"authmicro/pkg/logger"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	log := logger.NewLogger()
	cfg, err := configs.Load()
	if err != nil {
		log.Fatalf("invalid configuration: %v", err)
	}
	db, err := postgres.NewPostgresDB(cfg.DB)
	if err != nil {
		log.Fatalf("database startup failed: %v", err)
	}
	defer db.Close()

	users := postgres.NewUserRepository(db)
	sessions := postgres.NewSessionRepository(db)
	roles := postgres.NewRoleRepository(db)
	tokens := service.NewTokenService(cfg.JWT, sessions)
	emails := service.NewEmailService(cfg.SMTP)
	auth := service.NewAuthService(users, roles, sessions, tokens, emails, log, cfg.OTP)
	server := router.NewRouter(auth, tokens, log, db)

	errorsCh := make(chan error, 1)
	go func() {
		log.Infof("REST API listening on %s", cfg.HTTPServerAddress)
		if err := server.Start(cfg.HTTPServerAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errorsCh <- err
		}
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-signals:
		log.Infof("received signal %s", sig)
	case err := <-errorsCh:
		log.Errorf("HTTP server failed: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Errorf("graceful shutdown failed: %v", err)
	}
}
