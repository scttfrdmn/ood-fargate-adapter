package awscfg

import (
	"context"
	"testing"
)

func TestLoadOptions_NoAssumeRole(t *testing.T) {
	// Empty AssumeRoleARN => exactly the region option, no credentials provider (single-account).
	opts := LoadOptions(context.Background(), Options{Region: "us-west-2"})
	if len(opts) != 1 {
		t.Fatalf("expected 1 option (region only) when no role, got %d", len(opts))
	}
}

func TestRuntimeUser_OODUserOverride(t *testing.T) {
	t.Setenv("OOD_USER", "demo")
	if got := runtimeUser(); got != "demo" {
		t.Errorf("runtimeUser() = %q, want demo (from OOD_USER)", got)
	}
}

func TestLoadOptions_WithAssumeRole(t *testing.T) {
	// A role ARN appends a credentials-provider option on top of the region option.
	opts := LoadOptions(context.Background(), Options{
		Region:        "us-west-2",
		AssumeRoleARN: "arn:aws:iam::123456789012:role/ood-user-demo",
		ExternalID:    "ood",
		SessionName:   "demo",
	})
	if len(opts) != 2 {
		t.Fatalf("expected 2 options (region + creds provider) when role set, got %d", len(opts))
	}
}
