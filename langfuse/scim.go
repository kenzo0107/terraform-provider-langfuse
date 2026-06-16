package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type SCIMUserName struct {
	Formatted  string `json:"formatted,omitempty"`
	GivenName  string `json:"givenName,omitempty"`
	FamilyName string `json:"familyName,omitempty"`
}

type SCIMEmail struct {
	Value   string `json:"value"`
	Primary bool   `json:"primary,omitempty"`
}

type SCIMUser struct {
	ID         string       `json:"id"`
	ExternalID *string      `json:"externalId,omitempty"`
	UserName   string       `json:"userName"`
	Name       SCIMUserName `json:"name"`
	Emails     []SCIMEmail  `json:"emails"`
	Active     bool         `json:"active"`
}

type CreateSCIMUserRequest struct {
	UserName   string       `json:"userName"`
	Name       SCIMUserName `json:"name"`
	Emails     []SCIMEmail  `json:"emails"`
	Active     *bool        `json:"active,omitempty"`
	ExternalID *string      `json:"externalId,omitempty"`
	Password   *string      `json:"password,omitempty"`
}

func (c *Client) GetSCIMUser(ctx context.Context, userID string) (*SCIMUser, error) {
	path := fmt.Sprintf("/api/public/scim/Users/%s", userID)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var u SCIMUser
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("unmarshaling SCIM user response: %w", err)
	}

	return &u, nil
}

func (c *Client) CreateSCIMUser(ctx context.Context, req *CreateSCIMUserRequest) (*SCIMUser, error) {
	body, err := c.do(ctx, http.MethodPost, "/api/public/scim/Users", req)
	if err != nil {
		return nil, err
	}

	var u SCIMUser
	if err := json.Unmarshal(body, &u); err != nil {
		return nil, fmt.Errorf("unmarshaling create SCIM user response: %w", err)
	}

	return &u, nil
}

func (c *Client) DeleteSCIMUser(ctx context.Context, userID string) error {
	path := fmt.Sprintf("/api/public/scim/Users/%s", userID)
	_, err := c.do(ctx, http.MethodDelete, path, nil)
	return err
}
