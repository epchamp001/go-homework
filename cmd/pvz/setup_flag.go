package main

import (
	"fmt"
	"os"

	"github.com/spf13/pflag"
)

func setupFlags() (string, string) {
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
		fmt.Fprintf(os.Stdout, "No .env file specified; skipping env load\n")
	}

	return cfgPath, envPath
}
