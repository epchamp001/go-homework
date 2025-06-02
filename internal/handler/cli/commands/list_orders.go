package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/internal/usecase"
)

// NewListOrdersCmd возвращает CLI-команду `list-orders`, которая выводит список заказов клиента с возможностью пагинации.
func NewListOrdersCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-orders",
		Short: "Список заказов клиента",
		RunE: func(cmd *cobra.Command, _ []string) error {
			uid, _ := cmd.Flags().GetString("user-id")
			inPVZ, _ := cmd.Flags().GetBool("in-pvz")
			last, _ := cmd.Flags().GetInt("last")
			page, _ := cmd.Flags().GetInt("page")
			limit, _ := cmd.Flags().GetInt("limit")

			orders, total, err := svc.ListOrders(
				uid,
				inPVZ,
				last,
				vo.Pagination{Page: page, Limit: limit},
			)
			if err != nil {
				printErr(err)
				return nil
			}

			for _, o := range orders {
				fmt.Printf(
					"%s: %s %s %s %s %s %.2f %.2f₽\n",
					codes.CodeOrder,
					o.ID,
					o.UserID,
					o.Status,
					o.ExpiresAt.Format("2006-01-02"),
					o.Package,
					o.Weight,
					float64(o.Price)/100.0,
				)
			}

			fmt.Println(codes.CodeTotal+":", total)
			return nil
		},
	}

	cmd.Flags().String("user-id", "", "ID клиента")
	cmd.Flags().Bool("in-pvz", false, "Только находящиеся в ПВЗ")
	cmd.Flags().Int("last", 0, "Последние N")
	cmd.Flags().Int("page", 0, "Номер страницы")
	cmd.Flags().Int("limit", 0, "Размер страницы")
	_ = cmd.MarkFlagRequired("user-id")

	return cmd
}
