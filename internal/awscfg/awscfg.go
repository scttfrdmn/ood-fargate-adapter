// Package awscfg builds the AWS config option set shared by the adapter commands.
//
// #78 (aws-openondemand): in the multi-account best-practice topology, an adapter's AWS calls
// should run in the LOGGING-IN USER's account under a per-user role — not the OOD instance
// role. When --assume-role-arn is set, this returns an option that wraps the instance-role
// credentials in an STS AssumeRole provider for that role (with an ExternalID and a session
// name derived from the OOD username). When it is empty (the default / single-account on-ramp),
// the behavior is exactly the AWS default credential chain — no AssumeRole, instance role.
package awscfg

import (
	"context"
	"os"
	"os/user"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
)

// Options controls how the adapter obtains AWS credentials.
type Options struct {
	Region        string // AWS region
	AssumeRoleARN string // optional per-user role to assume; empty = use the instance role directly
	ExternalID    string // optional sts:ExternalId for the AssumeRole trust policy
	SessionName   string // optional RoleSessionName (e.g. the OOD username); defaults to "ood-adapter"
}

// runtimeUser returns the OS username the adapter is running as. OOD's PUN runs the adapter
// AS the logged-in user, so this is that user — used to expand {username} in a per-user role
// ARN and to name the STS session. Overridable via OOD_USER for non-PUN invocations.
func runtimeUser() string {
	if u := os.Getenv("OOD_USER"); u != "" {
		return u
	}
	if u, err := user.Current(); err == nil && u.Username != "" {
		return u.Username
	}
	return os.Getenv("USER")
}

// LoadOptions returns the config.LoadDefaultConfig option functions for these Options.
// Always sets the region; appends an AssumeRole credentials provider only when AssumeRoleARN
// is set. Empty AssumeRoleARN yields exactly the default chain (single-account behavior).
//
// #78: AssumeRoleARN may contain a "{username}" placeholder (e.g.
// arn:aws:iam::123:role/ood-user-{username}); it is expanded from the runtime user so each
// portal user assumes their OWN per-user role. The STS RoleSessionName defaults to that user.
func LoadOptions(ctx context.Context, o Options) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{config.WithRegion(o.Region)}
	if o.AssumeRoleARN == "" {
		return opts
	}

	uname := runtimeUser()
	roleARN := strings.ReplaceAll(o.AssumeRoleARN, "{username}", uname)

	// Base config (instance role) used only to call sts:AssumeRole into the per-user role.
	base, err := config.LoadDefaultConfig(ctx, config.WithRegion(o.Region))
	if err != nil {
		// Fall back to the default chain; the caller's LoadDefaultConfig will surface any error.
		return opts
	}
	stsClient := sts.NewFromConfig(base)
	sessionName := o.SessionName
	if sessionName == "" {
		sessionName = "ood-" + uname
	}
	if sessionName == "ood-" {
		sessionName = "ood-adapter"
	}
	provider := stscreds.NewAssumeRoleProvider(stsClient, roleARN, func(p *stscreds.AssumeRoleOptions) {
		p.RoleSessionName = sessionName
		if o.ExternalID != "" {
			p.ExternalID = aws.String(o.ExternalID)
		}
	})
	return append(opts, config.WithCredentialsProvider(aws.NewCredentialsCache(provider)))
}
