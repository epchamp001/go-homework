package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"pvz-cli/internal/usecase"
)

func NewClientReportCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "client-report",
		Short: "Отчёт по клиентам (кол-во заказов, возвратов, сумма выкупа). Сортируются по количеству заказов или сумме выкупа.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sortBy, _ := cmd.Flags().GetString("sort")
			outputPath, _ := cmd.Flags().GetString("output")

			data, err := svc.GenerateClientReportByte(sortBy)
			if err != nil {
				printErr(err)
				return nil
			}

			dir := filepath.Dir(outputPath)
			if dir != "." {
				if err := os.MkdirAll(dir, 0755); err != nil {
					printErr(err)
					return nil
				}
			}

			if err := os.WriteFile(outputPath, data, 0644); err != nil {
				printErr(err)
				return nil
			}

			fmt.Println("Отчёт успешно создан:", outputPath)
			return nil
		},
	}

	cmd.Flags().String("sort", "", "Сортировка: orders или sum (обязательно)")
	cmd.Flags().String("output", "clients_report.xlsx", "Путь к выходному файлу XLSX")
	_ = cmd.MarkFlagRequired("sort")

	return cmd
}
