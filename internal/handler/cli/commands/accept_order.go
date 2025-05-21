package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/usecase"
	"strings"
	"time"
)

func NewAcceptOrderCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "accept-order",
		Short:         "Принять заказ от курьера",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			orderID, _ := cmd.Flags().GetString("order-id")
			userID, _ := cmd.Flags().GetString("user-id")
			expStr, _ := cmd.Flags().GetString("expires")

			// валидация на корректные данные
			if orderID == "" || strings.HasPrefix(orderID, "-") {
				printErr(codes.ErrValidationFailed)
				return nil
			}
			if userID == "" || strings.HasPrefix(userID, "-") {
				printErr(codes.ErrValidationFailed)
				return nil
			}

			exp, err := time.Parse("2006-01-02", expStr)
			if err != nil {
				printErr(codes.ErrValidationFailed)
				return nil
			}

			if err := svc.AcceptOrder(orderID, userID, exp); err != nil {
				printErr(err)
				return nil
			}

			fmt.Println(codes.CodeOrderAccepted+":", orderID)
			return nil
		},
	}

	cmd.Flags().String("order-id", "", "ID заказа (обязательно)")
	cmd.Flags().String("user-id", "", "ID клиента (обязательно)")
	cmd.Flags().String("expires", "", "Срок хранения YYYY-MM-DD")

	return cmd
}
