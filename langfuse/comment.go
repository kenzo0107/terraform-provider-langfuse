package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type Comment struct {
	ID           string  `json:"id"`
	ProjectID    string  `json:"projectId"`
	ObjectType   string  `json:"objectType"`
	ObjectID     string  `json:"objectId"`
	Content      string  `json:"content"`
	AuthorUserID *string `json:"authorUserId,omitempty"`
}

type createCommentResponse struct {
	ID string `json:"id"`
}

type CreateCommentRequest struct {
	ObjectType   string  `json:"objectType"`
	ObjectID     string  `json:"objectId"`
	Content      string  `json:"content"`
	AuthorUserID *string `json:"authorUserId,omitempty"`
}

func (c *Client) GetComment(ctx context.Context, commentID string) (*Comment, error) {
	path := fmt.Sprintf("/api/public/comments/%s", commentID)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var cm Comment
	if err := json.Unmarshal(body, &cm); err != nil {
		return nil, fmt.Errorf("unmarshaling comment response: %w", err)
	}

	return &cm, nil
}

func (c *Client) CreateComment(ctx context.Context, req *CreateCommentRequest) (string, error) {
	body, err := c.do(ctx, http.MethodPost, "/api/public/comments", req)
	if err != nil {
		return "", err
	}

	var resp createCommentResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return "", fmt.Errorf("unmarshaling create comment response: %w", err)
	}

	return resp.ID, nil
}
