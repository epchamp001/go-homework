package commands

import (
	"fmt"

	"github.com/spf13/cobra"

	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/internal/usecase"
)

// NewListReturnsCmd возвращает CLI-команду `list-returns`, которая выводит список всех возвратов по заказам клиента.
func NewListReturnsCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-returns",
		Short: "Список возвратов",
		RunE: func(cmd *cobra.Command, _ []string) error {
			page, _ := cmd.Flags().GetInt("page")
			limit, _ := cmd.Flags().GetInt("limit")

			recs, err := svc.ListReturns(vo.Pagination{Page: page, Limit: limit})
			if err != nil {
				printErr(err)
				return nil
			}

			for _, r := range recs {
				fmt.Printf(
					"%s: %s %s %s\n",
					codes.CodeReturn,
					r.OrderID,
					r.UserID,
					r.ReturnedAt.Format("2006-01-02"),
				)
			}
			fmt.Printf("%s: %d %s: %d\n", codes.CodePage, page, codes.CodeLimit, limit)
			return nil
		},
	}

	cmd.Flags().Int("page", 1, "Номер страницы")
	cmd.Flags().Int("limit", 20, "Размер страницы")
	return cmd
}
