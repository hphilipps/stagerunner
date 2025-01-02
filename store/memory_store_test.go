package store

import (
	"context"
	"testing"

	"github.com/hphilipps/stagerunner/domain"
	"github.com/stretchr/testify/assert"
)

func TestMemoryStore_Pipeline(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	t.Run("CreatePipeline", func(t *testing.T) {
		pipeline := &domain.Pipeline{
			ID:   "test-pipeline",
			Name: "Test Pipeline",
		}

		// Test successful creation
		err := store.CreatePipeline(ctx, pipeline)
		assert.NoError(t, err)

		// Test duplicate creation
		err = store.CreatePipeline(ctx, pipeline)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAlreadyExists)
	})

	t.Run("GetPipeline", func(t *testing.T) {
		// Test successful retrieval
		pipeline, err := store.GetPipeline(ctx, "test-pipeline")
		assert.NoError(t, err)
		assert.Equal(t, "Test Pipeline", pipeline.Name)

		// Test non-existent pipeline
		_, err = store.GetPipeline(ctx, "non-existent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("UpdatePipeline", func(t *testing.T) {
		pipeline := &domain.Pipeline{
			ID:   "test-pipeline",
			Name: "Updated Pipeline",
		}

		// Test successful update
		err := store.UpdatePipeline(ctx, pipeline)
		assert.NoError(t, err)

		updated, err := store.GetPipeline(ctx, "test-pipeline")
		assert.NoError(t, err)
		assert.Equal(t, "Updated Pipeline", updated.Name)

		// Test update non-existent
		nonExistent := &domain.Pipeline{ID: "non-existent"}
		err = store.UpdatePipeline(ctx, nonExistent)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("ListPipelines", func(t *testing.T) {
		pipelines, err := store.ListPipelines(ctx)
		assert.NoError(t, err)
		assert.Len(t, pipelines, 1)
	})

	t.Run("DeletePipeline", func(t *testing.T) {
		// Test successful deletion
		err := store.DeletePipeline(ctx, "test-pipeline")
		assert.NoError(t, err)

		// Verify deletion
		_, err = store.GetPipeline(ctx, "test-pipeline")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
		// Test delete non-existent
		err = store.DeletePipeline(ctx, "non-existent")
		assert.Error(t, err)
	})
}

func TestMemoryStore_PipelineRun(t *testing.T) {
	store := NewMemoryStore()
	ctx := context.Background()

	// Create a test pipeline first
	pipeline := &domain.Pipeline{
		ID:   "test-pipeline",
		Name: "Test Pipeline",
	}
	_ = store.CreatePipeline(ctx, pipeline)

	t.Run("CreatePipelineRun", func(t *testing.T) {
		run := &domain.PipelineRun{
			ID:         "test-run",
			PipelineID: "test-pipeline",
			Status:     "pending",
		}

		// Test successful creation
		err := store.CreatePipelineRun(ctx, run)
		assert.NoError(t, err)

		// Test duplicate creation
		err = store.CreatePipelineRun(ctx, run)
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrAlreadyExists)
	})

	t.Run("GetPipelineRun", func(t *testing.T) {
		// Test successful retrieval
		run, err := store.GetPipelineRun(ctx, "test-run")
		assert.NoError(t, err)
		assert.Equal(t, "pending", run.Status)

		// Test non-existent run
		_, err = store.GetPipelineRun(ctx, "non-existent")
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("UpdatePipelineRun", func(t *testing.T) {
		// Test successful update
		err := store.UpdatePipelineRun(ctx, &domain.PipelineRun{
			ID:         "test-run",
			PipelineID: "test-pipeline",
			Status:     "running",
			Logs:       map[string]string{"build": "Building..."},
		})
		assert.NoError(t, err)

		updated, err := store.GetPipelineRun(ctx, "test-run")
		assert.NoError(t, err)
		assert.Equal(t, "running", updated.Status)
		assert.Equal(t, "Building...", updated.Logs["build"])
		assert.Equal(t, "test-pipeline", updated.PipelineID)

		// Test update non-existent
		err = store.UpdatePipelineRun(ctx, &domain.PipelineRun{
			ID:         "non-existent",
			PipelineID: "test-pipeline",
			Status:     "failed",
			Logs:       map[string]string{"test": "error"},
		})
		assert.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrNotFound)
	})

	t.Run("ListPipelineRuns", func(t *testing.T) {
		// Create another run for the same pipeline
		run2 := &domain.PipelineRun{
			ID:         "test-run-2",
			PipelineID: "test-pipeline",
			Status:     "pending",
		}
		_ = store.CreatePipelineRun(ctx, run2)

		// Test listing runs for a pipeline
		runs, err := store.ListPipelineRuns(ctx)
		assert.NoError(t, err)
		assert.Len(t, runs, 2)
	})
}
