package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robversluis/portainer-cli/internal/config"
)

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name       string
		username   string
		password   string
		serverFunc func(w http.ResponseWriter, r *http.Request)
		wantError  bool
		wantToken  bool
	}{
		{
			name:     "successful login",
			username: "admin",
			password: "password",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/auth" {
					w.WriteHeader(http.StatusNotFound)
					return
				}

				var req LoginRequest
				json.NewDecoder(r.Body).Decode(&req)

				if req.Username == "admin" && req.Password == "password" {
					w.WriteHeader(http.StatusOK)
					json.NewEncoder(w).Encode(LoginResponse{JWT: "test-jwt-token"})
					return
				}

				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
			},
			wantError: false,
			wantToken: true,
		},
		{
			name:     "invalid credentials",
			username: "admin",
			password: "wrong",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
			},
			wantError: true,
			wantToken: false,
		},
		{
			name:     "empty username",
			username: "",
			password: "password",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantError: true,
			wantToken: false,
		},
		{
			name:     "empty password",
			username: "admin",
			password: "",
			serverFunc: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			wantError: true,
			wantToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverFunc))
			defer server.Close()

			profile := &config.Profile{
				URL:      server.URL,
				Username: "admin",
			}

			client, err := NewClient(profile)
			if err != nil {
				t.Fatalf("failed to create client: %v", err)
			}

			authService := NewAuthService(client)
			token, err := authService.Login(tt.username, tt.password)

			if tt.wantError {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if tt.wantToken && token == "" {
				t.Error("expected token but got empty string")
			}

			if tt.wantToken && client.GetToken() != token {
				t.Error("token not set in client")
			}
		})
	}
}

func TestAuthService_GetStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/status" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(StatusResponse{Version: "2.19.0"})
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

	authService := NewAuthService(client)
	status, err := authService.GetStatus()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status.Version != "2.19.0" {
		t.Errorf("expected version '2.19.0', got '%s'", status.Version)
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/users" {
			if r.Header.Get("Authorization") != "Bearer valid-token" {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
				return
			}

			users := []UserInfo{
				{ID: 1, Username: "admin", Role: 1},
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(users)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	t.Run("valid token", func(t *testing.T) {
		profile := &config.Profile{
			URL:   server.URL,
			Token: "valid-token",
		}

		client, err := NewClient(profile)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		authService := NewAuthService(client)
		userInfo, err := authService.ValidateToken()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if userInfo.Username != "admin" {
			t.Errorf("expected username 'admin', got '%s'", userInfo.Username)
		}

		if userInfo.ID != 1 {
			t.Errorf("expected ID 1, got %d", userInfo.ID)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		profile := &config.Profile{
			URL:   server.URL,
			Token: "invalid-token",
		}

		client, err := NewClient(profile)
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		authService := NewAuthService(client)
		_, err = authService.ValidateToken()
		if err == nil {
			t.Error("expected error but got none")
		}
	})
}

func TestAuthService_Logout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/auth/logout" && r.Method == "POST" {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	profile := &config.Profile{
		URL:   server.URL,
		Token: "test-token",
	}

	client, err := NewClient(profile)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	authService := NewAuthService(client)

	if client.GetToken() != "test-token" {
		t.Error("token should be set initially")
	}

	err = authService.Logout()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if client.GetToken() != "" {
		t.Error("token should be cleared after logout")
	}
}
