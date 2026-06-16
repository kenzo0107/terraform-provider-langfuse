package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Project struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listProjectsResponse struct {
	Data []Project `json:"data"`
}

type createProjectRequest struct {
	Name string `json:"name"`
}

type updateProjectRequest struct {
	Name string `json:"name"`
}

func (c *Client) GetProject(ctx context.Context, id string) (*Project, error) {
	projects, err := c.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.ID == id {
			return &p, nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Body: fmt.Sprintf("project %q not found", id)}
}

func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	body, err := c.do(ctx, http.MethodGet, "/api/public/projects", nil)
	if err != nil {
		return nil, err
	}

	var resp listProjectsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshaling projects response: %w", err)
	}

	return resp.Data, nil
}

func (c *Client) CreateProject(ctx context.Context, name string) (*Project, error) {
	payload := createProjectRequest{Name: name}
	body, err := c.do(ctx, http.MethodPost, "/api/public/projects", payload)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("unmarshaling create project response: %w", err)
	}

	return &project, nil
}

func (c *Client) UpdateProject(ctx context.Context, id, name string) (*Project, error) {
	payload := updateProjectRequest{Name: name}
	body, err := c.do(ctx, http.MethodPatch, "/api/public/projects/"+id, payload)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := json.Unmarshal(body, &project); err != nil {
		return nil, fmt.Errorf("unmarshaling update project response: %w", err)
	}

	return &project, nil
}

func (c *Client) DeleteProject(ctx context.Context, id string) error {
	_, err := c.do(ctx, http.MethodDelete, "/api/public/projects/"+id, nil)
	return err
}

type APIKey struct {
	ID               string  `json:"id"`
	PublicKey        string  `json:"publicKey"`
	DisplaySecretKey string  `json:"displaySecretKey"`
	Note             *string `json:"note,omitempty"`
}

type APIKeyCreated struct {
	ID               string  `json:"id"`
	PublicKey        string  `json:"publicKey"`
	SecretKey        string  `json:"secretKey"`
	DisplaySecretKey string  `json:"displaySecretKey"`
	Note             *string `json:"note,omitempty"`
}

type listAPIKeysResponse struct {
	APIKeys []APIKey `json:"apiKeys"`
}

type createAPIKeyRequest struct {
	Note *string `json:"note,omitempty"`
}

func (c *Client) GetProjectAPIKeys(ctx context.Context, projectID string) ([]APIKey, error) {
	path := fmt.Sprintf("/api/public/projects/%s/api-keys", projectID)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp listAPIKeysResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshaling API keys response: %w", err)
	}

	return resp.APIKeys, nil
}

func (c *Client) CreateProjectAPIKey(ctx context.Context, projectID string, note *string) (*APIKeyCreated, error) {
	path := fmt.Sprintf("/api/public/projects/%s/api-keys", projectID)
	payload := createAPIKeyRequest{Note: note}
	body, err := c.do(ctx, http.MethodPost, path, payload)
	if err != nil {
		return nil, err
	}

	var key APIKeyCreated
	if err := json.Unmarshal(body, &key); err != nil {
		return nil, fmt.Errorf("unmarshaling create API key response: %w", err)
	}

	return &key, nil
}

func (c *Client) DeleteProjectAPIKey(ctx context.Context, projectID, apiKeyID string) error {
	path := fmt.Sprintf("/api/public/projects/%s/api-keys/%s", projectID, apiKeyID)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
