package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robversluis/portainer-cli/internal/config"
)

func TestEnvironmentService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/endpoints" {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		environments := []Environment{
			{
				Id:     1,
				Name:   "local",
				Type:   EnvironmentTypeDockerLocal,
				URL:    "unix:///var/run/docker.sock",
				Status: EnvironmentStatusUp,
			},
			{
				Id:     2,
				Name:   "production",
				Type:   EnvironmentTypeAgentOnKubernetes,
				URL:    "https://k8s.prod.com",
				Status: EnvironmentStatusUp,
			},
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(environments)
	}))
	defer server.Close()

	profile := &config.Profile{
		URL:    server.URL,
		APIKey: "test-key",
	}

	client, err := NewClient(profile)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	envService := NewEnvironmentService(client)
	environments, err := envService.List()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(environments) != 2 {
		t.Errorf("expected 2 environments, got %d", len(environments))
	}

	if environments[0].Name != "local" {
		t.Errorf("expected name 'local', got '%s'", environments[0].Name)
	}

	if environments[1].Type != EnvironmentTypeAgentOnKubernetes {
		t.Errorf("expected type %d, got %d", EnvironmentTypeAgentOnKubernetes, environments[1].Type)
	}
}

func TestEnvironmentService_Get(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/endpoints/1" {
			env := Environment{
				Id:     1,
				Name:   "local",
				Type:   EnvironmentTypeDockerLocal,
				URL:    "unix:///var/run/docker.sock",
				Status: EnvironmentStatusUp,
				GroupId: 1,
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(env)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	profile := &config.Profile{
		URL:    server.URL,
		APIKey: "test-key",
	}

	client, err := NewClient(profile)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	envService := NewEnvironmentService(client)

	t.Run("get existing environment", func(t *testing.T) {
		env, err := envService.Get(1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if env.Id != 1 {
			t.Errorf("expected ID 1, got %d", env.Id)
		}

		if env.Name != "local" {
			t.Errorf("expected name 'local', got '%s'", env.Name)
		}
	})

	t.Run("get non-existent environment", func(t *testing.T) {
		_, err := envService.Get(999)
		if err == nil {
			t.Error("expected error but got none")
		}
	})
}

func TestEnvironmentService_GetByName(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/endpoints" {
			environments := []Environment{
				{Id: 1, Name: "local", Type: EnvironmentTypeDockerLocal, URL: "unix:///var/run/docker.sock", Status: EnvironmentStatusUp},
				{Id: 2, Name: "production", Type: EnvironmentTypeAgentOnKubernetes, URL: "https://k8s.prod.com", Status: EnvironmentStatusUp},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(environments)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	profile := &config.Profile{
		URL:    server.URL,
		APIKey: "test-key",
	}

	client, err := NewClient(profile)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	envService := NewEnvironmentService(client)

	t.Run("get by existing name", func(t *testing.T) {
		env, err := envService.GetByName("production")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if env.Name != "production" {
			t.Errorf("expected name 'production', got '%s'", env.Name)
		}

		if env.Id != 2 {
			t.Errorf("expected ID 2, got %d", env.Id)
		}
	})

	t.Run("get by non-existent name", func(t *testing.T) {
		_, err := envService.GetByName("nonexistent")
		if err == nil {
			t.Error("expected error but got none")
		}
	})
}

func TestEnvironmentService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		if r.URL.Path == "/api/endpoints/1" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]string{"message": "not found"})
	}))
	defer server.Close()

	profile := &config.Profile{
		URL:    server.URL,
		APIKey: "test-key",
	}

	client, err := NewClient(profile)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	envService := NewEnvironmentService(client)

	t.Run("delete existing environment", func(t *testing.T) {
		err := envService.Delete(1)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("delete non-existent environment", func(t *testing.T) {
		err := envService.Delete(999)
		if err == nil {
			t.Error("expected error but got none")
		}
	})
}

func TestEnvironment_TypeString(t *testing.T) {
	tests := []struct {
		envType  int
		expected string
	}{
		{EnvironmentTypeDockerLocal, "Docker (Local)"},
		{EnvironmentTypeAgentOnDocker, "Docker (Agent)"},
		{EnvironmentTypeAzure, "Azure"},
		{EnvironmentTypeAgentOnKubernetes, "Kubernetes (Agent)"},
		{EnvironmentTypeEdgeAgentOnDocker, "Docker (Edge)"},
		{EnvironmentTypeEdgeAgentOnKubernetes, "Kubernetes (Edge)"},
		{EnvironmentTypeKubeLocal, "Kubernetes (Local)"},
		{999, "Unknown (999)"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			env := Environment{Type: tt.envType}
			result := env.TypeString()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestEnvironment_StatusString(t *testing.T) {
	tests := []struct {
		status   int
		expected string
	}{
		{EnvironmentStatusUp, "Up"},
		{EnvironmentStatusDown, "Down"},
		{999, "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			env := Environment{Status: tt.status}
			result := env.StatusString()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestEnvironment_GetLatestSnapshot(t *testing.T) {
	t.Run("no snapshots", func(t *testing.T) {
		env := Environment{Snapshots: []Snapshot{}}
		snapshot := env.GetLatestSnapshot()
		if snapshot != nil {
			t.Error("expected nil snapshot")
		}
	})

	t.Run("single snapshot", func(t *testing.T) {
		env := Environment{
			Snapshots: []Snapshot{
				{Time: 1000, RunningContainerCount: 5},
			},
		}
		snapshot := env.GetLatestSnapshot()
		if snapshot == nil {
			t.Fatal("expected snapshot but got nil")
		}
		if snapshot.RunningContainerCount != 5 {
			t.Errorf("expected 5 running containers, got %d", snapshot.RunningContainerCount)
		}
	})

	t.Run("multiple snapshots", func(t *testing.T) {
		env := Environment{
			Snapshots: []Snapshot{
				{Time: 1000, RunningContainerCount: 5},
				{Time: 3000, RunningContainerCount: 10},
				{Time: 2000, RunningContainerCount: 7},
			},
		}
		snapshot := env.GetLatestSnapshot()
		if snapshot == nil {
			t.Fatal("expected snapshot but got nil")
		}
		if snapshot.Time != 3000 {
			t.Errorf("expected time 3000, got %d", snapshot.Time)
		}
		if snapshot.RunningContainerCount != 10 {
			t.Errorf("expected 10 running containers, got %d", snapshot.RunningContainerCount)
		}
	})
}
