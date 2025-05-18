package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/usecase"
)

func NewReturnOrderCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "return-order",
		Short: "Вернуть заказ курьеру",
		RunE: func(cmd *cobra.Command, _ []string) error {
			id, _ := cmd.Flags().GetString("order-id")

			if err := svc.ReturnOrder(id); err != nil {
				printErr(err)
				return nil
			}

			fmt.Println(codes.CodeOrderReturned+":", id)
			return nil
		},
	}

	cmd.Flags().String("order-id", "", "ID заказа (обязательно)")
	_ = cmd.MarkFlagRequired("order-id")
	return cmd
}
