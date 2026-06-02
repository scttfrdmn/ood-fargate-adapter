//go:build integration

package cmd

import (
	"bytes"
	"context"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awsecs "github.com/aws/aws-sdk-go-v2/service/ecs"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	substrate "github.com/scttfrdmn/substrate"
)

// runFargateCmd pipes stdinData into os.Stdin, captures os.Stdout, executes
// rootCmd with args, and returns the trimmed output.
func runFargateCmd(t *testing.T, stdinData string, args ...string) string {
	t.Helper()

	if stdinData != "" {
		stdinR, stdinW, err := os.Pipe()
		if err != nil {
			t.Fatalf("create stdin pipe: %v", err)
		}
		origStdin := os.Stdin
		os.Stdin = stdinR
		defer func() {
			os.Stdin = origStdin
			stdinR.Close()
		}()
		if _, err := io.WriteString(stdinW, stdinData); err != nil {
			t.Fatalf("write stdin pipe: %v", err)
		}
		stdinW.Close()
	}

	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	origStdout := os.Stdout
	os.Stdout = stdoutW
	defer func() { os.Stdout = origStdout }()

	rootCmd.SilenceUsage = true
	rootCmd.SilenceErrors = true
	rootCmd.SetArgs(args)
	execErr := rootCmd.Execute()

	stdoutW.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, readErr := io.Copy(&buf, stdoutR); readErr != nil {
		t.Fatalf("read stdout pipe: %v", readErr)
	}
	stdoutR.Close()

	if execErr != nil {
		t.Fatalf("rootCmd.Execute(%v): %v", args, execErr)
	}
	return strings.TrimSpace(buf.String())
}

func TestCLISubmitStatusDelete_Fargate_Substrate(t *testing.T) {
	ts := substrate.StartTestServer(t)
	t.Setenv("AWS_ENDPOINT_URL", ts.URL)
	t.Setenv("AWS_ACCESS_KEY_ID", "test")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "test")

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion("us-east-1"),
		config.WithBaseEndpoint(ts.URL),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	if err != nil {
		t.Fatalf("load aws config: %v", err)
	}
	raw := awsecs.NewFromConfig(cfg)

	// Create the ECS cluster that the CLI will target.
	clusterOut, err := raw.CreateCluster(ctx, &awsecs.CreateClusterInput{
		ClusterName: aws.String("ood-cli-cluster"),
	})
	if err != nil {
		t.Fatalf("CreateCluster: %v", err)
	}
	clusterName := aws.ToString(clusterOut.Cluster.ClusterName)
	t.Logf("created cluster: %s", clusterName)

	// Register a minimal Fargate task definition.
	_, err = raw.RegisterTaskDefinition(ctx, &awsecs.RegisterTaskDefinitionInput{
		Family:      aws.String("ood-cli-task"),
		NetworkMode: ecstypes.NetworkModeAwsvpc,
		ContainerDefinitions: []ecstypes.ContainerDefinition{
			{
				Name:    aws.String("app"),
				Image:   aws.String("alpine:3.18"),
				Command: []string{"echo", "hello"},
			},
		},
		RequiresCompatibilities: []ecstypes.Compatibility{ecstypes.CompatibilityFargate},
		Cpu:                     aws.String("256"),
		Memory:                  aws.String("512"),
	})
	if err != nil {
		t.Fatalf("RegisterTaskDefinition: %v", err)
	}
	t.Log("registered task definition: ood-cli-task:1")

	// submit — pass --subnets here only; the slice persists for subsequent calls.
	spec := `{"job_name":"cli-fargate-test"}`
	taskArn := runFargateCmd(t, spec,
		"submit",
		"--region", "us-east-1",
		"--cluster", clusterName,
		"--task-definition", "ood-cli-task:1",
		"--subnets", "subnet-test123",
	)
	if taskArn == "" {
		t.Fatal("submit: expected non-empty task ARN")
	}
	if !strings.Contains(taskArn, "arn:aws") {
		t.Errorf("submit: task ARN does not look like an ARN: %s", taskArn)
	}
	t.Logf("submitted task: %s", taskArn)

	// status — omit --subnets; the value persists from the submit call.
	statusOut := runFargateCmd(t, "",
		"status",
		"--region", "us-east-1",
		"--cluster", clusterName,
		taskArn,
	)
	t.Logf("status output: %s", statusOut)
	if !strings.Contains(statusOut, "running") && !strings.Contains(statusOut, "completed") &&
		!strings.Contains(statusOut, "queued") {
		t.Errorf("status output does not contain a recognised status: %s", statusOut)
	}

	// delete — omit --subnets for the same reason.
	deleteOut := runFargateCmd(t, "",
		"delete",
		"--region", "us-east-1",
		"--cluster", clusterName,
		taskArn,
	)
	t.Logf("delete output: %s", deleteOut)
	if !strings.Contains(deleteOut, taskArn) {
		t.Errorf("delete output does not reference task ARN %q: %s", taskArn, deleteOut)
	}
}
