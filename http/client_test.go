package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient(t *testing.T) {
	// Setup test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check authorization header
		token := r.Header.Get("Authorization")
		if token != "test-token" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		switch r.URL.Path {
		case "/pipelines":
			switch r.Method {
			case http.MethodPost:
				var req PipelineRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				resp := PipelineResponse{
					ID:         "test-id",
					Name:       req.Name,
					Repository: req.Repository,
				}
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(resp)
			case http.MethodGet:
				pipelines := []PipelineResponse{
					{
						ID:         "test-id-1",
						Name:       "pipeline-1",
						Repository: "repo-1",
					},
					{
						ID:         "test-id-2",
						Name:       "pipeline-2",
						Repository: "repo-2",
					},
				}
				json.NewEncoder(w).Encode(pipelines)
			}

		case "/pipelines/test-id":
			switch r.Method {
			case http.MethodGet:
				resp := PipelineResponse{
					ID:         "test-id",
					Name:       "test-pipeline",
					Repository: "test-repo",
				}
				json.NewEncoder(w).Encode(resp)
			case http.MethodPut:
				var req PipelineRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				resp := PipelineResponse{
					ID:         "test-id",
					Name:       req.Name,
					Repository: req.Repository,
				}
				json.NewEncoder(w).Encode(resp)
			case http.MethodDelete:
				w.WriteHeader(http.StatusNoContent)
			}

		case "/pipelines/test-id/trigger":
			if r.Method == http.MethodPost {
				var req TriggerPipelineRequest
				if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				resp := TriggerPipelineResponse{
					ID: "run-id",
				}
				w.WriteHeader(http.StatusAccepted)
				json.NewEncoder(w).Encode(resp)
			}

		case "/runs":
			switch r.Method {
			case http.MethodGet:
				json.NewEncoder(w).Encode([]pipelineRunResponse{
					{
						ID: "run-id",
					},
				})
			}

		case "/runs/run-id":
			switch r.Method {
			case http.MethodGet:
				json.NewEncoder(w).Encode(pipelineRunResponse{
					ID: "run-id",
				})
			}

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// Create client
	client := NewClient(
		server.URL,
		WithToken("test-token"),
		WithTimeout(5*time.Second),
	)

	ctx := context.Background()

	t.Run("CreatePipeline", func(t *testing.T) {
		req := PipelineRequest{
			Name:       "test-pipeline",
			Repository: "test-repo",
			Stages: Stages{
				RunStage: RunStage{
					Command:       "go test ./...",
					ContinueOnErr: false,
				},
			},
		}

		resp, err := client.CreatePipeline(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "test-id", resp.ID)
		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, req.Repository, resp.Repository)
	})

	t.Run("GetPipeline", func(t *testing.T) {
		resp, err := client.GetPipeline(ctx, "test-id")
		require.NoError(t, err)
		assert.Equal(t, "test-id", resp.ID)
		assert.Equal(t, "test-pipeline", resp.Name)
		assert.Equal(t, "test-repo", resp.Repository)
	})

	t.Run("ListPipelines", func(t *testing.T) {
		resp, err := client.ListPipelines(ctx)
		require.NoError(t, err)
		assert.Len(t, resp, 2)
		assert.Equal(t, "test-id-1", resp[0].ID)
		assert.Equal(t, "test-id-2", resp[1].ID)
	})

	t.Run("UpdatePipeline", func(t *testing.T) {
		req := PipelineRequest{
			Name:       "updated-pipeline",
			Repository: "updated-repo",
		}

		resp, err := client.UpdatePipeline(ctx, "test-id", req)
		require.NoError(t, err)
		assert.Equal(t, "test-id", resp.ID)
		assert.Equal(t, req.Name, resp.Name)
		assert.Equal(t, req.Repository, resp.Repository)
	})

	t.Run("DeletePipeline", func(t *testing.T) {
		err := client.DeletePipeline(ctx, "test-id")
		require.NoError(t, err)
	})

	t.Run("TriggerPipeline", func(t *testing.T) {
		resp, err := client.TriggerPipeline(ctx, "test-id", "main")
		require.NoError(t, err)
		assert.Equal(t, "run-id", resp.ID)
	})

	t.Run("ListRuns", func(t *testing.T) {
		resp, err := client.ListRuns(ctx)
		require.NoError(t, err)
		assert.Len(t, resp, 1)
		assert.Equal(t, "run-id", resp[0].ID)
	})

	t.Run("GetRun", func(t *testing.T) {
		resp, err := client.GetRun(ctx, "run-id")
		require.NoError(t, err)
		assert.Equal(t, "run-id", resp.ID)
	})

	t.Run("UnauthorizedRequest", func(t *testing.T) {
		unauthorizedClient := NewClient(
			server.URL,
			WithToken("invalid-token"),
		)

		_, err := unauthorizedClient.ListPipelines(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unauthorized")
	})

	t.Run("InvalidEndpoint", func(t *testing.T) {
		_, err := client.GetPipeline(ctx, "non-existent")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "404")
	})

	t.Run("ClientTimeout", func(t *testing.T) {
		timeoutServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
		}))
		defer timeoutServer.Close()

		timeoutClient := NewClient(
			timeoutServer.URL,
			WithTimeout(50*time.Millisecond),
		)

		_, err := timeoutClient.ListPipelines(ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "deadline exceeded")
	})
}
