package app

import (
	"context"
	"pvz-cli/internal/handler/cli"
	"pvz-cli/pkg/closer"
	"pvz-cli/pkg/logger"
)

// App объединяет зависимости, чтобы наружу остались два метода: Run и Close.
type App struct {
	repl   *cli.REPL
	closer *closer.Closer
	log    logger.Logger
}

func New(log logger.Logger) *App {
	c := closer.NewCloser()

	repo := // TODO создаем здесь репу
		c.Add(repo.Close)

	svc := usecase.NewService(repo, log) // TODO сделать usecase

	rootCmd := cli.BuildRootCommand(svc, log) /* TODO тут скорее всего буду передать svc
	потому что он мне будет нужен для создания команд
	а команды добавляю в BuildRootCommand
	*/
	repl := cli.NewREPL(rootCmd)

	return &App{
		repl:   repl,
		closer: c,
		log:    log,
	}
}

func (a *App) Run(ctx context.Context) error {
	return a.repl.Run(ctx)
}

func (a *App) Shutdown(ctx context.Context) error {
	return a.closer.Close(ctx)
}
