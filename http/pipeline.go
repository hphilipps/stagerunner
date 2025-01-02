package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hphilipps/stagerunner/domain"
)

// RunStage is used to construct a run stage for requests and responses
type RunStage struct {
	Command       string `json:"command"`
	ContinueOnErr bool   `json:"continue_on_error"`
}

// BuildStage is used to construct a build stage for requests and responses
type BuildStage struct {
	DockerfilePath string `json:"dockerfile_path"`
	ContinueOnErr  bool   `json:"continue_on_error"`
}

// DeployStage is used to construct a deploy stage for requests and responses
type DeployStage struct {
	ClusterName   string `json:"cluster_name"`
	ManifestPath  string `json:"manifest_path"`
	ContinueOnErr bool   `json:"continue_on_error"`
}

// Stages is used to construct a pipeline with multiple stages for requests and responses
type Stages struct {
	RunStage    RunStage    `json:"run_stage"`
	BuildStage  BuildStage  `json:"build_stage"`
	DeployStage DeployStage `json:"deploy_stage"`
}

// PipelineRequest is used to construct a pipeline for requests
type PipelineRequest struct {
	Name       string `json:"name"`
	Repository string `json:"repository"`
	Stages     Stages `json:"stages"`
}

// PipelineResponse is used to construct a pipeline from a response
type PipelineResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Repository  string      `json:"repository"`
	RunStage    RunStage    `json:"run_stage"`
	BuildStage  BuildStage  `json:"build_stage"`
	DeployStage DeployStage `json:"deploy_stage"`
}

// String is a helper function to print the pipeline response in a friendly format
func (p *PipelineResponse) String() string {
	return fmt.Sprintf(`ID: %s
  Name: %s
  Repository: %s
  RunStage: %+v
  BuildStage: %+v 
  DeployStage: %+v`,
		p.ID,
		p.Name,
		p.Repository,
		p.RunStage,
		p.BuildStage,
		p.DeployStage)
}

// CreatePipelineResponse is used to construct a response for a pipeline creation request
type CreatePipelineResponse struct {
	ID string `json:"id"`
}

// createPipelineResponse is used to construct a pipeline response from a pipeline domain object
func createPipelineResponse(pipeline *domain.Pipeline) PipelineResponse {
	runStage := RunStage{
		Command:       pipeline.Stages[domain.StageRun].(*domain.RunStage).Command,
		ContinueOnErr: pipeline.Stages[domain.StageRun].(*domain.RunStage).ContOnError,
	}
	buildStage := BuildStage{
		DockerfilePath: pipeline.Stages[domain.StageBuild].(*domain.BuildStage).DockerfilePath,
		ContinueOnErr:  pipeline.Stages[domain.StageBuild].(*domain.BuildStage).ContOnError,
	}
	deployStage := DeployStage{
		ClusterName:   pipeline.Stages[domain.StageDeploy].(*domain.DeployStage).ClusterName,
		ManifestPath:  pipeline.Stages[domain.StageDeploy].(*domain.DeployStage).ManifestPath,
		ContinueOnErr: pipeline.Stages[domain.StageDeploy].(*domain.DeployStage).ContOnError,
	}

	return PipelineResponse{
		ID:          pipeline.ID,
		Name:        pipeline.Name,
		Repository:  pipeline.Repository,
		RunStage:    runStage,
		BuildStage:  buildStage,
		DeployStage: deployStage,
	}
}

// createPipeline is a handler for creating a pipeline
func (api *API) createPipeline(w http.ResponseWriter, r *http.Request) {
	var req PipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// validate request
	runStage := domain.NewRunStage(domain.StageRun, req.Stages.RunStage.Command, req.Stages.RunStage.ContinueOnErr)
	if err := runStage.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	buildStage := domain.NewBuildStage(domain.StageBuild, req.Stages.BuildStage.DockerfilePath, req.Stages.BuildStage.ContinueOnErr)
	if err := buildStage.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	deployStage := domain.NewDeployStage(domain.StageDeploy, req.Stages.DeployStage.ClusterName, req.Stages.DeployStage.ManifestPath, req.Stages.DeployStage.ContinueOnErr)
	if err := deployStage.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	pipeline := domain.NewPipeline(req.Repository)
	pipeline.Name = req.Name
	pipeline.Repository = req.Repository
	pipeline.Stages[domain.StageRun] = runStage
	pipeline.Stages[domain.StageBuild] = buildStage
	pipeline.Stages[domain.StageDeploy] = deployStage

	if err := api.store.CreatePipeline(r.Context(), pipeline); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusCreated, CreatePipelineResponse{ID: pipeline.ID})
}

// getPipeline is a handler for getting a pipeline
func (api *API) getPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	pipeline, err := api.store.GetPipeline(r.Context(), vars["id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Pipeline not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	resp := createPipelineResponse(pipeline)

	respondWithJSON(w, http.StatusOK, resp)
}

// updatePipeline is a handler for updating a pipeline
func (api *API) updatePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req PipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// get pipeline from store
	pipeline, err := api.store.GetPipeline(r.Context(), vars["id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Pipeline not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// validate request
	runStage := domain.NewRunStage(domain.StageRun, req.Stages.RunStage.Command, req.Stages.RunStage.ContinueOnErr)
	if err := runStage.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	buildStage := domain.NewBuildStage(domain.StageBuild, req.Stages.BuildStage.DockerfilePath, req.Stages.BuildStage.ContinueOnErr)
	if err := buildStage.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}
	deployStage := domain.NewDeployStage(domain.StageDeploy, req.Stages.DeployStage.ClusterName, req.Stages.DeployStage.ManifestPath, req.Stages.DeployStage.ContinueOnErr)
	if err := deployStage.Validate(); err != nil {
		respondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Update pipeline fields
	pipeline.Name = req.Name
	pipeline.Repository = req.Repository
	pipeline.Stages[domain.StageRun] = runStage
	pipeline.Stages[domain.StageBuild] = buildStage
	pipeline.Stages[domain.StageDeploy] = deployStage

	if err := api.store.UpdatePipeline(r.Context(), pipeline); err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusOK, pipeline)
}

// deletePipeline is a handler for deleting a pipeline
func (api *API) deletePipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	if err := api.store.DeletePipeline(r.Context(), vars["id"]); err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Pipeline not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// listPipelines is a handler for listing pipelines
func (api *API) listPipelines(w http.ResponseWriter, r *http.Request) {
	pipelines, err := api.store.ListPipelines(r.Context())
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	pipelineResponses := make([]PipelineResponse, 0, len(pipelines))
	for _, pipeline := range pipelines {
		pipelineResponses = append(pipelineResponses, createPipelineResponse(pipeline))
	}
	respondWithJSON(w, http.StatusOK, pipelineResponses)
}

// TriggerPipelineRequest is used to construct a request for triggering a pipeline
type TriggerPipelineRequest struct {
	// the branch in the repo we want to run on
	GitRef string `json:"git_ref"`
}

// TriggerPipelineResponse is used to construct a response for triggering a pipeline
type TriggerPipelineResponse struct {
	ID string `json:"id"`
}

// triggerPipeline is a handler for triggering a pipeline
func (api *API) triggerPipeline(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	var req TriggerPipelineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	pipeline, err := api.store.GetPipeline(r.Context(), vars["id"])
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			respondWithError(w, http.StatusNotFound, "Pipeline not found")
			return
		}
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	run, err := api.executor.TriggerPipeline(r.Context(), pipeline, req.GitRef)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondWithJSON(w, http.StatusAccepted, TriggerPipelineResponse{ID: run.ID})
}
