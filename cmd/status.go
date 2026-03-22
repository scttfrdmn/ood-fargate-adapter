package cmd

import (
	"context"
	"encoding/json"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/scttfrdmn/ood-fargate-adapter/internal/fargate"
	internalood "github.com/scttfrdmn/ood-fargate-adapter/internal/ood"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status <task-arn>",
	Short: "Get the status of a Fargate task",
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

		js := internalood.JobStatus{
			ID:     args[0],
			Status: fargateTaskToOod(task),
		}
		if task.StoppedReason != nil {
			js.Message = *task.StoppedReason
		}
		for _, c := range task.Containers {
			if c.ExitCode != nil {
				js.ExitCode = int(aws.ToInt32(c.ExitCode))
				break
			}
		}

		return json.NewEncoder(os.Stdout).Encode(js)
	},
}

func fargateTaskToOod(task *ecstypes.Task) string {
	lastStatus := aws.ToString(task.LastStatus)
	switch lastStatus {
	case "PROVISIONING", "PENDING", "ACTIVATING":
		return internalood.StatusQueued
	case "RUNNING", "DEPROVISIONING", "STOPPING":
		return internalood.StatusRunning
	case "STOPPED":
		// Check exit codes to determine success vs failure
		for _, c := range task.Containers {
			if c.ExitCode != nil && aws.ToInt32(c.ExitCode) != 0 {
				return internalood.StatusFailed
			}
		}
		return internalood.StatusCompleted
	default:
		return internalood.StatusUnknown
	}
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
