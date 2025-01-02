package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	StatusPending = "pending"
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

// PipelineRun represents a single run of a pipeline.
type PipelineRun struct {
	ID         string
	PipelineID string
	// GitRef is the git reference (branch) that is used for this run
	GitRef       string
	RunStatus    string
	BuildStatus  string
	DeployStatus string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	// Logs is a map of stage names to logs
	Logs map[string]string
}

func NewPipelineRun(pipelineID, gitRef string) *PipelineRun {
	now := time.Now()
	return &PipelineRun{
		ID:           uuid.New().String(),
		PipelineID:   pipelineID,
		GitRef:       gitRef,
		RunStatus:    StatusPending,
		BuildStatus:  StatusPending,
		DeployStatus: StatusPending,
		Status:       StatusPending,
		CreatedAt:    now,
		UpdatedAt:    now,
		Logs:         make(map[string]string),
	}
}
