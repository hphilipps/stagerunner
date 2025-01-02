package http

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/hphilipps/stagerunner/domain"
)

// pipelineRunResponse is used to construct a response for a pipeline run
type pipelineRunResponse struct {
	ID           string            `json:"id"`
	PipelineID   string            `json:"pipeline_id"`
	GitRef       string            `json:"git_ref"`
	Status       string            `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	RunStatus    string            `json:"run_status"`
	BuildStatus  string            `json:"build_status"`
	DeployStatus string            `json:"deploy_status"`
	Logs         map[string]string `json:"logs"`
}

// String is a helper function to print the pipeline run response in a friendly format
func (p *pipelineRunResponse) String() string {
	return fmt.Sprintf(`ID: %s
  PipelineID: %s
  GitRef: %s
  Status: %s
  CreatedAt: %s
  UpdatedAt: %s
  RunStatus: %s
  BuildStatus: %s
  DeployStatus: %s
  Logs: %+v`,
		p.ID,
		p.PipelineID,
		p.GitRef,
		p.Status,
		p.CreatedAt,
		p.UpdatedAt,
		p.RunStatus,
		p.BuildStatus,
		p.DeployStatus,
		p.Logs)
}

// createPipelineRunResponse is used to construct a pipeline run response from a pipeline run domain object
func createPipelineRunResponse(run *domain.PipelineRun) pipelineRunResponse {
	return pipelineRunResponse{
		ID:           run.ID,
		PipelineID:   run.PipelineID,
		GitRef:       run.GitRef,
		Status:       run.Status,
		CreatedAt:    run.CreatedAt,
		UpdatedAt:    run.UpdatedAt,
		RunStatus:    run.RunStatus,
		BuildStatus:  run.BuildStatus,
		DeployStatus: run.DeployStatus,
		Logs:         run.Logs,
	}
}

// getPipelineRun is a handler for getting a pipeline run
func (api *API) getPipelineRun(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	run, err := api.store.GetPipelineRun(r.Context(), vars["run_id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Pipeline run not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, createPipelineRunResponse(run))
}

// listPipelineRuns is a handler for listing pipeline runs
func (api *API) listPipelineRuns(w http.ResponseWriter, r *http.Request) {
	runs, err := api.store.ListPipelineRuns(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	runResponses := make([]pipelineRunResponse, 0, len(runs))
	for _, run := range runs {
		runResponses = append(runResponses, createPipelineRunResponse(run))
	}
	respondWithJSON(w, http.StatusOK, runResponses)
}
