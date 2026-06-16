package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type AnnotationQueue struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Description    *string  `json:"description,omitempty"`
	ScoreConfigIDs []string `json:"scoreConfigIds"`
	ProjectID      string   `json:"projectId"`
}

type CreateAnnotationQueueRequest struct {
	Name           string   `json:"name"`
	Description    *string  `json:"description,omitempty"`
	ScoreConfigIDs []string `json:"scoreConfigIds,omitempty"`
}

type UpdateAnnotationQueueRequest struct {
	Name           *string  `json:"name,omitempty"`
	Description    *string  `json:"description,omitempty"`
	ScoreConfigIDs []string `json:"scoreConfigIds,omitempty"`
}

func (c *Client) GetAnnotationQueue(ctx context.Context, id string) (*AnnotationQueue, error) {
	path := fmt.Sprintf("/api/public/annotation-queues/%s", id)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var aq AnnotationQueue
	if err := json.Unmarshal(body, &aq); err != nil {
		return nil, fmt.Errorf("unmarshaling annotation queue response: %w", err)
	}

	return &aq, nil
}

func (c *Client) CreateAnnotationQueue(ctx context.Context, req *CreateAnnotationQueueRequest) (*AnnotationQueue, error) {
	body, err := c.do(ctx, http.MethodPost, "/api/public/annotation-queues", req)
	if err != nil {
		return nil, err
	}

	var aq AnnotationQueue
	if err := json.Unmarshal(body, &aq); err != nil {
		return nil, fmt.Errorf("unmarshaling create annotation queue response: %w", err)
	}

	return &aq, nil
}

func (c *Client) UpdateAnnotationQueue(ctx context.Context, id string, req *UpdateAnnotationQueueRequest) (*AnnotationQueue, error) {
	path := fmt.Sprintf("/api/public/annotation-queues/%s", id)
	body, err := c.do(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}

	var aq AnnotationQueue
	if err := json.Unmarshal(body, &aq); err != nil {
		return nil, fmt.Errorf("unmarshaling update annotation queue response: %w", err)
	}

	return &aq, nil
}

func (c *Client) DeleteAnnotationQueue(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/public/annotation-queues/%s", id)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
