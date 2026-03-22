// Package ood defines the OOD job spec types shared across adapters.
package ood

// JobStatus maps adapter-specific states to OOD status strings.
type JobStatus struct {
	ID       string `json:"id"`
	Status   string `json:"status"`
	ExitCode int    `json:"exit_code,omitempty"`
	Message  string `json:"message,omitempty"`
}

// OOD status constants
const (
	StatusQueued    = "queued"
	StatusRunning   = "running"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
	StatusCancelled = "cancelled"
	StatusUnknown   = "undetermined"
)
