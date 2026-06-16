package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type BlobStorageIntegration struct {
	ID           string  `json:"id"`
	Type         string  `json:"type"`
	BucketName   string  `json:"bucketName"`
	Prefix       *string `json:"prefix,omitempty"`
	Region       *string `json:"region,omitempty"`
	Endpoint     *string `json:"endpoint,omitempty"`
	ExportPrefix *string `json:"exportPrefix,omitempty"`
	Enabled      bool    `json:"enabled"`
}

type UpsertBlobStorageIntegrationRequest struct {
	Type            string  `json:"type"`
	BucketName      string  `json:"bucketName"`
	Prefix          *string `json:"prefix,omitempty"`
	Region          *string `json:"region,omitempty"`
	Endpoint        *string `json:"endpoint,omitempty"`
	ExportPrefix    *string `json:"exportPrefix,omitempty"`
	AccessKeyID     *string `json:"accessKeyId,omitempty"`
	SecretAccessKey *string `json:"secretAccessKey,omitempty"`
	Enabled         *bool   `json:"enabled,omitempty"`
}

func (c *Client) GetBlobStorageIntegration(ctx context.Context, id string) (*BlobStorageIntegration, error) {
	path := fmt.Sprintf("/api/public/integrations/blob-storage/%s", id)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var b BlobStorageIntegration
	if err := json.Unmarshal(body, &b); err != nil {
		return nil, fmt.Errorf("unmarshaling blob storage integration response: %w", err)
	}

	return &b, nil
}

func (c *Client) UpsertBlobStorageIntegration(ctx context.Context, req *UpsertBlobStorageIntegrationRequest) (*BlobStorageIntegration, error) {
	body, err := c.do(ctx, http.MethodPut, "/api/public/integrations/blob-storage", req)
	if err != nil {
		return nil, err
	}

	var b BlobStorageIntegration
	if err := json.Unmarshal(body, &b); err != nil {
		return nil, fmt.Errorf("unmarshaling upsert blob storage integration response: %w", err)
	}

	return &b, nil
}

func (c *Client) DeleteBlobStorageIntegration(ctx context.Context, id string) error {
	path := fmt.Sprintf("/api/public/integrations/blob-storage/%s", id)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
