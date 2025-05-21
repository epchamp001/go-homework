package app

import (
	"context"
	"pvz-cli/internal/handler/cli"
	"pvz-cli/internal/repository/storage/filerepo"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/logger"
)

type App struct {
	repl *cli.REPL
	log  logger.Logger
}

func New(log logger.Logger) *App {

	repo, err := filerepo.NewFileRepo("data")
	if err != nil {
		log.Fatalw("could not create repo",
			"error", err,
		)
	}

	svc := usecase.NewService(repo)

	repl := cli.NewREPL(svc)

	return &App{
		repl: repl,
		log:  log,
	}
}

func (a *App) Run(ctx context.Context) error {
	return a.repl.Run()
}
