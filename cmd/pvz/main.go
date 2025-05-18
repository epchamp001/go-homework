package main

import (
	"context"
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

	cfg, err := config.LoadConfig("configs/")
	if err != nil {
		panic(err)
	}

	log, err := app.SetupLogger(cfg.Logging)
	if err != nil {
		panic(err)
	}
	defer log.Sync()

	a := app.New(log)

	if err := a.Run(ctx); err != nil {
		log.Fatalw("Failed to start server",
			"error", err,
		)
		os.Exit(1)
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.App.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := a.Shutdown(shutdownCtx); err != nil {
		log.Errorw("Shutdown failed",
			"error", err,
		)
		os.Exit(1)
	}

	log.Info("Application stopped gracefully")
}
