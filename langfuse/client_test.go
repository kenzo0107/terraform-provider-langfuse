package langfuse

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	c := New("pub", "sec")
	if c.publicKey != "pub" {
		t.Errorf("expected publicKey %q, got %q", "pub", c.publicKey)
	}
	if c.secretKey != "sec" {
		t.Errorf("expected secretKey %q, got %q", "sec", c.secretKey)
	}
	if c.host != DefaultHost {
		t.Errorf("expected host %q, got %q", DefaultHost, c.host)
	}
}

func TestWithHost(t *testing.T) {
	c := New("pub", "sec", WithHost("https://custom.example.com"))
	if c.host != "https://custom.example.com" {
		t.Errorf("expected host %q, got %q", "https://custom.example.com", c.host)
	}
}

func TestDo_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/api/test" {
			t.Errorf("expected path /api/test, got %s", r.URL.Path)
		}
		user, pass, ok := r.BasicAuth()
		if !ok || user != "pub" || pass != "sec" {
			t.Errorf("unexpected basic auth: user=%s, pass=%s, ok=%v", user, pass, ok)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"key":"value"}`))
	}))
	defer srv.Close()

	c := New("pub", "sec", WithHost(srv.URL))
	body, err := c.do(context.Background(), http.MethodGet, "/api/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(body) != `{"key":"value"}` {
		t.Errorf("unexpected body: %s", string(body))
	}
}

func TestDo_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Unauthorized"}`))
	}))
	defer srv.Close()

	c := New("bad", "key", WithHost(srv.URL))
	_, err := c.do(context.Background(), http.MethodGet, "/api/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", apiErr.StatusCode)
	}
}

func TestDo_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"message":"Not Found"}`))
	}))
	defer srv.Close()

	c := New("pub", "sec", WithHost(srv.URL))
	_, err := c.do(context.Background(), http.MethodGet, "/api/test", nil)
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", apiErr.StatusCode)
	}
}

func TestDo_WithBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	c := New("pub", "sec", WithHost(srv.URL))
	_, err := c.do(context.Background(), http.MethodPost, "/api/test", map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAPIError_Error(t *testing.T) {
	err := &APIError{StatusCode: 404, Body: "not found"}
	expected := "langfuse API error (status 404): not found"
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
