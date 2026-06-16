package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ScoreConfigDataType string

const (
	ScoreConfigDataTypeNumeric     ScoreConfigDataType = "NUMERIC"
	ScoreConfigDataTypeBoolean     ScoreConfigDataType = "BOOLEAN"
	ScoreConfigDataTypeCategorical ScoreConfigDataType = "CATEGORICAL"
)

type ConfigCategory struct {
	Value float64 `json:"value"`
	Label string  `json:"label"`
}

type ScoreConfig struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	ProjectID   string              `json:"projectId"`
	DataType    ScoreConfigDataType `json:"dataType"`
	IsArchived  bool                `json:"isArchived"`
	MinValue    *float64            `json:"minValue,omitempty"`
	MaxValue    *float64            `json:"maxValue,omitempty"`
	Categories  []*ConfigCategory   `json:"categories,omitempty"`
	Description *string             `json:"description,omitempty"`
}

type listScoreConfigsResponse struct {
	Data []*ScoreConfig `json:"data"`
}

type CreateScoreConfigRequest struct {
	Name        string              `json:"name"`
	DataType    ScoreConfigDataType `json:"dataType"`
	Categories  []*ConfigCategory   `json:"categories,omitempty"`
	MinValue    *float64            `json:"minValue,omitempty"`
	MaxValue    *float64            `json:"maxValue,omitempty"`
	Description *string             `json:"description,omitempty"`
}

type UpdateScoreConfigRequest struct {
	IsArchived  *bool             `json:"isArchived,omitempty"`
	Name        *string           `json:"name,omitempty"`
	Categories  []*ConfigCategory `json:"categories,omitempty"`
	MinValue    *float64          `json:"minValue,omitempty"`
	MaxValue    *float64          `json:"maxValue,omitempty"`
	Description *string           `json:"description,omitempty"`
}

func (c *Client) GetScoreConfig(ctx context.Context, id string) (*ScoreConfig, error) {
	path := fmt.Sprintf("/api/public/score-configs/%s", id)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var sc ScoreConfig
	if err := json.Unmarshal(body, &sc); err != nil {
		return nil, fmt.Errorf("unmarshaling score config response: %w", err)
	}

	return &sc, nil
}

func (c *Client) CreateScoreConfig(ctx context.Context, req *CreateScoreConfigRequest) (*ScoreConfig, error) {
	body, err := c.do(ctx, http.MethodPost, "/api/public/score-configs", req)
	if err != nil {
		return nil, err
	}

	var sc ScoreConfig
	if err := json.Unmarshal(body, &sc); err != nil {
		return nil, fmt.Errorf("unmarshaling create score config response: %w", err)
	}

	return &sc, nil
}

func (c *Client) UpdateScoreConfig(ctx context.Context, id string, req *UpdateScoreConfigRequest) (*ScoreConfig, error) {
	path := fmt.Sprintf("/api/public/score-configs/%s", id)
	body, err := c.do(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}

	var sc ScoreConfig
	if err := json.Unmarshal(body, &sc); err != nil {
		return nil, fmt.Errorf("unmarshaling update score config response: %w", err)
	}

	return &sc, nil
}

func (c *Client) ArchiveScoreConfig(ctx context.Context, id string) error {
	archived := true
	_, err := c.UpdateScoreConfig(ctx, id, &UpdateScoreConfigRequest{IsArchived: &archived})
	return err
}

func (c *Client) ListScoreConfigs(ctx context.Context) ([]*ScoreConfig, error) {
	body, err := c.do(ctx, http.MethodGet, "/api/public/score-configs", nil)
	if err != nil {
		return nil, err
	}

	var resp listScoreConfigsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshaling score configs response: %w", err)
	}

	return resp.Data, nil
}
