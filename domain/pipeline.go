package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrQueueEmpty    = errors.New("queue is empty")
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

type Pipeline struct {
	ID         string
	Name       string
	Repository string
	Stages     map[string]Stage
}

func NewPipeline(repository string) *Pipeline {
	return &Pipeline{
		ID:         uuid.New().String(),
		Repository: repository,
		Stages: map[string]Stage{
			StageRun:    &RunStage{Name: StageRun, Command: "", ContOnError: false},
			StageBuild:  &BuildStage{Name: StageBuild, DockerfilePath: "", ContOnError: false},
			StageDeploy: &DeployStage{Name: StageDeploy, ClusterName: "", ManifestPath: "", ContOnError: false},
		},
	}
}
