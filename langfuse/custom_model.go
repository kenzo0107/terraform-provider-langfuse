package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type CustomModel struct {
	ID                string   `json:"id"`
	ModelName         string   `json:"modelName"`
	MatchPattern      string   `json:"matchPattern"`
	InputPrice        *float64 `json:"inputPrice,omitempty"`
	OutputPrice       *float64 `json:"outputPrice,omitempty"`
	TotalPrice        *float64 `json:"totalPrice,omitempty"`
	Unit              *string  `json:"unit,omitempty"`
	IsLangfuseManaged bool     `json:"isLangfuseManaged"`
}

type CreateCustomModelRequest struct {
	ModelName    string   `json:"modelName"`
	MatchPattern string   `json:"matchPattern"`
	InputPrice   *float64 `json:"inputPrice,omitempty"`
	OutputPrice  *float64 `json:"outputPrice,omitempty"`
	TotalPrice   *float64 `json:"totalPrice,omitempty"`
	Unit         *string  `json:"unit,omitempty"`
}

func (c *Client) GetCustomModel(ctx context.Context, id string) (*CustomModel, error) {
	path := fmt.Sprintf("/api/public/models/%s", id)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var m CustomModel
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("unmarshaling custom model response: %w", err)
	}

	return &m, nil
}

func (c *Client) CreateCustomModel(ctx context.Context, req *CreateCustomModelRequest) (*CustomModel, error) {
	body, err := c.do(ctx, http.MethodPost, "/api/public/models", req)
	if err != nil {
		return nil, err
	}

	var m CustomModel
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("unmarshaling create custom model response: %w", err)
	}

	return &m, nil
}

func (c *Client) DeleteCustomModel(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/public/models/%s", id)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
