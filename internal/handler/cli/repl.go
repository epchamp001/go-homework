// Package cli предоставляет пользовательский интерфейс командной строки (CLI) для работы с приложением.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"pvz-cli/internal/handler/cli/commands"
	"pvz-cli/internal/usecase"
	"strings"

	"github.com/spf13/cobra"
)

// REPL реализует простой цикл командной строки (Read-Eval-Print Loop) для взаимодействия с пользователем через текстовый интерфейс.
type REPL struct {
	svc usecase.Service
}

// NewREPL создает новый экземпляр REPL с сервисом бизнес-логики.
func NewREPL(svc usecase.Service) *REPL {
	return &REPL{svc: svc}
}

// Run запускает бесконечный цикл. Завершается при "exit"/EOF или ctx.Done()
func (r *REPL) Run() error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		if !scanner.Scan() {
			return nil
		}
		line := strings.TrimSpace(scanner.Text())
		switch line {
		case "":
			continue
		case "exit":
			return nil
		}

		root := buildRootCommand(r.svc)

		// парсим args и выполняем
		args := strings.Fields(line)
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
		}
	}
}

func buildRootCommand(svc usecase.Service) *cobra.Command {
	root := &cobra.Command{
		Use:   "pvz",
		Short: "Интерактивный Пункт Выдачи Заказов",
	}

	root.CompletionOptions.DisableDefaultCmd = true

	root.AddCommand(
		commands.NewAcceptOrderCmd(svc),
		commands.NewReturnOrderCmd(svc),
		commands.NewProcessOrdersCmd(svc),
		commands.NewListOrdersCmd(svc),
		commands.NewListReturnsCmd(svc),
		commands.NewOrderHistoryCmd(svc),
		commands.NewScrollOrdersCmd(svc),
		commands.NewImportOrdersCmd(svc),
		commands.NewClientReportCmd(svc),
	)

	return root
}
