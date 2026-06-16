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

type AnnotationQueueItem struct {
	ID            string  `json:"id"`
	QueueID       string  `json:"queueId"`
	TraceID       string  `json:"traceId"`
	ObservationID *string `json:"observationId,omitempty"`
	Status        string  `json:"status"`
}

type CreateAnnotationQueueItemRequest struct {
	TraceID       string  `json:"traceId"`
	ObservationID *string `json:"observationId,omitempty"`
}

type UpdateAnnotationQueueItemRequest struct {
	Status string `json:"status"`
}

func (c *Client) GetAnnotationQueueItem(ctx context.Context, queueID, itemID string) (*AnnotationQueueItem, error) {
	path := fmt.Sprintf("/api/public/annotation-queues/%s/items/%s", queueID, itemID)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var item AnnotationQueueItem
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("unmarshaling annotation queue item response: %w", err)
	}

	return &item, nil
}

func (c *Client) CreateAnnotationQueueItem(ctx context.Context, queueID string, req *CreateAnnotationQueueItemRequest) (*AnnotationQueueItem, error) {
	path := fmt.Sprintf("/api/public/annotation-queues/%s/items", queueID)
	body, err := c.do(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}

	var item AnnotationQueueItem
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("unmarshaling create annotation queue item response: %w", err)
	}

	return &item, nil
}

func (c *Client) UpdateAnnotationQueueItem(ctx context.Context, queueID, itemID string, req *UpdateAnnotationQueueItemRequest) (*AnnotationQueueItem, error) {
	path := fmt.Sprintf("/api/public/annotation-queues/%s/items/%s", queueID, itemID)
	body, err := c.do(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}

	var item AnnotationQueueItem
	if err := json.Unmarshal(body, &item); err != nil {
		return nil, fmt.Errorf("unmarshaling update annotation queue item response: %w", err)
	}

	return &item, nil
}

func (c *Client) DeleteAnnotationQueueItem(ctx context.Context, queueID, itemID string) error {
	path := fmt.Sprintf("/api/public/annotation-queues/%s/items/%s", queueID, itemID)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
