package provider_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	langfuseprovider "github.com/kenzo0107/terraform-provider-langfuse/internal/provider"
)

// mockProject represents a project stored in the mock server.
type mockProject struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// mockLangfuseServer is an in-memory mock of the Langfuse HTTP API.
type mockLangfuseServer struct {
	mu       sync.Mutex
	projects map[string]mockProject
	counter  int
	srv      *httptest.Server
}

func newMockLangfuseServer(t *testing.T) *mockLangfuseServer {
	t.Helper()
	m := &mockLangfuseServer{
		projects: make(map[string]mockProject),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/public/projects/", m.handleProjectByID)
	mux.HandleFunc("/api/public/projects", m.handleProjectsCollection)
	m.srv = httptest.NewServer(mux)
	t.Cleanup(m.srv.Close)
	return m
}

func (m *mockLangfuseServer) URL() string {
	return m.srv.URL
}

func (m *mockLangfuseServer) newID() string {
	m.counter++
	return fmt.Sprintf("proj-%04d", m.counter)
}

func (m *mockLangfuseServer) handleProjectsCollection(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		m.mu.Lock()
		defer m.mu.Unlock()
		projects := make([]mockProject, 0, len(m.projects))
		for _, p := range m.projects {
			projects = append(projects, p)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"data": projects})
	case http.MethodPost:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		p := mockProject{ID: m.newID(), Name: req.Name}
		m.projects[p.ID] = p
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(p)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (m *mockLangfuseServer) handleProjectByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/public/projects/")
	if id == "" {
		http.Error(w, "missing id", http.StatusBadRequest)
		return
	}
	switch r.Method {
	case http.MethodPatch:
		var req struct {
			Name string `json:"name"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		p, ok := m.projects[id]
		if !ok {
			http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
			return
		}
		p.Name = req.Name
		m.projects[id] = p
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(p)
	case http.MethodDelete:
		m.mu.Lock()
		defer m.mu.Unlock()
		if _, ok := m.projects[id]; !ok {
			http.Error(w, `{"message":"not found"}`, http.StatusNotFound)
			return
		}
		delete(m.projects, id)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func protoV6ProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"langfuse": providerserver.NewProtocol6WithError(langfuseprovider.New("test")()),
	}
}

func providerConfig(serverURL string) string {
	return fmt.Sprintf(`
provider "langfuse" {
  public_key = "test-public-key"
  secret_key = "test-secret-key"
  host       = %q
}
`, serverURL)
}

func TestProvider_MissingPublicKey(t *testing.T) {
	t.Setenv("LANGFUSE_PUBLIC_KEY", "")
	t.Setenv("LANGFUSE_SECRET_KEY", "")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "langfuse" {
  secret_key = "sec"
}
data "langfuse_project" "test" {
  id = "some-id"
}
`,
				ExpectError: regexp.MustCompile(`Missing Langfuse Public Key`),
			},
		},
	})
}

func TestProvider_MissingSecretKey(t *testing.T) {
	t.Setenv("LANGFUSE_PUBLIC_KEY", "")
	t.Setenv("LANGFUSE_SECRET_KEY", "")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: protoV6ProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: `
provider "langfuse" {
  public_key = "pub"
}
data "langfuse_project" "test" {
  id = "some-id"
}
`,
				ExpectError: regexp.MustCompile(`Missing Langfuse Secret Key`),
			},
		},
	})
}
