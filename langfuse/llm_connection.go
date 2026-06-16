package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type LLMConnection struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Provider          string  `json:"provider"`
	BaseURL           *string `json:"baseURL,omitempty"`
	WithDefaultModels bool    `json:"withDefaultModels"`
}

type UpsertLLMConnectionRequest struct {
	Name              string  `json:"name"`
	Provider          string  `json:"provider"`
	BaseURL           *string `json:"baseURL,omitempty"`
	APIKey            *string `json:"apiKey,omitempty"`
	WithDefaultModels *bool   `json:"withDefaultModels,omitempty"`
}

type listLLMConnectionsResponse struct {
	Data []LLMConnection `json:"data"`
}

func (c *Client) ListLLMConnections(ctx context.Context) ([]LLMConnection, error) {
	body, err := c.do(ctx, http.MethodGet, "/api/public/llm-connections", nil)
	if err != nil {
		return nil, err
	}

	var resp listLLMConnectionsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshaling LLM connections response: %w", err)
	}

	return resp.Data, nil
}

func (c *Client) GetLLMConnectionByName(ctx context.Context, name string) (*LLMConnection, error) {
	connections, err := c.ListLLMConnections(ctx)
	if err != nil {
		return nil, err
	}
	for _, conn := range connections {
		if conn.Name == name {
			return &conn, nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Body: fmt.Sprintf("LLM connection %q not found", name)}
}

func (c *Client) UpsertLLMConnection(ctx context.Context, req *UpsertLLMConnectionRequest) (*LLMConnection, error) {
	body, err := c.do(ctx, http.MethodPut, "/api/public/llm-connections", req)
	if err != nil {
		return nil, err
	}

	var conn LLMConnection
	if err := json.Unmarshal(body, &conn); err != nil {
		return nil, fmt.Errorf("unmarshaling upsert LLM connection response: %w", err)
	}

	return &conn, nil
}

func (c *Client) DeleteLLMConnection(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/public/llm-connections/%s", id)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
