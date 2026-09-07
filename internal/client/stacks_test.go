package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robversluis/portainer-cli/internal/config"
)

// Portainer 2.45 answers 405 on the old POST /stacks?type=..&method=.. route, so the
// create call has to go to /stacks/create/standalone/string with a JSON body.
func TestStackService_Deploy(t *testing.T) {
	var gotPath, gotQuery, gotContentType string
	var gotBody struct {
		Name             string     `json:"name"`
		StackFileContent string     `json:"stackFileContent"`
		Env              []StackEnv `json:"env"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery
		gotContentType = r.Header.Get("Content-Type")

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Stack{Id: 65, Name: gotBody.Name, EndpointId: 8})
	}))
	defer server.Close()

	client, err := NewClient(&config.Profile{URL: server.URL, APIKey: "test-key"})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	env := []StackEnv{{Name: "SERVICE_TAG", Value: "acc"}}
	stack, err := NewStackService(client).Deploy(8, "gilivillas-acc", "services: {}\n", env)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotPath != "/api/stacks/create/standalone/string" {
		t.Errorf("expected the 2.19+ create route, got %q", gotPath)
	}
	if gotQuery != "endpointId=8" {
		t.Errorf("expected endpointId=8, got %q", gotQuery)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected a JSON body, got content-type %q", gotContentType)
	}
	if gotBody.Name != "gilivillas-acc" {
		t.Errorf("expected the stack name in the body, got %q", gotBody.Name)
	}
	if gotBody.StackFileContent != "services: {}\n" {
		t.Errorf("expected the compose content in the body, got %q", gotBody.StackFileContent)
	}
	if len(gotBody.Env) != 1 || gotBody.Env[0].Name != "SERVICE_TAG" || gotBody.Env[0].Value != "acc" {
		t.Errorf("expected the env to be passed through, got %+v", gotBody.Env)
	}
	if stack.Id != 65 {
		t.Errorf("expected the decoded stack, got %+v", stack)
	}
}
