package http

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/hphilipps/stagerunner/domain"
)

type API struct {
	store    domain.Store
	executor *domain.Executor
}

func NewAPI(store domain.Store, executor *domain.Executor) *API {
	return &API{
		store:    store,
		executor: executor,
	}
}

// SetupRouter configures all routes and middleware
func (api *API) SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Middleware
	r.Use(loggingMiddleware)
	r.Use(rateLimitMiddleware)
	r.Use(authMiddleware)

	// Pipeline routes
	r.HandleFunc("/pipelines", api.listPipelines).Methods(http.MethodGet)
	r.HandleFunc("/pipelines", api.createPipeline).Methods(http.MethodPost)
	r.HandleFunc("/pipelines/{id}", api.getPipeline).Methods(http.MethodGet)
	r.HandleFunc("/pipelines/{id}", api.updatePipeline).Methods(http.MethodPut)
	r.HandleFunc("/pipelines/{id}", api.deletePipeline).Methods(http.MethodDelete)
	r.HandleFunc("/pipelines/{id}/trigger", api.triggerPipeline).Methods(http.MethodPost)
	r.HandleFunc("/runs", api.listPipelineRuns).Methods(http.MethodGet)
	r.HandleFunc("/runs/{run_id}", api.getPipelineRun).Methods(http.MethodGet)

	return r
}

// Helper functions for HTTP responses
func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		log.Println(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}
