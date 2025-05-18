package cli

import (
	"bufio"
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"pvz-cli/pkg/logger"
	"strings"
)

func BuildRootCommand(log logger.Logger) *cobra.Command {
	root := &cobra.Command{
		Use:   "pvz",
		Short: "Интерактивный Пункт Выдачи Заказов",
	}

	root.AddCommand(
	// TODO вызовы функций для создания команд из cli/commands
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
func (r *REPL) Run(ctx context.Context) error {
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")

		select {
		case <-ctx.Done():
			fmt.Println("\nShutting down App …")
			return nil
		default:
		}

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
		if err := r.root.ExecuteContext(ctx); err != nil {
			fmt.Fprintln(os.Stderr, "ERROR:", err)
		}
	}
}
