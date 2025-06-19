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

	"github.com/spf13/pflag"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	var (
		cfgPath string
		envPath string
	)
	pflag.StringVar(&cfgPath, "config", "", "path to config.yaml (or set $PVZ_CONFIG)")
	pflag.StringVar(&envPath, "env", "", "path to .env file (or set $PVZ_ENV)")
	pflag.Parse()

	// config path
	if cfgPath == "" {
		cfgPath = os.Getenv("PVZ_CONFIG")
	}
	if cfgPath == "" {
		cfgPath = "configs/default_config.yaml"
		fmt.Fprintf(os.Stdout,
			"No config specified; using default: %s\n"+
				"Override with --config or $PVZ_CONFIG\n\n",
			cfgPath,
		)
	} else {
		fmt.Fprintf(os.Stdout, "Using config file: %s\n\n", cfgPath)
	}

	// env path (optional)
	if envPath == "" {
		envPath = os.Getenv("PVZ_ENV")
	}
	if envPath != "" {
		fmt.Fprintf(os.Stdout, "Loading environment from: %s\n\n", envPath)
	} else {
		fmt.Fprintln(os.Stdout, "No .env file specified; skipping env load\n")
	}

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
