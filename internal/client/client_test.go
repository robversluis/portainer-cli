package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robversluis/portainer-cli/internal/config"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		profile   *config.Profile
		wantError bool
	}{
		{
			name: "valid profile with API key",
			profile: &config.Profile{
				URL:    "https://test.example.com",
				APIKey: "test-key",
			},
			wantError: false,
		},
		{
			name: "valid profile with token",
			profile: &config.Profile{
				URL:   "https://test.example.com",
				Token: "jwt-token",
			},
			wantError: false,
		},
		{
			name: "valid profile with username",
			profile: &config.Profile{
				URL:      "https://test.example.com",
				Username: "admin",
			},
			wantError: false,
		},
		{
			name:      "nil profile",
			profile:   nil,
			wantError: true,
		},
		{
			name: "invalid URL - no scheme",
			profile: &config.Profile{
				URL:    "test.example.com",
				APIKey: "test-key",
			},
			wantError: true,
		},
		{
			name: "missing URL",
			profile: &config.Profile{
				APIKey: "test-key",
			},
			wantError: true,
		},
		{
			name: "missing auth method",
			profile: &config.Profile{
				URL: "https://test.example.com",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.profile)
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

			if client == nil {
				t.Error("client should not be nil")
			}
		})
	}
}

func TestClient_SetToken(t *testing.T) {
	profile := &config.Profile{
		URL:    "https://test.example.com",
		APIKey: "test-key",
	}

	client, err := NewClient(profile)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	token := "new-jwt-token"
	client.SetToken(token)

	if client.GetToken() != token {
		t.Errorf("expected token %s, got %s", token, client.GetToken())
	}
}

func TestClient_buildURL(t *testing.T) {
	profile := &config.Profile{
		URL:    "https://test.example.com",
		APIKey: "test-key",
	}

	client, err := NewClient(profile)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	tests := []struct {
		path     string
		expected string
	}{
		{
			path:     "auth",
			expected: "https://test.example.com/api/auth",
		},
		{
			path:     "/auth",
			expected: "https://test.example.com/api/auth",
		},
		{
			path:     "users/1",
			expected: "https://test.example.com/api/users/1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := client.buildURL(tt.path)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestClient_DoRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-KEY") != "test-key" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"message": "unauthorized"})
			return
		}

		if r.URL.Path == "/api/test" {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"result": "success"})
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

	t.Run("successful request", func(t *testing.T) {
		var result map[string]string
		err := client.Get("test", &result)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if result["result"] != "success" {
			t.Errorf("expected result 'success', got '%s'", result["result"])
		}
	})

	t.Run("not found error", func(t *testing.T) {
		var result map[string]string
		err := client.Get("nonexistent", &result)
		if err == nil {
			t.Error("expected error but got none")
		}

		if !IsNotFoundError(err) {
			t.Errorf("expected not found error, got: %v", err)
		}
	})
}

func TestClient_WithOptions(t *testing.T) {
	profile := &config.Profile{
		URL:    "https://test.example.com",
		APIKey: "test-key",
	}

	t.Run("with verbose", func(t *testing.T) {
		client, err := NewClient(profile, WithVerbose(true))
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		if !client.verbose {
			t.Error("verbose should be true")
		}
	})

	t.Run("with timeout", func(t *testing.T) {
		timeout := 5 * time.Second
		client, err := NewClient(profile, WithTimeout(timeout))
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		if client.httpClient.Timeout != timeout {
			t.Errorf("expected timeout %v, got %v", timeout, client.httpClient.Timeout)
		}
	})

	t.Run("with max retries", func(t *testing.T) {
		retries := 5
		client, err := NewClient(profile, WithMaxRetries(retries))
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		if client.maxRetries != retries {
			t.Errorf("expected max retries %d, got %d", retries, client.maxRetries)
		}
	})

	t.Run("with insecure", func(t *testing.T) {
		client, err := NewClient(profile, WithInsecure(true))
		if err != nil {
			t.Fatalf("failed to create client: %v", err)
		}

		transport := client.httpClient.Transport.(*http.Transport)
		if !transport.TLSClientConfig.InsecureSkipVerify {
			t.Error("InsecureSkipVerify should be true")
		}
	})
}

func TestAPIError(t *testing.T) {
	tests := []struct {
		name     string
		err      *APIError
		expected string
	}{
		{
			name: "with details",
			err: &APIError{
				StatusCode: 400,
				Message:    "Bad Request",
				Details:    "Invalid parameter",
			},
			expected: "API error (HTTP 400): Bad Request - Invalid parameter",
		},
		{
			name: "without details",
			err: &APIError{
				StatusCode: 404,
				Message:    "Not Found",
			},
			expected: "API error (HTTP 404): Not Found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.err.Error()
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestErrorCheckers(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		is404    bool
		is401    bool
		is403    bool
	}{
		{
			name:     "404 error",
			err:      &APIError{StatusCode: 404, Message: "Not Found"},
			is404:    true,
			is401:    false,
			is403:    false,
		},
		{
			name:     "401 error",
			err:      &APIError{StatusCode: 401, Message: "Unauthorized"},
			is404:    false,
			is401:    true,
			is403:    false,
		},
		{
			name:     "403 error",
			err:      &APIError{StatusCode: 403, Message: "Forbidden"},
			is404:    false,
			is401:    false,
			is403:    true,
		},
		{
			name:     "non-API error",
			err:      http.ErrServerClosed,
			is404:    false,
			is401:    false,
			is403:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if IsNotFoundError(tt.err) != tt.is404 {
				t.Errorf("IsNotFoundError: expected %v, got %v", tt.is404, IsNotFoundError(tt.err))
			}
			if IsUnauthorizedError(tt.err) != tt.is401 {
				t.Errorf("IsUnauthorizedError: expected %v, got %v", tt.is401, IsUnauthorizedError(tt.err))
			}
			if IsForbiddenError(tt.err) != tt.is403 {
				t.Errorf("IsForbiddenError: expected %v, got %v", tt.is403, IsForbiddenError(tt.err))
			}
		})
	}
}

func TestClient_HTTPMethods(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]string{
			"method": r.Method,
			"path":   r.URL.Path,
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
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

	tests := []struct {
		name           string
		method         func() error
		expectedMethod string
	}{
		{
			name: "GET",
			method: func() error {
				var result map[string]string
				return client.Get("test", &result)
			},
			expectedMethod: "GET",
		},
		{
			name: "POST",
			method: func() error {
				var result map[string]string
				return client.Post("test", map[string]string{"key": "value"}, &result)
			},
			expectedMethod: "POST",
		},
		{
			name: "PUT",
			method: func() error {
				var result map[string]string
				return client.Put("test", map[string]string{"key": "value"}, &result)
			},
			expectedMethod: "PUT",
		},
		{
			name: "DELETE",
			method: func() error {
				return client.Delete("test")
			},
			expectedMethod: "DELETE",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.method()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
