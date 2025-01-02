package domain

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var (
	queueSize     = 5
	pipelineLimit = 2
)

func TestExecutor_Integration(t *testing.T) {
	// Setup
	store := NewMemoryStore()
	executor := NewExecutor(store, 2, queueSize, pipelineLimit, 0.0, 10*time.Millisecond)

	// Start the executor
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	go executor.Start(ctx)

	// Create a test pipeline
	pipeline := &Pipeline{
		ID:         "test-pipeline",
		Name:       "Test Pipeline",
		Repository: "github.com/test/repo",
		Stages: map[string]Stage{
			StageRun:    &RunStage{Name: StageRun, Command: "go test ./...", ContOnError: false},
			StageBuild:  &BuildStage{Name: StageBuild, DockerfilePath: "Dockerfile", ContOnError: false},
			StageDeploy: &DeployStage{Name: StageDeploy, ClusterName: "test-cluster", ManifestPath: "k8s/", ContOnError: false},
		},
	}

	err := store.CreatePipeline(ctx, pipeline)
	assert.NoError(t, err)

	t.Run("TriggerPipeline creates new run", func(t *testing.T) {
		gitRef := "main"
		run, err := executor.TriggerPipeline(ctx, pipeline, gitRef)
		assert.NoError(t, err)
		assert.NotNil(t, run)
		assert.Equal(t, pipeline.ID, run.PipelineID)
		assert.Equal(t, gitRef, run.GitRef)
		assert.Equal(t, StatusPending, run.Status)

		// wait for run to finish
		time.Sleep(2 * time.Second)

		// check run status
		updatedRun, err := store.GetPipelineRun(ctx, run.ID)
		assert.NoError(t, err)
		assert.Equal(t, StatusSuccess, updatedRun.Status)
		assert.Contains(t, updatedRun.Logs[StageDeploy], "finished")
	})

	t.Run("Max per pipeline limits are respected", func(t *testing.T) {
		// Try to queue more runs than allowed
		for i := 0; i < pipelineLimit+1; i++ {
			_, err := executor.TriggerPipeline(ctx, pipeline, "branch")
			if i == pipelineLimit {
				assert.Error(t, err, "Should error when queue limit is reached")
			} else {
				assert.NoError(t, err)
			}
		}
	})

	t.Run("Queue limits are respected", func(t *testing.T) {
		// create new store and executor
		store = NewMemoryStore()
		executor = NewExecutor(store, 1, queueSize, pipelineLimit, 0.0, 10*time.Millisecond)

		// create different pipelines
		pipelines := []*Pipeline{}
		for i := 0; i < queueSize+1; i++ {
			pipelines = append(pipelines, &Pipeline{
				ID:         fmt.Sprintf("pipeline-%d", i),
				Name:       fmt.Sprintf("Pipeline %d", i),
				Repository: fmt.Sprintf("github.com/test/repo%d", i),
			})
		}

		for j, pipeline := range pipelines {
			_, err := executor.TriggerPipeline(ctx, pipeline, "branch")
			if j >= queueSize {
				assert.Error(t, err, fmt.Sprintf("Should error when pipeline limit is reached for pipeline %d", j))
			} else {
				assert.NoError(t, err)
			}
		}
	})
}
