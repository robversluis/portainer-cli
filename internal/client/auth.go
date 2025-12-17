package client

import (
	"fmt"
	"net/http"

	"github.com/rob/portainer-cli/internal/config"
)

type AuthService struct {
	client *Client
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	JWT string `json:"jwt"`
}

type UserInfo struct {
	ID       int    `json:"Id"`
	Username string `json:"Username"`
	Role     int    `json:"Role"`
}

type StatusResponse struct {
	Version string `json:"Version"`
}

func NewAuthService(client *Client) *AuthService {
	return &AuthService{client: client}
}

func (s *AuthService) Login(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", fmt.Errorf("username and password are required")
	}

	req := LoginRequest{
		Username: username,
		Password: password,
	}

	var resp LoginResponse
	if err := s.client.Post("auth", req, &resp); err != nil {
		if IsUnauthorizedError(err) {
			return "", fmt.Errorf("invalid credentials")
		}
		return "", fmt.Errorf("login failed: %w", err)
	}

	if resp.JWT == "" {
		return "", fmt.Errorf("no token returned from server")
	}

	s.client.SetToken(resp.JWT)

	return resp.JWT, nil
}

func (s *AuthService) Logout() error {
	req, err := s.client.newRequest(http.MethodPost, "auth/logout", nil)
	if err != nil {
		return err
	}

	resp, err := s.client.do(req)
	if err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	defer resp.Body.Close()

	s.client.SetToken("")

	return nil
}

func (s *AuthService) ValidateToken() (*UserInfo, error) {
	var users []UserInfo
	if err := s.client.Get("users", &users); err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	if len(users) == 0 {
		return nil, fmt.Errorf("no user information available")
	}

	return &users[0], nil
}

func (s *AuthService) GetStatus() (*StatusResponse, error) {
	var status StatusResponse
	if err := s.client.Get("status", &status); err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	return &status, nil
}

func LoginAndSaveToken(profile *config.Profile, username, password string) (string, error) {
	client, err := NewClient(profile, WithVerbose(false))
	if err != nil {
		return "", err
	}

	authService := NewAuthService(client)
	token, err := authService.Login(username, password)
	if err != nil {
		return "", err
	}

	cfg, err := config.Load()
	if err != nil {
		return token, fmt.Errorf("logged in but failed to load config: %w", err)
	}

	profileName := cfg.CurrentProfile
	if profileName == "" {
		return token, fmt.Errorf("logged in but no current profile set")
	}

	storedProfile, err := cfg.GetProfile(profileName)
	if err != nil {
		return token, fmt.Errorf("logged in but failed to get profile: %w", err)
	}

	storedProfile.Token = token
	storedProfile.Username = username

	if err := cfg.Save(); err != nil {
		return token, fmt.Errorf("logged in but failed to save token: %w", err)
	}

	return token, nil
}

func ValidateAuthentication(profile *config.Profile) error {
	client, err := NewClient(profile, WithVerbose(false))
	if err != nil {
		return err
	}

	authService := NewAuthService(client)
	_, err = authService.ValidateToken()
	return err
}
