package langfuse

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func newTestServer(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New("pub", "sec", WithHost(srv.URL))
}

func TestListProjects(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/public/projects" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listProjectsResponse{
			Data: []Project{
				{ID: "id-1", Name: "project-1"},
				{ID: "id-2", Name: "project-2"},
			},
		})
	}))

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
	if projects[0].ID != "id-1" || projects[0].Name != "project-1" {
		t.Errorf("unexpected first project: %+v", projects[0])
	}
	if projects[1].ID != "id-2" || projects[1].Name != "project-2" {
		t.Errorf("unexpected second project: %+v", projects[1])
	}
}

func TestListProjects_Empty(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listProjectsResponse{Data: []Project{}})
	}))

	projects, err := client.ListProjects(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(projects) != 0 {
		t.Errorf("expected empty list, got %d projects", len(projects))
	}
}

func TestListProjects_APIError(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("internal server error"))
	}))

	_, err := client.ListProjects(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", apiErr.StatusCode)
		}
	}
}

func TestGetProject_Found(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listProjectsResponse{
			Data: []Project{
				{ID: "id-1", Name: "project-1"},
				{ID: "id-2", Name: "project-2"},
			},
		})
	}))

	project, err := client.GetProject(context.Background(), "id-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != "id-2" || project.Name != "project-2" {
		t.Errorf("unexpected project: %+v", project)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(listProjectsResponse{
			Data: []Project{{ID: "id-1", Name: "project-1"}},
		})
	}))

	_, err := client.GetProject(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", apiErr.StatusCode)
	}
}

func TestCreateProject(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/public/projects" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req createProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if req.Name != "new-project" {
			t.Errorf("expected name %q, got %q", "new-project", req.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(Project{ID: "new-id", Name: "new-project"})
	}))

	project, err := client.CreateProject(context.Background(), "new-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != "new-id" || project.Name != "new-project" {
		t.Errorf("unexpected project: %+v", project)
	}
}

func TestCreateProject_APIError(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"message":"invalid name"}`))
	}))

	_, err := client.CreateProject(context.Background(), "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateProject(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/public/projects/id-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		var req updateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		if req.Name != "renamed" {
			t.Errorf("expected name %q, got %q", "renamed", req.Name)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Project{ID: "id-1", Name: "renamed"})
	}))

	project, err := client.UpdateProject(context.Background(), "id-1", "renamed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if project.ID != "id-1" || project.Name != "renamed" {
		t.Errorf("unexpected project: %+v", project)
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))

	_, err := client.UpdateProject(context.Background(), "nonexistent", "name")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404, got %d", apiErr.StatusCode)
		}
	}
}

func TestDeleteProject(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/api/public/projects/id-1" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{}`))
	}))

	err := client.DeleteProject(context.Background(), "id-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	client := newTestServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message":"not found"}`))
	}))

	err := client.DeleteProject(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if apiErr, ok := err.(*APIError); ok {
		if apiErr.StatusCode != http.StatusNotFound {
			t.Errorf("expected 404, got %d", apiErr.StatusCode)
		}
	}
}
