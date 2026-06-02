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

// LoadOptions returns the config.LoadDefaultConfig option functions for these Options.
// Always sets the region; appends an AssumeRole credentials provider only when AssumeRoleARN
// is set. Empty AssumeRoleARN yields exactly the default chain (single-account behavior).
func LoadOptions(ctx context.Context, o Options) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{config.WithRegion(o.Region)}
	if o.AssumeRoleARN == "" {
		return opts
	}

	// Base config (instance role) used only to call sts:AssumeRole into the per-user role.
	base, err := config.LoadDefaultConfig(ctx, config.WithRegion(o.Region))
	if err != nil {
		// Fall back to the default chain; the caller's LoadDefaultConfig will surface any error.
		return opts
	}
	stsClient := sts.NewFromConfig(base)
	sessionName := o.SessionName
	if sessionName == "" {
		sessionName = "ood-adapter"
	}
	provider := stscreds.NewAssumeRoleProvider(stsClient, o.AssumeRoleARN, func(p *stscreds.AssumeRoleOptions) {
		p.RoleSessionName = sessionName
		if o.ExternalID != "" {
			p.ExternalID = aws.String(o.ExternalID)
		}
	})
	return append(opts, config.WithCredentialsProvider(aws.NewCredentialsCache(provider)))
}
