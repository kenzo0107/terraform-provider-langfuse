package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Evaluator struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Scope      string  `json:"scope"`
	Prompt     *string `json:"prompt,omitempty"`
	SourceCode *string `json:"sourceCode,omitempty"`
}

type CreateEvaluatorRequest struct {
	Name       string  `json:"name"`
	Type       string  `json:"type"`
	Prompt     *string `json:"prompt,omitempty"`
	SourceCode *string `json:"sourceCode,omitempty"`
}

type UpdateEvaluatorRequest struct {
	Name       *string `json:"name,omitempty"`
	Prompt     *string `json:"prompt,omitempty"`
	SourceCode *string `json:"sourceCode,omitempty"`
}

type evaluatorResponse struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Scope      string          `json:"scope"`
	Prompt     json.RawMessage `json:"prompt,omitempty"`
	SourceCode *string         `json:"sourceCode,omitempty"`
}

type createEvaluatorRequestRaw struct {
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Prompt     json.RawMessage `json:"prompt,omitempty"`
	SourceCode *string         `json:"sourceCode,omitempty"`
}

type updateEvaluatorRequestRaw struct {
	Name       *string         `json:"name,omitempty"`
	Prompt     json.RawMessage `json:"prompt,omitempty"`
	SourceCode *string         `json:"sourceCode,omitempty"`
}

func evaluatorFromResponse(r *evaluatorResponse) *Evaluator {
	return &Evaluator{
		ID:         r.ID,
		Name:       r.Name,
		Type:       r.Type,
		Scope:      r.Scope,
		Prompt:     rawMessageToStringPtrJSON(r.Prompt),
		SourceCode: r.SourceCode,
	}
}

func (c *Client) GetEvaluator(ctx context.Context, id string) (*Evaluator, error) {
	path := fmt.Sprintf("/api/public/unstable/evaluators/%s", id)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var r evaluatorResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling evaluator response: %w", err)
	}

	return evaluatorFromResponse(&r), nil
}

func (c *Client) CreateEvaluator(ctx context.Context, req *CreateEvaluatorRequest) (*Evaluator, error) {
	raw := createEvaluatorRequestRaw{
		Name:       req.Name,
		Type:       req.Type,
		Prompt:     stringPtrToRawMessage(req.Prompt),
		SourceCode: req.SourceCode,
	}

	body, err := c.do(ctx, http.MethodPost, "/api/public/unstable/evaluators", raw)
	if err != nil {
		return nil, err
	}

	var r evaluatorResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling create evaluator response: %w", err)
	}

	return evaluatorFromResponse(&r), nil
}

func (c *Client) UpdateEvaluator(ctx context.Context, id string, req *UpdateEvaluatorRequest) (*Evaluator, error) {
	raw := updateEvaluatorRequestRaw{
		Name:       req.Name,
		Prompt:     stringPtrToRawMessage(req.Prompt),
		SourceCode: req.SourceCode,
	}

	path := fmt.Sprintf("/api/public/unstable/evaluators/%s", id)
	body, err := c.do(ctx, http.MethodPatch, path, raw)
	if err != nil {
		return nil, err
	}

	var r evaluatorResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling update evaluator response: %w", err)
	}

	return evaluatorFromResponse(&r), nil
}

func (c *Client) DeleteEvaluator(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/public/unstable/evaluators/%s", id)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
