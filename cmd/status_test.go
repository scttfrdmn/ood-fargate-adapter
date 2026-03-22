package cmd

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	ecstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
)

func TestFargateTaskToOod(t *testing.T) {
	tests := []struct {
		name       string
		lastStatus string
		exitCodes  []*int32 // nil = no exit code
		want       string
	}{
		{"provisioning", "PROVISIONING", nil, "queued"},
		{"pending", "PENDING", nil, "queued"},
		{"activating", "ACTIVATING", nil, "queued"},
		{"running", "RUNNING", nil, "running"},
		{"deprovisioning", "DEPROVISIONING", nil, "running"},
		{"stopping", "STOPPING", nil, "running"},
		{"stopped_success", "STOPPED", []*int32{aws.Int32(0)}, "completed"},
		{"stopped_failure", "STOPPED", []*int32{aws.Int32(1)}, "failed"},
		{"stopped_no_containers", "STOPPED", nil, "completed"},
		{"unknown", "WEIRD_STATE", nil, "undetermined"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			task := &ecstypes.Task{
				LastStatus: aws.String(tt.lastStatus),
			}
			for _, ec := range tt.exitCodes {
				task.Containers = append(task.Containers, ecstypes.Container{
					ExitCode: ec,
				})
			}
			got := fargateTaskToOod(task)
			if got != tt.want {
				t.Errorf("fargateTaskToOod(%q) = %q, want %q", tt.lastStatus, got, tt.want)
			}
		})
	}
}
