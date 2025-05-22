package main

import (
	"context"
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"os/signal"
	"pvz-cli/internal/app"
	"pvz-cli/internal/config"
	"pvz-cli/pkg/closer"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	var cfgPath string
	pflag.StringVar(&cfgPath, "config", "", "path to config.yaml (or set $PVZ_CONFIG)")
	pflag.Parse()

	if cfgPath == "" {
		cfgPath = os.Getenv("PVZ_CONFIG")
	}

	if cfgPath == "" {
		defaultCfg := "configs/default_config.yaml"
		fmt.Fprintf(os.Stdout,
			"No config specified; using default: %s\n"+
				"You can override with --config or $PVZ_CONFIG\n\n",
			defaultCfg,
		)
		cfgPath = defaultCfg
	} else {
		fmt.Fprintf(os.Stdout, "Using config: %s\n\n", cfgPath)
	}

	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	c := closer.NewCloser()

	log, err := app.SetupLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger init error: %v\n", err)
		os.Exit(1)
	}

	c.Add(func(_ context.Context) error {
		log.Sync()
		return nil
	})

	a := app.New(log)

	runErr := make(chan error, 1)
	go func() {
		runErr <- a.Run(ctx)
	}()

	select {
	case <-ctx.Done():
	case err := <-runErr:
		if err != nil {
			log.Fatalw("Failed to start server",
				"error", err,
			)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		time.Duration(cfg.App.ShutdownTimeout)*time.Second,
	)
	defer cancel()

	if err := c.Close(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "Shutdown error: %v\n", err)
		os.Exit(1)
	}

	// делаю через fmt, потому что логгер уже закрыт
	fmt.Println("Application stopped gracefully")
}
