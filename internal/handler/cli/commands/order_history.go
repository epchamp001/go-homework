package commands

import (
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/usecase"
)

// NewOrderHistoryCmd возвращает CLI-команду `order-history`, которая показывает историю изменения заказов.
func NewOrderHistoryCmd(svc usecase.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "order-history",
		Short: "История изменений заказов",
		RunE: func(cmd *cobra.Command, _ []string) error {
			hist, err := svc.OrderHistory()
			if err != nil {
				printErr(err)
				return nil
			}

			for _, e := range hist {
				fmt.Printf(
					"%s: %s %s %s\n",
					codes.CodeHistory,
					e.OrderID,
					e.Status,
					e.Time.Format(time.RFC3339),
				)
			}
			return nil
		},
	}
}
