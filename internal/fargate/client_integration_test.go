//go:build integration

package fargate_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/scttfrdmn/ood-fargate-adapter/internal/fargate"
	substrate "github.com/scttfrdmn/substrate"
)

// registerTestTaskDef registers a minimal task definition in substrate so that
// RunTask can resolve it. Returns the registered task definition family:revision string.
func registerTestTaskDef(t *testing.T, ctx context.Context, endpointURL string) string {
	t.Helper()
	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint(endpointURL),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		t.Fatalf("registerTestTaskDef: load config: %v", err)
	}
	ecsSvc := ecs.NewFromConfig(cfg)
	out, err := ecsSvc.RegisterTaskDefinition(ctx, &ecs.RegisterTaskDefinitionInput{
		Family: aws.String("test-task"),
		ContainerDefinitions: []ecstypes.ContainerDefinition{
			{
				Name:  aws.String("app"),
				Image: aws.String("busybox"),
			},
		},
		NetworkMode:             ecstypes.NetworkModeAwsvpc,
		RequiresCompatibilities: []ecstypes.Compatibility{ecstypes.CompatibilityFargate},
		Cpu:                     aws.String("256"),
		Memory:                  aws.String("512"),
	})
	if err != nil {
		t.Fatalf("registerTestTaskDef: RegisterTaskDefinition: %v", err)
	}
	return fmt.Sprintf("%s:%d", aws.ToString(out.TaskDefinition.Family), out.TaskDefinition.Revision)
}

func TestRunTask_Substrate(t *testing.T) {
	ts := substrate.StartTestServer(t)
	t.Setenv("AWS_ENDPOINT_URL", ts.URL)
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	ctx := context.Background()

	// Register task definition before running the task.
	tdRef := registerTestTaskDef(t, ctx, ts.URL)
	t.Logf("registered task definition: %s", tdRef)

	client, err := fargate.New(ctx, "us-east-1")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	spec := fargate.TaskSpec{
		Cluster:        "test-cluster",
		TaskDefinition: tdRef,
		Script:         "echo hello",
		CPU:            "256",
		Memory:         "512",
		Subnets:        []string{"subnet-12345"},
		JobName:        "test-job",
	}

	taskArn, err := client.RunTask(ctx, spec)
	if err != nil {
		t.Fatalf("RunTask: %v", err)
	}
	if taskArn == "" {
		t.Fatal("expected non-empty task ARN")
	}
	t.Logf("task ARN: %s", taskArn)

	// Describe the task.
	task, err := client.DescribeTask(ctx, spec.Cluster, taskArn)
	if err != nil {
		t.Fatalf("DescribeTask: %v", err)
	}
	if aws.ToString(task.TaskArn) == "" {
		t.Error("expected non-empty TaskArn in response")
	}
	lastStatus := aws.ToString(task.LastStatus)
	t.Logf("task status: %s", lastStatus)
	if lastStatus == "" {
		t.Error("expected non-empty LastStatus")
	}

	// Stop the task.
	err = client.StopTask(ctx, spec.Cluster, taskArn, "test teardown")
	if err != nil {
		t.Fatalf("StopTask: %v", err)
	}
	t.Logf("task stopped")
}

func TestDescribeTask_NotFound_Substrate(t *testing.T) {
	ts := substrate.StartTestServer(t)
	t.Setenv("AWS_ENDPOINT_URL", ts.URL)
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	ctx := context.Background()
	client, err := fargate.New(ctx, "us-east-1")
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	_, err = client.DescribeTask(ctx, "test-cluster", "arn:aws:ecs:us-east-1:123456789012:task/does-not-exist")
	if err == nil {
		t.Fatal("expected error for non-existent task, got nil")
	}
	if !strings.Contains(err.Error(), "not found") && !strings.Contains(err.Error(), "does-not-exist") {
		t.Logf("error (acceptable): %v", err)
	}
}
