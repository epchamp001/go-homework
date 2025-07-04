package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"pvz-cli/internal/app"
	"pvz-cli/internal/config"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	cfgPath, envPath := setupFlags()

	cfg, err := config.LoadConfig(cfgPath, envPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	log, err := app.SetupLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger init error: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	server := app.NewServer(cfg, log)

	if err := server.Run(ctx); err != nil {
		log.Fatalw("Failed to start server", "error", err)
	}

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.GRPCServer.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Errorw("Shutdown failed", "error", err)
		os.Exit(1)
	}

	log.Info("Application stopped gracefully")
}
