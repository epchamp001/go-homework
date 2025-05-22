package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/usecase"
	"strings"
)

func NewProcessOrdersCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "process-orders",
		Short: "Выдать заказы или принять возврат клиента",
		RunE: func(cmd *cobra.Command, _ []string) error {
			uid, _ := cmd.Flags().GetString("user-id")
			action, _ := cmd.Flags().GetString("action")
			raw, _ := cmd.Flags().GetString("order-ids")

			ids := strings.Split(strings.TrimSpace(raw), ",")
			if len(ids) == 0 || ids[0] == "" {
				printErr(codes.ErrValidationFailed)
				return nil
			}

			var (
				res map[string]error
				err error
			)

			switch action {
			case "issue":
				res, err = svc.IssueOrders(uid, ids)
			case "return":
				res, err = svc.ReturnOrdersByClient(uid, ids)
			default:
				printErr(codes.ErrValidationFailed)
				return nil
			}

			if err != nil {
				printErr(err)
				return nil
			}

			for id, e := range res {
				if e != nil {
					fmt.Fprintf(os.Stderr, "ERROR %s: %v\n", id, e)
				} else {
					fmt.Println(codes.CodeProcessed+":", id)
				}
			}
			return nil
		},
	}

	cmd.Flags().String("user-id", "", "ID клиента (обязательно)")
	cmd.Flags().String("action", "", "issue|return")
	cmd.Flags().String("order-ids", "", "Список id1,id2,…")
	_ = cmd.MarkFlagRequired("user-id")
	_ = cmd.MarkFlagRequired("action")
	_ = cmd.MarkFlagRequired("order-ids")

	return cmd
}
