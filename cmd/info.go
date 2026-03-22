package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/scttfrdmn/ood-fargate-adapter/internal/fargate"
	"github.com/spf13/cobra"
)

var infoCmd = &cobra.Command{
	Use:   "info <task-arn>",
	Short: "Print full Fargate task details as JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		client, err := fargate.New(ctx, region)
		if err != nil {
			return err
		}
		task, err := client.DescribeTask(ctx, cluster, args[0])
		if err != nil {
			return err
		}
		return json.NewEncoder(os.Stdout).Encode(task)
	},
}

func init() {
	rootCmd.AddCommand(infoCmd)
}
