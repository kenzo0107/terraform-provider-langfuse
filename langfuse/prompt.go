package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type PromptResponse struct {
	Name        string        `json:"name"`
	Version     int           `json:"version"`
	Type        string        `json:"type"`
	TextContent string        `json:"prompt,omitempty"`
	Messages    []ChatMessage `json:"messages,omitempty"`
	Labels      []string      `json:"labels,omitempty"`
	Tags        []string      `json:"tags,omitempty"`
}

type CreatePromptRequest struct {
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	Text     string        `json:"prompt,omitempty"`
	Messages []ChatMessage `json:"messages,omitempty"`
	Labels   []string      `json:"labels,omitempty"`
	Tags     []string      `json:"tags,omitempty"`
}

func (c *Client) GetPrompt(ctx context.Context, name string, version int) (*PromptResponse, error) {
	path := fmt.Sprintf("/api/public/v2/prompts/%s?version=%d", name, version)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var p PromptResponse
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("unmarshaling prompt response: %w", err)
	}

	return &p, nil
}

func (c *Client) CreatePrompt(ctx context.Context, req *CreatePromptRequest) (*PromptResponse, error) {
	var payload any
	if req.Type == "chat" {
		payload = struct {
			Name     string        `json:"name"`
			Type     string        `json:"type"`
			Prompt   []ChatMessage `json:"prompt"`
			Labels   []string      `json:"labels,omitempty"`
			Tags     []string      `json:"tags,omitempty"`
		}{
			Name:   req.Name,
			Type:   req.Type,
			Prompt: req.Messages,
			Labels: req.Labels,
			Tags:   req.Tags,
		}
	} else {
		payload = struct {
			Name   string   `json:"name"`
			Type   string   `json:"type"`
			Prompt string   `json:"prompt"`
			Labels []string `json:"labels,omitempty"`
			Tags   []string `json:"tags,omitempty"`
		}{
			Name:   req.Name,
			Type:   req.Type,
			Prompt: req.Text,
			Labels: req.Labels,
			Tags:   req.Tags,
		}
	}

	body, err := c.do(ctx, http.MethodPost, "/api/public/v2/prompts", payload)
	if err != nil {
		return nil, err
	}

	var p PromptResponse
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, fmt.Errorf("unmarshaling create prompt response: %w", err)
	}

	return &p, nil
}

func (c *Client) DeletePrompt(ctx context.Context, name string) error {
	path := fmt.Sprintf("/api/public/v2/prompts/%s", name)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
