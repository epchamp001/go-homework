package cli

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"pvz-cli/internal/handler/cli/commands"
	"pvz-cli/internal/usecase"
	"strings"
)

func BuildRootCommand(svc usecase.Service) *cobra.Command {
	root := &cobra.Command{
		Use:   "pvz",
		Short: "Интерактивный Пункт Выдачи Заказов",
	}

	root.AddCommand(
		commands.NewAcceptOrderCmd(svc),
		commands.NewReturnOrderCmd(svc),
		commands.NewProcessOrdersCmd(svc),
		commands.NewListOrdersCmd(svc),
		commands.NewListReturnsCmd(svc),
		commands.NewOrderHistoryCmd(svc),
		commands.NewScrollOrdersCmd(svc),
		commands.NewImportOrdersCmd(svc),
	)

	return root
}

// REPL - Read, Eval, Print, Loop - прочитать, вычислить, вывести, цикл
type REPL struct {
	root *cobra.Command
}

func NewREPL(root *cobra.Command) *REPL {
	return &REPL{root: root}
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

		// проксируем строку в cobra
		r.root.SetArgs(strings.Fields(line))
		if err := r.root.Execute(); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
		}
	}
}
