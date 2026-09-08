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

// A stack update has to ask Portainer to prune, or containers for services that were
// removed from the compose file keep running as orphans.
func TestStackService_Update(t *testing.T) {
	var gotMethod, gotPath, gotQuery string
	var gotBody struct {
		StackFileContent       string     `json:"stackFileContent"`
		Env                    []StackEnv `json:"env"`
		RepullImageAndRedeploy bool       `json:"repullImageAndRedeploy"`
		Prune                  bool       `json:"prune"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		gotPath = r.URL.Path
		gotQuery = r.URL.RawQuery

		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(&config.Profile{URL: server.URL, APIKey: "test-key"})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	env := []StackEnv{{Name: "SERVICE_TAG", Value: "prod"}}
	if err := NewStackService(client).Update(66, 8, "services: {}", env, true, true); err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	if gotMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", gotMethod)
	}
	if want := "/api/stacks/66"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if want := "endpointId=8"; gotQuery != want {
		t.Errorf("query = %q, want %q", gotQuery, want)
	}
	if gotBody.StackFileContent != "services: {}" {
		t.Errorf("stackFileContent = %q", gotBody.StackFileContent)
	}
	if !gotBody.RepullImageAndRedeploy {
		t.Error("repullImageAndRedeploy = false, want true")
	}
	if !gotBody.Prune {
		t.Error("prune = false, want true")
	}
	if len(gotBody.Env) != 1 || gotBody.Env[0].Name != "SERVICE_TAG" {
		t.Errorf("env = %+v, want the one variable passed in", gotBody.Env)
	}
}

// Without --prune the flag must not be sent as true: pruning is destructive and only
// happens when it is asked for.
func TestStackService_UpdateWithoutPrune(t *testing.T) {
	var gotBody struct {
		Prune bool `json:"prune"`
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotBody)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, err := NewClient(&config.Profile{URL: server.URL, APIKey: "test-key"})
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	if err := NewStackService(client).Update(66, 8, "services: {}", nil, false, false); err != nil {
		t.Fatalf("Update returned an error: %v", err)
	}

	if gotBody.Prune {
		t.Error("prune = true, want false")
	}
}
