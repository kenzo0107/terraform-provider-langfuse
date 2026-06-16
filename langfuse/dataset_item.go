package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type DatasetItem struct {
	ID             string  `json:"id"`
	DatasetName    string  `json:"datasetName"`
	Status         string  `json:"status"`
	Input          *string `json:"input,omitempty"`
	ExpectedOutput *string `json:"expectedOutput,omitempty"`
}

type CreateDatasetItemRequest struct {
	DatasetName    string  `json:"datasetName"`
	Input          *string `json:"input,omitempty"`
	ExpectedOutput *string `json:"expectedOutput,omitempty"`
	Status         string  `json:"status,omitempty"`
	ID             *string `json:"id,omitempty"`
}

type datasetItemResponse struct {
	ID             string          `json:"id"`
	DatasetName    string          `json:"datasetName"`
	Status         string          `json:"status"`
	Input          json.RawMessage `json:"input,omitempty"`
	ExpectedOutput json.RawMessage `json:"expectedOutput,omitempty"`
}

type createDatasetItemRequestRaw struct {
	DatasetName    string          `json:"datasetName"`
	Input          json.RawMessage `json:"input,omitempty"`
	ExpectedOutput json.RawMessage `json:"expectedOutput,omitempty"`
	Status         string          `json:"status,omitempty"`
	ID             *string         `json:"id,omitempty"`
}

func rawMessageToStringPtr(raw json.RawMessage) *string {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	s := string(raw)
	return &s
}

func stringPtrToRawMessage(s *string) json.RawMessage {
	if s == nil {
		return nil
	}
	return json.RawMessage(*s)
}

func datasetItemFromResponse(r *datasetItemResponse) *DatasetItem {
	return &DatasetItem{
		ID:             r.ID,
		DatasetName:    r.DatasetName,
		Status:         r.Status,
		Input:          rawMessageToStringPtr(r.Input),
		ExpectedOutput: rawMessageToStringPtr(r.ExpectedOutput),
	}
}

func (c *Client) GetDatasetItem(ctx context.Context, itemID string) (*DatasetItem, error) {
	path := fmt.Sprintf("/api/public/dataset-items/%s", itemID)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var r datasetItemResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling dataset item response: %w", err)
	}

	return datasetItemFromResponse(&r), nil
}

func (c *Client) CreateDatasetItem(ctx context.Context, req *CreateDatasetItemRequest) (*DatasetItem, error) {
	raw := createDatasetItemRequestRaw{
		DatasetName:    req.DatasetName,
		Input:          stringPtrToRawMessage(req.Input),
		ExpectedOutput: stringPtrToRawMessage(req.ExpectedOutput),
		Status:         req.Status,
		ID:             req.ID,
	}

	body, err := c.do(ctx, http.MethodPost, "/api/public/dataset-items", raw)
	if err != nil {
		return nil, err
	}

	var r datasetItemResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("unmarshaling create dataset item response: %w", err)
	}

	return datasetItemFromResponse(&r), nil
}

func (c *Client) DeleteDatasetItem(ctx context.Context, itemID string) error {
	path := fmt.Sprintf("/api/public/dataset-items/%s", itemID)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
