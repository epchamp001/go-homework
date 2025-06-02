package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/usecase"
)

// NewImportOrdersCmd возвращает CLI-команду `import-orders`, которая импортирует заказы из внешнего файла.
func NewImportOrdersCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import-orders",
		Short: "Импорт заказов из JSON-файла",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path, _ := cmd.Flags().GetString("file")

			n, err := svc.ImportOrders(path)
			if err != nil {
				printErr(err)
				return nil
			}

			fmt.Println(codes.CodeImported+":", n)
			return nil
		},
	}

	cmd.Flags().String("file", "", "Путь к JSON")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}
