package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-backend/configs"
	"go-backend/internals/api"
	"go-backend/internals/websocket"
	"go-backend/pkg/database"
	"go-backend/pkg/logger"
)

func main() {
	configPath := "configs/configs.yaml"
	if os.Getenv("CONFIG_PATH") != "" {
		configPath = os.Getenv("CONFIG_PATH")
	}

	cfg, err := configs.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("[FATAL] Failed to load config: %v\n", err)
		os.Exit(1)
	}

	logger.InitLogger(cfg.Server.Mode)
	defer logger.Sync()

	logger.Info("========================================================")
	logger.Info("            ITAM Golang Backend Starting...             ")
	logger.Info("========================================================")

	db, err := database.InitDB(cfg)
	if err != nil {
		logger.Fatalf("Database initialization failed: %v", err)
	}

	hub := websocket.InitHub()

	router := api.SetupRouter(db, cfg, hub)

	port := cfg.Server.Port
	if port <= 0 {
		port = 44318
	}

	addr := fmt.Sprintf(":%d", port)
	srv := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	go func() {
		if cfg.Server.CertFile != "" && cfg.Server.KeyFile != "" {
			logger.Infof("Server is listening on https://localhost%s", addr)
			if err := srv.ListenAndServeTLS(cfg.Server.CertFile, cfg.Server.KeyFile); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("Server TLS listen failed: %v", err)
			}
		} else {
			logger.Infof("Server is listening on http://localhost%s", addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatalf("Server listen failed: %v", err)
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatalf("Server forced to shutdown: %v", err)
	}

	logger.Info("Server exited properly.")
}
