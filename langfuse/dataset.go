package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Dataset struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	ProjectID   string  `json:"projectId"`
	Description *string `json:"description,omitempty"`
}

type CreateDatasetRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

func (c *Client) GetDataset(ctx context.Context, name string) (*Dataset, error) {
	path := fmt.Sprintf("/api/public/v2/datasets/%s", name)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var d Dataset
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, fmt.Errorf("unmarshaling dataset response: %w", err)
	}

	return &d, nil
}

func (c *Client) CreateDataset(ctx context.Context, req *CreateDatasetRequest) (*Dataset, error) {
	body, err := c.do(ctx, http.MethodPost, "/api/public/v2/datasets", req)
	if err != nil {
		return nil, err
	}

	var d Dataset
	if err := json.Unmarshal(body, &d); err != nil {
		return nil, fmt.Errorf("unmarshaling create dataset response: %w", err)
	}

	return &d, nil
}
