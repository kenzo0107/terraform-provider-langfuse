package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type EvaluationRule struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	State       string   `json:"state"`
	Target      string   `json:"target"`
	EvaluatorID string   `json:"evaluatorId"`
	Filter      *string  `json:"filter,omitempty"`
	Mapping     *string  `json:"mapping,omitempty"`
	Sampling    *float64 `json:"sampling,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
}

type CreateEvaluationRuleRequest struct {
	Name        string   `json:"name"`
	Target      string   `json:"target"`
	EvaluatorID string   `json:"evaluatorId"`
	Filter      *string  `json:"filter,omitempty"`
	Mapping     *string  `json:"mapping,omitempty"`
	Sampling    *float64 `json:"sampling,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
}

type UpdateEvaluationRuleRequest struct {
	Name     *string  `json:"name,omitempty"`
	State    *string  `json:"state,omitempty"`
	Filter   *string  `json:"filter,omitempty"`
	Mapping  *string  `json:"mapping,omitempty"`
	Sampling *float64 `json:"sampling,omitempty"`
	Priority *int     `json:"priority,omitempty"`
}

type evaluationRuleResponse struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	State       string          `json:"state"`
	Target      string          `json:"target"`
	EvaluatorID string          `json:"evaluatorId"`
	Filter      json.RawMessage `json:"filter,omitempty"`
	Mapping     json.RawMessage `json:"mapping,omitempty"`
	Sampling    *float64        `json:"sampling,omitempty"`
	Priority    *int            `json:"priority,omitempty"`
}

type createEvaluationRuleRequestRaw struct {
	Name        string          `json:"name"`
	Target      string          `json:"target"`
	EvaluatorID string          `json:"evaluatorId"`
	Filter      json.RawMessage `json:"filter,omitempty"`
	Mapping     json.RawMessage `json:"mapping,omitempty"`
	Sampling    *float64        `json:"sampling,omitempty"`
	Priority    *int            `json:"priority,omitempty"`
}

type updateEvaluationRuleRequestRaw struct {
	Name     *string         `json:"name,omitempty"`
	State    *string         `json:"state,omitempty"`
	Filter   json.RawMessage `json:"filter,omitempty"`
	Mapping  json.RawMessage `json:"mapping,omitempty"`
	Sampling *float64        `json:"sampling,omitempty"`
	Priority *int            `json:"priority,omitempty"`
}

func rawMessageToStringPtrJSON(raw json.RawMessage) *string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	b, err := json.Marshal(raw)
	if err != nil {
		return nil
	}
	// RawMessage is already the JSON bytes; convert directly
	s := string(raw)
	_ = b
	return &s
}

func evaluationRuleFromResponse(r *evaluationRuleResponse) *EvaluationRule {
	return &EvaluationRule{
		ID:          r.ID,
		Name:        r.Name,
		State:       r.State,
		Target:      r.Target,
		EvaluatorID: r.EvaluatorID,
		Filter:      rawMessageToStringPtrJSON(r.Filter),
		Mapping:     rawMessageToStringPtrJSON(r.Mapping),
		Sampling:    r.Sampling,
		Priority:    r.Priority,
	}
}

func (c *Client) GetEvaluationRule(ctx context.Context, id string) (*EvaluationRule, error) {
	path := fmt.Sprintf("/api/public/unstable/evaluation-rules/%s", id)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var r evaluationRuleResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling evaluation rule response: %w", err)
	}

	return evaluationRuleFromResponse(&r), nil
}

func (c *Client) CreateEvaluationRule(ctx context.Context, req *CreateEvaluationRuleRequest) (*EvaluationRule, error) {
	raw := createEvaluationRuleRequestRaw{
		Name:        req.Name,
		Target:      req.Target,
		EvaluatorID: req.EvaluatorID,
		Filter:      stringPtrToRawMessage(req.Filter),
		Mapping:     stringPtrToRawMessage(req.Mapping),
		Sampling:    req.Sampling,
		Priority:    req.Priority,
	}

	body, err := c.do(ctx, http.MethodPost, "/api/public/unstable/evaluation-rules", raw)
	if err != nil {
		return nil, err
	}

	var r evaluationRuleResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling create evaluation rule response: %w", err)
	}

	return evaluationRuleFromResponse(&r), nil
}

func (c *Client) UpdateEvaluationRule(ctx context.Context, id string, req *UpdateEvaluationRuleRequest) (*EvaluationRule, error) {
	raw := updateEvaluationRuleRequestRaw{
		Name:     req.Name,
		State:    req.State,
		Filter:   stringPtrToRawMessage(req.Filter),
		Mapping:  stringPtrToRawMessage(req.Mapping),
		Sampling: req.Sampling,
		Priority: req.Priority,
	}

	path := fmt.Sprintf("/api/public/unstable/evaluation-rules/%s", id)
	body, err := c.do(ctx, http.MethodPatch, path, raw)
	if err != nil {
		return nil, err
	}

	var r evaluationRuleResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling update evaluation rule response: %w", err)
	}

	return evaluationRuleFromResponse(&r), nil
}

func (c *Client) DeleteEvaluationRule(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/public/unstable/evaluation-rules/%s", id)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
