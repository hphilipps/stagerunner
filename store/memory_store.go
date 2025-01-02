package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/hphilipps/stagerunner/domain"
)

// MemoryStore implements Store interface using in-memory maps
type MemoryStore struct {
	pipelines    map[string]*domain.Pipeline
	pipelineRuns map[string]*domain.PipelineRun
	mu           sync.RWMutex
}

// NewMemoryStore creates a new instance of MemoryStore
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		pipelines:    make(map[string]*domain.Pipeline),
		pipelineRuns: make(map[string]*domain.PipelineRun),
	}
}

// CreatePipeline implements PipelineStore interface
func (s *MemoryStore) CreatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pipelines[pipeline.ID]; exists {
		return fmt.Errorf("%w: pipeline with ID %s already exists", domain.ErrAlreadyExists, pipeline.ID)
	}

	s.pipelines[pipeline.ID] = pipeline
	return nil
}

// GetPipeline implements PipelineStore interface
func (s *MemoryStore) GetPipeline(ctx context.Context, id string) (*domain.Pipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipeline, exists := s.pipelines[id]
	if !exists {
		return nil, fmt.Errorf("%w: pipeline with ID %s not found", domain.ErrNotFound, id)
	}
	return pipeline, nil
}

// UpdatePipeline implements PipelineStore interface
func (s *MemoryStore) UpdatePipeline(ctx context.Context, pipeline *domain.Pipeline) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pipelines[pipeline.ID]; !exists {
		return fmt.Errorf("%w: pipeline with ID %s not found", domain.ErrNotFound, pipeline.ID)
	}

	s.pipelines[pipeline.ID] = pipeline
	return nil
}

// DeletePipeline implements PipelineStore interface
func (s *MemoryStore) DeletePipeline(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pipelines[id]; !exists {
		return fmt.Errorf("%w: pipeline with ID %s not found", domain.ErrNotFound, id)
	}

	delete(s.pipelines, id)
	return nil
}

// ListPipelines implements PipelineStore interface
func (s *MemoryStore) ListPipelines(ctx context.Context) ([]*domain.Pipeline, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipelines := make([]*domain.Pipeline, 0, len(s.pipelines))
	for _, p := range s.pipelines {
		pipelines = append(pipelines, p)
	}
	return pipelines, nil
}

// CreatePipelineRun implements PipelineRunStore interface
func (s *MemoryStore) CreatePipelineRun(ctx context.Context, pipelineRun *domain.PipelineRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.pipelineRuns[pipelineRun.ID]; exists {
		return fmt.Errorf("%w: pipeline run with ID %s already exists", domain.ErrAlreadyExists, pipelineRun.ID)
	}

	s.pipelineRuns[pipelineRun.ID] = pipelineRun
	return nil
}

// GetPipelineRun implements PipelineRunStore interface
func (s *MemoryStore) GetPipelineRun(ctx context.Context, id string) (*domain.PipelineRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pipelineRun, exists := s.pipelineRuns[id]
	if !exists {
		return nil, fmt.Errorf("%w: pipeline run with ID %s not found", domain.ErrNotFound, id)
	}
	return pipelineRun, nil
}

// UpdatePipelineRun implements PipelineRunStore interface
func (s *MemoryStore) UpdatePipelineRun(ctx context.Context, run *domain.PipelineRun) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.pipelineRuns[run.ID]
	if !exists {
		return fmt.Errorf("%w: pipeline run with ID %s not found", domain.ErrNotFound, run.ID)
	}

	s.pipelineRuns[run.ID] = run
	return nil
}

// ListPipelineRuns implements PipelineRunStore interface
func (s *MemoryStore) ListPipelineRuns(ctx context.Context) ([]*domain.PipelineRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var runs []*domain.PipelineRun
	for _, run := range s.pipelineRuns {
		runs = append(runs, run)
	}
	return runs, nil
}
