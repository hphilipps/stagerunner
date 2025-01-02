package domain

import "context"

// PipelineStore supports basic CRUD operations for pipelines.
type PipelineStore interface {
	CreatePipeline(ctx context.Context, pipeline *Pipeline) error
	GetPipeline(ctx context.Context, id string) (*Pipeline, error)
	UpdatePipeline(ctx context.Context, pipeline *Pipeline) error
	DeletePipeline(ctx context.Context, id string) error
	ListPipelines(ctx context.Context) ([]*Pipeline, error)
}

// PipelineRunStore supports basic CRUD operations for pipeline runs.
type PipelineRunStore interface {
	CreatePipelineRun(ctx context.Context, pipelineRun *PipelineRun) error
	GetPipelineRun(ctx context.Context, id string) (*PipelineRun, error)
	UpdatePipelineRun(ctx context.Context, run *PipelineRun) error
	ListPipelineRuns(ctx context.Context) ([]*PipelineRun, error)
}

// Store is an interface for storing Pipelines and PipelineRuns.
// For simplicity, we're providing a single interface here.
type Store interface {
	PipelineStore
	PipelineRunStore
}
