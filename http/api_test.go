package http

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/hphilipps/stagerunner/domain"
	"github.com/hphilipps/stagerunner/store"
	"github.com/stretchr/testify/assert"
)

func TestApi_Pipeline(t *testing.T) {
	store := store.NewMemoryStore()
	executor := domain.NewExecutor(store, 2, 5, 2, 0.0, 10*time.Millisecond)
	api := NewAPI(store, executor)

	id := ""
	runID := ""
	t.Run("CreatePipeline", func(t *testing.T) {
		tests := []struct {
			name       string
			payload    PipelineRequest
			wantStatus int
		}{
			{
				name: "successful creation",
				payload: PipelineRequest{
					Name:       "test-pipeline",
					Repository: "github.com/test/repo",
					Stages: Stages{
						RunStage: RunStage{
							Command:       "go test ./...",
							ContinueOnErr: false,
						},
						BuildStage: BuildStage{
							DockerfilePath: "Dockerfile",
							ContinueOnErr:  false,
						},
						DeployStage: DeployStage{
							ClusterName:   "prod",
							ManifestPath:  "k8s/",
							ContinueOnErr: false,
						},
					},
				},
				wantStatus: http.StatusCreated,
			},
			{
				name: "invalid payload",
				payload: PipelineRequest{
					Name: "test-pipeline",
				},
				wantStatus: http.StatusBadRequest,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {

				payload, err := json.Marshal(tt.payload)
				if err != nil {
					t.Fatalf("failed to marshal request body: %v", err)
				}
				req := httptest.NewRequest(http.MethodPost, "/pipelines", bytes.NewBuffer(payload))
				req.Header.Set("Authorization", "test-token")
				w := httptest.NewRecorder()

				router := api.SetupRouter()
				router.ServeHTTP(w, req)

				body, err := io.ReadAll(w.Body)
				if err != nil {
					t.Fatalf("failed to read response body: %v", err)
				}

				var pipeline CreatePipelineResponse
				if err := json.Unmarshal(body, &pipeline); err != nil {
					t.Fatalf("failed to unmarshal response body: %v", err)
				}

				assert.Equal(t, tt.wantStatus, w.Code)
			})
		}
	})

	t.Run("ListPipelines", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pipelines", nil)
		req.Header.Set("Authorization", "test-token")
		w := httptest.NewRecorder()

		router := api.SetupRouter()
		router.ServeHTTP(w, req)

		body, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		var pipelines []PipelineResponse
		if err := json.Unmarshal(body, &pipelines); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}

		if len(pipelines) != 1 {
			t.Fatalf("expected 1 pipeline, got %d", len(pipelines))
		}

		id = pipelines[0].ID

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "test-pipeline", pipelines[0].Name)
		assert.Equal(t, "github.com/test/repo", pipelines[0].Repository)
		assert.Equal(t, id, pipelines[0].ID)
	})

	t.Run("GetPipeline", func(t *testing.T) {

		tests := []struct {
			name       string
			pipelineID string
			wantStatus int
		}{
			{
				name:       "pipeline found",
				pipelineID: id,
				wantStatus: http.StatusOK,
			},
			{
				name:       "pipeline not found",
				pipelineID: "456",
				wantStatus: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, "/pipelines/"+tt.pipelineID, nil)
				req.Header.Set("Authorization", "test-token")
				w := httptest.NewRecorder()

				router := api.SetupRouter()
				router.ServeHTTP(w, req)

				assert.Equal(t, tt.wantStatus, w.Code)
			})
		}
	})

	t.Run("UpdatePipeline", func(t *testing.T) {
		pipeline := PipelineRequest{
			Name:       "test-pipeline-updated",
			Repository: "github.com/test/repo-updated",
			Stages: Stages{
				RunStage: RunStage{
					Command:       "go test -v ./...",
					ContinueOnErr: false,
				},
				BuildStage: BuildStage{
					DockerfilePath: "Dockerfile",
					ContinueOnErr:  false,
				},
				DeployStage: DeployStage{
					ClusterName:   "prod",
					ManifestPath:  "k8s/",
					ContinueOnErr: false,
				},
			},
		}
		payload, err := json.Marshal(pipeline)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}

		req := httptest.NewRequest(http.MethodPut, "/pipelines/"+id, bytes.NewBuffer(payload))
		req.Header.Set("Authorization", "test-token")
		req.Body = io.NopCloser(bytes.NewBuffer(payload))
		w := httptest.NewRecorder()

		router := api.SetupRouter()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if p, err := api.store.GetPipeline(context.Background(), id); err != nil {
			t.Fatalf("failed to get pipeline: %v", err)
		} else {
			assert.Equal(t, pipeline.Name, p.Name)
			assert.Equal(t, pipeline.Repository, p.Repository)
		}
	})

	t.Run("TriggerPipeline", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pipelines/"+id+"/trigger", nil)
		req.Header.Set("Authorization", "test-token")
		req.Body = io.NopCloser(bytes.NewBufferString(`{"git_ref": "test-ref"}`))
		w := httptest.NewRecorder()

		router := api.SetupRouter()
		router.ServeHTTP(w, req)

		body, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		var response TriggerPipelineResponse
		if err := json.Unmarshal(body, &response); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}

		runID = response.ID

		assert.Equal(t, http.StatusAccepted, w.Code)
		assert.NotEmpty(t, response.ID)
	})

	t.Run("GetPipelineRun", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/runs/"+runID, nil)
		req.Header.Set("Authorization", "test-token")
		w := httptest.NewRecorder()

		router := api.SetupRouter()
		router.ServeHTTP(w, req)

		body, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		var response pipelineRunResponse
		if err := json.Unmarshal(body, &response); err != nil {
			log.Println(string(body))
		}

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, runID, response.ID)
		assert.Equal(t, id, response.PipelineID)
	})

	t.Run("ListPipelineRuns", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/runs", nil)
		req.Header.Set("Authorization", "test-token")
		w := httptest.NewRecorder()

		router := api.SetupRouter()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		body, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		var runs []pipelineRunResponse
		if err := json.Unmarshal(body, &runs); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 1, len(runs))
		assert.Equal(t, runID, runs[0].ID)
		assert.Equal(t, id, runs[0].PipelineID)
	})

	t.Run("DeletePipeline", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/pipelines/"+id, nil)
		req.Header.Set("Authorization", "test-token")
		w := httptest.NewRecorder()

		router := api.SetupRouter()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("DeletePipelineNotFound", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/pipelines/123", nil)
		req.Header.Set("Authorization", "test-token")
		w := httptest.NewRecorder()

		router := api.SetupRouter()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("ListPipelinesAfterDelete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pipelines", nil)
		req.Header.Set("Authorization", "test-token")
		w := httptest.NewRecorder()

		router := api.SetupRouter()
		router.ServeHTTP(w, req)

		body, err := io.ReadAll(w.Body)
		if err != nil {
			t.Fatalf("failed to read response body: %v", err)
		}

		var pipelines []PipelineResponse
		if err := json.Unmarshal(body, &pipelines); err != nil {
			t.Fatalf("failed to unmarshal response body: %v", err)
		}

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, 0, len(pipelines))
	})
}
