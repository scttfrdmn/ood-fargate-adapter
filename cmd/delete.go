package cmd

import (
	"context"
	"fmt"

	"github.com/scttfrdmn/ood-fargate-adapter/internal/fargate"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <task-arn>",
	Short: "Stop a Fargate task",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := fargate.New(ctx, region)
		if err != nil {
			return err
		}
		if err := client.StopTask(ctx, cluster, args[0], "Cancelled via OOD"); err != nil {
			return err
		}
		fmt.Printf("Task %s stopped\n", args[0])
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
}
