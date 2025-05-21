package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"pvz-cli/internal/app"
	"pvz-cli/internal/config"
)

func main() {
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

	log, err := app.SetupLogger(cfg.Logging)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Logger init error: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	a := app.New(log)

	if err := a.Run(); err != nil {
		log.Fatalw("Failed to start server",
			"error", err,
		)
	}
}
