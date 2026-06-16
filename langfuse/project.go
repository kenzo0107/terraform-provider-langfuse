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
