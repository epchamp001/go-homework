// Package app инициализирует и запускает основное приложение,
// включая конфигурацию, зависимости, маршруты и graceful shutdown.
//
// Этот пакет связывает все модули проекта и является точкой входа при запуске бинарника.
package app

import (
	"context"
	"pvz-cli/internal/handler/cli"
	"pvz-cli/internal/repository/storage/filerepo"
	"pvz-cli/internal/usecase"
	"pvz-cli/pkg/logger"
)

// App позволяет удобно и аккуратно поднимать весь проект и его зависимости.
type App struct {
	repl *cli.REPL
	log  logger.Logger
}

// New создаёт новое приложение с инициализацией хранилища, бизнес-логики и REPL.
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

// Run запускает REPL-приложение, обрабатывающее пользовательские команды.
func (a *App) Run(ctx context.Context) error {
	return a.repl.Run()
}
