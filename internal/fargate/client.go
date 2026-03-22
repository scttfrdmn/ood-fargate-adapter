// Package fargate wraps the Amazon ECS API for the OOD adapter.
package fargate

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

// Client wraps the Amazon ECS client.
type Client struct {
	svc    *ecs.Client
	region string
}

// New creates an ECS client using the default AWS credential chain.
func New(ctx context.Context, region string) (*Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))
	if err != nil {
		return nil, fmt.Errorf("load AWS config: %w", err)
	}
	return &Client{svc: ecs.NewFromConfig(cfg), region: region}, nil
}

// TaskSpec holds the parameters for a Fargate task.
type TaskSpec struct {
	Cluster        string
	TaskDefinition string
	Script         string
	CPU            string
	Memory         string
	Env            map[string]string
	Subnets        []string
	SecurityGroups []string
	JobName        string
}

// RunTask launches a Fargate task and returns the task ARN.
func (c *Client) RunTask(ctx context.Context, spec TaskSpec) (string, error) {
	var envVars []types.KeyValuePair
	for k, v := range spec.Env {
		envVars = append(envVars, types.KeyValuePair{
			Name:  aws.String(k),
			Value: aws.String(v),
		})
	}

	input := &ecs.RunTaskInput{
		Cluster:        aws.String(spec.Cluster),
		TaskDefinition: aws.String(spec.TaskDefinition),
		LaunchType:     types.LaunchTypeFargate,
		NetworkConfiguration: &types.NetworkConfiguration{
			AwsvpcConfiguration: &types.AwsVpcConfiguration{
				Subnets:        spec.Subnets,
				SecurityGroups: spec.SecurityGroups,
				AssignPublicIp: types.AssignPublicIpEnabled,
			},
		},
		Overrides: &types.TaskOverride{
			ContainerOverrides: []types.ContainerOverride{
				{
					Name:        aws.String("app"),
					Environment: envVars,
				},
			},
		},
	}
	if spec.CPU != "" {
		input.Overrides.Cpu = aws.String(spec.CPU)
	}
	if spec.Memory != "" {
		input.Overrides.Memory = aws.String(spec.Memory)
	}
	if spec.Script != "" {
		input.Overrides.ContainerOverrides[0].Command = []string{"/bin/sh", "-c", spec.Script}
	}

	out, err := c.svc.RunTask(ctx, input)
	if err != nil {
		return "", fmt.Errorf("ecs RunTask: %w", err)
	}
	if len(out.Failures) > 0 {
		return "", fmt.Errorf("ecs RunTask failure: %s: %s", aws.ToString(out.Failures[0].Arn), aws.ToString(out.Failures[0].Reason))
	}
	if len(out.Tasks) == 0 {
		return "", fmt.Errorf("ecs RunTask returned no tasks")
	}
	return aws.ToString(out.Tasks[0].TaskArn), nil
}

// DescribeTask returns the current detail of an ECS task.
func (c *Client) DescribeTask(ctx context.Context, cluster, taskArn string) (*types.Task, error) {
	out, err := c.svc.DescribeTasks(ctx, &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   []string{taskArn},
	})
	if err != nil {
		return nil, fmt.Errorf("ecs DescribeTasks: %w", err)
	}
	if len(out.Tasks) == 0 {
		return nil, fmt.Errorf("task %q not found", taskArn)
	}
	return &out.Tasks[0], nil
}

// StopTask stops an ECS task.
func (c *Client) StopTask(ctx context.Context, cluster, taskArn, reason string) error {
	_, err := c.svc.StopTask(ctx, &ecs.StopTaskInput{
		Cluster: aws.String(cluster),
		Task:    aws.String(taskArn),
		Reason:  aws.String(reason),
	})
	if err != nil {
		return fmt.Errorf("ecs StopTask: %w", err)
	}
	return nil
}
