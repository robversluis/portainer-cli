package client

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/robversluis/portainer-cli/internal/config"
)

const (
	defaultTimeout    = 300 * time.Second
	defaultMaxRetries = 3
	defaultRetryDelay = 2 * time.Second
	userAgent         = "portainer-cli"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
	token      string
	username   string
	verbose    bool
	maxRetries int
	retryDelay time.Duration
}

type ClientOption func(*Client)

func WithVerbose(verbose bool) ClientOption {
	return func(c *Client) {
		c.verbose = verbose
	}
}

func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = timeout
	}
}

func WithMaxRetries(retries int) ClientOption {
	return func(c *Client) {
		c.maxRetries = retries
	}
}

func WithInsecure(insecure bool) ClientOption {
	return func(c *Client) {
		if insecure {
			transport, ok := c.httpClient.Transport.(*http.Transport)
			if !ok {
				return
			}
			transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		}
	}
}

func WithCustomCA(certPool *tls.Config) ClientOption {
	return func(c *Client) {
		transport, ok := c.httpClient.Transport.(*http.Transport)
		if !ok {
			return
		}
		transport.TLSClientConfig = certPool
	}
}

func NewClient(profile *config.Profile, opts ...ClientOption) (*Client, error) {
	if profile == nil {
		return nil, fmt.Errorf("profile cannot be nil")
	}

	if err := profile.Validate(); err != nil {
		return nil, fmt.Errorf("invalid profile: %w", err)
	}

	baseURL := strings.TrimSuffix(profile.URL, "/")
	if !strings.HasPrefix(baseURL, "http://") && !strings.HasPrefix(baseURL, "https://") {
		return nil, fmt.Errorf("invalid URL: must start with http:// or https://")
	}

	client := &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					MinVersion: tls.VersionTLS12,
				},
			},
		},
		apiKey:     profile.APIKey,
		token:      profile.Token,
		username:   profile.Username,
		maxRetries: defaultMaxRetries,
		retryDelay: defaultRetryDelay,
	}

	if profile.Insecure {
		opts = append(opts, WithInsecure(true))
	}

	for _, opt := range opts {
		if opt != nil {
			opt(client)
		}
	}

	return client, nil
}

func (c *Client) SetToken(token string) {
	c.token = token
}

func (c *Client) GetToken() string {
	return c.token
}

func (c *Client) buildURL(path string) string {
	path = strings.TrimPrefix(path, "/")
	return fmt.Sprintf("%s/api/%s", c.baseURL, path)
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	url := c.buildURL(path)

	var bodyReader io.Reader
	var jsonData []byte
	if body != nil {
		var err error
		jsonData, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.GetBody = func() (io.ReadCloser, error) {
			return io.NopCloser(bytes.NewReader(jsonData)), nil
		}
	}

	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if c.apiKey != "" {
		req.Header.Set("X-API-KEY", c.apiKey)
	} else if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	return req, nil
}

func (c *Client) do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; attempt <= c.maxRetries; attempt++ {
		if attempt > 0 {
			if c.verbose {
				fmt.Printf("Retry attempt %d/%d after %v\n", attempt, c.maxRetries, c.retryDelay)
			}
			time.Sleep(c.retryDelay)

			// Reset request body for retry
			if req.GetBody != nil {
				req.Body, err = req.GetBody()
				if err != nil {
					return nil, fmt.Errorf("failed to reset request body: %w", err)
				}
			}
		}

		if c.verbose {
			fmt.Printf("%s %s\n", req.Method, req.URL.String())
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			if attempt < c.maxRetries && isRetryableError(err) {
				continue
			}
			return nil, fmt.Errorf("request failed: %w", err)
		}

		if resp.StatusCode >= 500 && attempt < c.maxRetries {
			resp.Body.Close()
			continue
		}

		break
	}

	if c.verbose && resp != nil {
		fmt.Printf("Response: %d %s\n", resp.StatusCode, resp.Status)
	}

	return resp, nil
}

func (c *Client) DoRequest(method, path string, body interface{}, result interface{}) error {
	req, err := c.newRequest(method, path, body)
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return err
	}

	if result != nil && resp.StatusCode != http.StatusNoContent {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

func (c *Client) Get(path string, result interface{}) error {
	return c.DoRequest(http.MethodGet, path, nil, result)
}

func (c *Client) Post(path string, body interface{}, result interface{}) error {
	return c.DoRequest(http.MethodPost, path, body, result)
}

func (c *Client) Put(path string, body interface{}, result interface{}) error {
	return c.DoRequest(http.MethodPut, path, body, result)
}

func (c *Client) Delete(path string) error {
	return c.DoRequest(http.MethodDelete, path, nil, nil)
}

func checkResponse(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    fmt.Sprintf("HTTP %d: failed to read response body", resp.StatusCode),
		}
	}
	bodyString := string(bodyBytes)

	var apiError APIError
	if err := json.Unmarshal(bodyBytes, &apiError); err == nil && apiError.Message != "" {
		apiError.StatusCode = resp.StatusCode
		return &apiError
	}

	return &APIError{
		StatusCode: resp.StatusCode,
		Message:    fmt.Sprintf("HTTP %d: %s", resp.StatusCode, bodyString),
	}
}

func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	if urlErr, ok := err.(*url.Error); ok {
		if urlErr.Timeout() || urlErr.Temporary() {
			return true
		}
	}

	errStr := err.Error()
	return strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "timeout")
}

type APIError struct {
	StatusCode int    `json:"-"`
	Message    string `json:"message"`
	Details    string `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("API error (HTTP %d): %s - %s", e.StatusCode, e.Message, e.Details)
	}
	return fmt.Sprintf("API error (HTTP %d): %s", e.StatusCode, e.Message)
}

func IsNotFoundError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}

func IsUnauthorizedError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

func IsForbiddenError(err error) bool {
	if apiErr, ok := err.(*APIError); ok {
		return apiErr.StatusCode == http.StatusForbidden
	}
	return false
}
