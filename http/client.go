package http

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// ClientOption allows for customizing the client
type ClientOption func(*Client)

// WithTimeout sets a custom timeout for the HTTP client
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

// WithToken sets the authorization token
func WithToken(token string) ClientOption {
	return func(c *Client) {
		c.token = token
	}
}

// NewClient creates a new API client
func NewClient(baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CreatePipeline creates a new pipeline
func (c *Client) CreatePipeline(ctx context.Context, req PipelineRequest) (*PipelineResponse, error) {
	var resp PipelineResponse
	err := c.doRequest(ctx, http.MethodPost, "/pipelines", req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetPipeline retrieves a pipeline by ID
func (c *Client) GetPipeline(ctx context.Context, id string) (*PipelineResponse, error) {
	var resp PipelineResponse
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/pipelines/%s", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListPipelines retrieves all pipelines
func (c *Client) ListPipelines(ctx context.Context) ([]PipelineResponse, error) {
	var resp []PipelineResponse
	err := c.doRequest(ctx, http.MethodGet, "/pipelines", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// UpdatePipeline updates an existing pipeline
func (c *Client) UpdatePipeline(ctx context.Context, id string, req PipelineRequest) (*PipelineResponse, error) {
	var resp PipelineResponse
	err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/pipelines/%s", id), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// DeletePipeline deletes a pipeline by ID
func (c *Client) DeletePipeline(ctx context.Context, id string) error {
	return c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/pipelines/%s", id), nil, nil)
}

// TriggerPipeline triggers a pipeline run
func (c *Client) TriggerPipeline(ctx context.Context, id string, gitRef string) (*TriggerPipelineResponse, error) {
	req := TriggerPipelineRequest{GitRef: gitRef}
	var resp TriggerPipelineResponse
	err := c.doRequest(ctx, http.MethodPost, fmt.Sprintf("/pipelines/%s/trigger", id), req, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// ListRuns retrieves all pipeline runs
func (c *Client) ListRuns(ctx context.Context) ([]pipelineRunResponse, error) {
	var resp []pipelineRunResponse
	err := c.doRequest(ctx, http.MethodGet, "/runs", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// GetRun retrieves a pipeline run by ID
func (c *Client) GetRun(ctx context.Context, id string) (*pipelineRunResponse, error) {
	var resp pipelineRunResponse
	err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/runs/%s", id), nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp, nil
}

// Generic request handler
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}, response interface{}) error {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bytes.NewBuffer(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error string `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("request failed with status %d", resp.StatusCode)
		}
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, errResp.Error)
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}
