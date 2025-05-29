// Package commands содержит определения CLI-команд для приложения.
package commands

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/models"
	"pvz-cli/internal/usecase"
)

// NewAcceptOrderCmd возвращает CLI-команду `accept-order`, которая позволяет принять заказ от курьера с указанием упаковки.
func NewAcceptOrderCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "accept-order",
		Short:         "Принять заказ от курьера с выбором упаковки",
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, _ []string) error {
			orderID, _ := cmd.Flags().GetString("order-id")
			userID, _ := cmd.Flags().GetString("user-id")
			expStr, _ := cmd.Flags().GetString("expires")
			weight, _ := cmd.Flags().GetFloat64("weight")
			priceRub, _ := cmd.Flags().GetFloat64("price")
			packageStr, _ := cmd.Flags().GetString("package")

			if orderID == "" || strings.HasPrefix(orderID, "-") ||
				userID == "" || strings.HasPrefix(userID, "-") ||
				weight <= 0 || priceRub <= 0 {
				printErr(codes.ErrValidationFailed)
				return nil
			}

			exp, err := time.Parse("2006-01-02", expStr)
			if err != nil {
				printErr(codes.ErrValidationFailed)
				return nil
			}

			pkgType := models.PackageType(packageStr)
			priceKopecks := models.PriceKopecks(priceRub * 100)

			total, err := svc.AcceptOrder(orderID, userID, exp, weight, priceKopecks, pkgType)
			if err != nil {
				printErr(err)
				return nil
			}

			fmt.Println(codes.CodeOrderAccepted+":", orderID)
			fmt.Println("PACKAGE:", packageStr)
			fmt.Printf("TOTAL_PRICE: %.2f₽\n", float64(total)/100.0)
			return nil
		},
	}

	cmd.Flags().String("order-id", "", "ID заказа (обязательно)")
	cmd.Flags().String("user-id", "", "ID клиента (обязательно)")
	cmd.Flags().String("expires", "", "Срок хранения YYYY-MM-DD (обязательно)")
	cmd.Flags().Float64("weight", 0, "Вес заказа в кг (обязательно > 0)")
	cmd.Flags().Float64("price", 0, "Цена заказа в рублях (обязательно > 0)")
	cmd.Flags().String("package", "", "Тип упаковки: bag | box | film | bag+film | box+film (опционально)")

	return cmd
}
