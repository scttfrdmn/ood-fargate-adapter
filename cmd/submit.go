package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/scttfrdmn/ood-fargate-adapter/internal/fargate"
	"github.com/spf13/cobra"
)

// JobSpec is the Fargate-specific job submission payload.
type JobSpec struct {
	Cluster        string            `json:"cluster,omitempty"`
	TaskDefinition string            `json:"task_definition,omitempty"`
	Script         string            `json:"script,omitempty"`
	CPU            string            `json:"cpu,omitempty"`
	Memory         string            `json:"memory,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
	Subnets        []string          `json:"subnets,omitempty"`
	JobName        string            `json:"job_name,omitempty"`
}

var submitCmd = &cobra.Command{
	Use:   "submit",
	Short: "Submit an OOD job to Amazon ECS Fargate",
	Long:  "Reads a JSON job spec from stdin and runs a Fargate task.",
	RunE: func(cmd *cobra.Command, args []string) error {
		var spec JobSpec
		if err := json.NewDecoder(os.Stdin).Decode(&spec); err != nil {
			return fmt.Errorf("decode job spec: %w", err)
		}

		effectiveCluster := cluster
		if effectiveCluster == "" {
			effectiveCluster = spec.Cluster
		}
		if effectiveCluster == "" {
			return fmt.Errorf("--cluster is required (or set cluster in job spec)")
		}

		effectiveTD := taskDefinition
		if effectiveTD == "" {
			effectiveTD = spec.TaskDefinition
		}
		if effectiveTD == "" {
			return fmt.Errorf("--task-definition is required (or set task_definition in job spec)")
		}

		effectiveSubnets := subnets
		if len(effectiveSubnets) == 0 {
			effectiveSubnets = spec.Subnets
		}
		if len(effectiveSubnets) == 0 {
			return fmt.Errorf("--subnets is required (or set subnets in job spec)")
		}

		ctx := context.Background()
		client, err := fargate.New(ctx, region)
		if err != nil {
			return err
		}

		taskArn, err := client.RunTask(ctx, fargate.TaskSpec{
			Cluster:        effectiveCluster,
			TaskDefinition: effectiveTD,
			Script:         spec.Script,
			CPU:            spec.CPU,
			Memory:         spec.Memory,
			Env:            spec.Env,
			Subnets:        effectiveSubnets,
			SecurityGroups: securityGroups,
			JobName:        spec.JobName,
		})
		if err != nil {
			return err
		}

		fmt.Println(taskArn)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(submitCmd)
}
