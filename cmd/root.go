package cmd

import (
	"github.com/spf13/cobra"
)

var (
	region         string
	cluster        string
	taskDefinition string
	subnets        []string
	securityGroups []string
)

var rootCmd = &cobra.Command{
	Use:   "ood-fargate-adapter",
	Short: "OOD compute adapter for Amazon ECS/Fargate",
	Long:  "Translates Open OnDemand job submissions to Amazon ECS Fargate API calls.",
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&region, "region", "us-east-1", "AWS region")
	rootCmd.PersistentFlags().StringVar(&cluster, "cluster", "", "ECS cluster ARN or name")
	rootCmd.PersistentFlags().StringVar(&taskDefinition, "task-definition", "", "ECS task definition family:revision")
	rootCmd.PersistentFlags().StringSliceVar(&subnets, "subnets", nil, "Subnet IDs for the Fargate task (comma-separated)")
	rootCmd.PersistentFlags().StringSliceVar(&securityGroups, "security-groups", nil, "Security group IDs (comma-separated)")
}
