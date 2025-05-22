package commands

import (
	"bufio"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"pvz-cli/internal/domain/codes"
	"pvz-cli/internal/domain/vo"
	"pvz-cli/internal/usecase"
	"strings"
)

func NewScrollOrdersCmd(svc usecase.Service) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scroll-orders",
		Short: "Бесконечный просмотр заказов",
		RunE: func(cmd *cobra.Command, _ []string) error {
			uid, _ := cmd.Flags().GetString("user-id")
			limit, _ := cmd.Flags().GetInt("limit")

			cursor := vo.ScrollCursor{LastID: "", Limit: limit}
			scanner := bufio.NewScanner(os.Stdin)

			for {
				orders, nextCur, err := svc.ScrollOrders(uid, cursor)
				if err != nil {
					printErr(err)
					return nil
				}

				for _, o := range orders {
					fmt.Printf(
						"%s: %s %s %s %s\n",
						codes.CodeOrder,
						o.ID, o.UserID, o.Status,
						o.ExpiresAt.Format("2006-01-02"),
					)
				}

				if nextCur.LastID == "" {
					return nil
				}

				fmt.Println(codes.CodeNext+":", nextCur.LastID)
				fmt.Print("next / exit > ")

				if !scanner.Scan() {
					return nil
				}
				if strings.TrimSpace(scanner.Text()) == "exit" {
					return nil
				}
				cursor = nextCur
			}
		},
	}

	cmd.Flags().String("user-id", "", "ID клиента (обязательно)")
	cmd.Flags().Int("limit", 20, "Размер пачки")
	_ = cmd.MarkFlagRequired("user-id")

	return cmd
}
