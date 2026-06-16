package langfuse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type MembershipRole string

const (
	MembershipRoleOwner  MembershipRole = "OWNER"
	MembershipRoleAdmin  MembershipRole = "ADMIN"
	MembershipRoleMember MembershipRole = "MEMBER"
	MembershipRoleViewer MembershipRole = "VIEWER"
)

type Membership struct {
	UserID string         `json:"userId"`
	Role   MembershipRole `json:"role"`
	Email  string         `json:"email"`
	Name   string         `json:"name"`
}

type getMembershipsResponse struct {
	Memberships []*Membership `json:"memberships"`
}

type upsertMembershipRequest struct {
	UserID string         `json:"userId"`
	Role   MembershipRole `json:"role"`
}

type deleteMembershipRequest struct {
	UserID string `json:"userId"`
}

func (c *Client) GetProjectMembership(ctx context.Context, projectID, userID string) (*Membership, error) {
	memberships, err := c.ListProjectMemberships(ctx, projectID)
	if err != nil {
		return nil, err
	}
	for _, m := range memberships {
		if m.UserID == userID {
			return m, nil
		}
	}
	return nil, &APIError{StatusCode: http.StatusNotFound, Body: fmt.Sprintf("membership for user %q in project %q not found", userID, projectID)}
}

func (c *Client) ListProjectMemberships(ctx context.Context, projectID string) ([]*Membership, error) {
	path := fmt.Sprintf("/api/public/projects/%s/memberships", projectID)
	body, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var resp getMembershipsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("unmarshaling memberships response: %w", err)
	}

	return resp.Memberships, nil
}

func (c *Client) UpsertProjectMembership(ctx context.Context, projectID, userID string, role MembershipRole) (*Membership, error) {
	path := fmt.Sprintf("/api/public/projects/%s/memberships", projectID)
	payload := upsertMembershipRequest{UserID: userID, Role: role}
	body, err := c.do(ctx, http.MethodPut, path, payload)
	if err != nil {
		return nil, err
	}

	var m Membership
	if err := json.Unmarshal(body, &m); err != nil {
		return nil, fmt.Errorf("unmarshaling upsert membership response: %w", err)
	}

	return &m, nil
}

func (c *Client) DeleteProjectMembership(ctx context.Context, projectID, userID string) error {
	path := fmt.Sprintf("/api/public/projects/%s/memberships", projectID)
	payload := deleteMembershipRequest{UserID: userID}
	_, err := c.do(ctx, http.MethodDelete, path, payload)
	return err
}
