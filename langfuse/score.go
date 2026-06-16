package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Score struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Source        string   `json:"source"`
	DataType      *string  `json:"dataType,omitempty"`
	Value         *float64 `json:"value,omitempty"`
	StringValue   *string  `json:"stringValue,omitempty"`
	TraceID       *string  `json:"traceId,omitempty"`
	ObservationID *string  `json:"observationId,omitempty"`
	ConfigID      *string  `json:"configId,omitempty"`
	Comment       *string  `json:"comment,omitempty"`
	Environment   string   `json:"environment"`
}

type createScoreResponse struct {
	ID string `json:"id"`
}

type CreateScoreRequest struct {
	Name          string   `json:"name"`
	Value         *float64 `json:"value,omitempty"`
	StringValue   *string  `json:"stringValue,omitempty"`
	TraceID       *string  `json:"traceId,omitempty"`
	ObservationID *string  `json:"observationId,omitempty"`
	ConfigID      *string  `json:"configId,omitempty"`
	DataType      *string  `json:"dataType,omitempty"`
	Comment       *string  `json:"comment,omitempty"`
	Environment   *string  `json:"environment,omitempty"`
}

func (c *Client) GetScore(ctx context.Context, scoreID string) (*Score, error) {
	path := fmt.Sprintf("/api/public/scores/%s", scoreID)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var s Score
	if err := json.Unmarshal(body, &s); err != nil {
		return nil, fmt.Errorf("unmarshaling score response: %w", err)
	}

	return &s, nil
}

func (c *Client) CreateScore(ctx context.Context, req *CreateScoreRequest) (string, error) {
	body, err := c.do(ctx, http.MethodPost, "/api/public/scores", req)
	if err != nil {
		return "", err
	}

	var resp createScoreResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("unmarshaling create score response: %w", err)
	}

	return resp.ID, nil
}

func (c *Client) DeleteScore(ctx context.Context, scoreID string) error {
	path := fmt.Sprintf("/api/public/scores/%s", scoreID)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
