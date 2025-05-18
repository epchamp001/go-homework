package main

import (
	"pvz-cli/internal/app"
	"pvz-cli/internal/config"
)

func main() {
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

	if err := a.Run(); err != nil {
		log.Fatalw("Failed to start server",
			"error", err,
		)
	}
}
