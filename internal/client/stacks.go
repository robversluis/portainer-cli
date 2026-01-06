package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

type StackService struct {
	client *Client
}

type Stack struct {
	Id              int              `json:"Id"`
	Name            string           `json:"Name"`
	Type            int              `json:"Type"`
	EndpointId      int              `json:"EndpointId"`
	SwarmId         string           `json:"SwarmId,omitempty"`
	EntryPoint      string           `json:"EntryPoint"`
	Env             []StackEnv       `json:"Env"`
	ResourceControl *ResourceControl `json:"ResourceControl,omitempty"`
	Status          int              `json:"Status,omitempty"`
	ProjectPath     string           `json:"ProjectPath,omitempty"`
	CreationDate    int64            `json:"CreationDate,omitempty"`
	CreatedBy       string           `json:"CreatedBy,omitempty"`
	UpdateDate      int64            `json:"UpdateDate,omitempty"`
	UpdatedBy       string           `json:"UpdatedBy,omitempty"`
	AdditionalFiles []string         `json:"AdditionalFiles,omitempty"`
	AutoUpdate      *StackAutoUpdate `json:"AutoUpdate,omitempty"`
	GitConfig       *StackGitConfig  `json:"GitConfig,omitempty"`
	FromAppTemplate bool             `json:"FromAppTemplate,omitempty"`
	Namespace       string           `json:"Namespace,omitempty"`
	IsComposeFormat bool             `json:"IsComposeFormat,omitempty"`
}

type StackEnv struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ResourceControl struct {
	Id             int      `json:"Id"`
	ResourceId     string   `json:"ResourceId"`
	SubResourceIds []string `json:"SubResourceIds"`
	Type           int      `json:"Type"`
	UserAccesses   []int    `json:"UserAccesses"`
	TeamAccesses   []int    `json:"TeamAccesses"`
	Public         bool     `json:"Public"`
}

type StackAutoUpdate struct {
	Interval string `json:"Interval,omitempty"`
	Webhook  string `json:"Webhook,omitempty"`
}

type StackGitConfig struct {
	URL            string             `json:"URL"`
	ReferenceName  string             `json:"ReferenceName"`
	ConfigFilePath string             `json:"ConfigFilePath"`
	Authentication *GitAuthentication `json:"Authentication,omitempty"`
}

type GitAuthentication struct {
	Username string `json:"Username,omitempty"`
	Password string `json:"Password,omitempty"`
}

type StackDeployRequest struct {
	Name             string     `json:"Name"`
	SwarmID          string     `json:"SwarmID,omitempty"`
	StackFileContent string     `json:"StackFileContent,omitempty"`
	Env              []StackEnv `json:"Env,omitempty"`
	FromAppTemplate  bool       `json:"FromAppTemplate,omitempty"`
}

const (
	StackTypeSwarm      = 1
	StackTypeCompose    = 2
	StackTypeKubernetes = 3
)

const (
	StackStatusActive   = 1
	StackStatusInactive = 2
)

func NewStackService(client *Client) *StackService {
	return &StackService{client: client}
}

func (s *StackService) List(endpointID int) ([]Stack, error) {
	path := fmt.Sprintf("stacks?filters={\"EndpointId\":%d}", endpointID)

	var stacks []Stack
	if err := s.client.Get(path, &stacks); err != nil {
		return nil, fmt.Errorf("failed to list stacks: %w", err)
	}
	return stacks, nil
}

func (s *StackService) Get(id int) (*Stack, error) {
	path := fmt.Sprintf("stacks/%d", id)

	var stack Stack
	if err := s.client.Get(path, &stack); err != nil {
		return nil, fmt.Errorf("failed to get stack: %w", err)
	}
	return &stack, nil
}

func (s *StackService) GetByName(endpointID int, name string) (*Stack, error) {
	stacks, err := s.List(endpointID)
	if err != nil {
		return nil, err
	}

	for _, stack := range stacks {
		if stack.Name == name {
			return &stack, nil
		}
	}

	return nil, fmt.Errorf("stack '%s' not found", name)
}

func (s *StackService) DeployFromFile(endpointID int, name, filePath string, env []StackEnv) (*Stack, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read stack file: %w", err)
	}

	return s.Deploy(endpointID, name, string(content), env)
}

func (s *StackService) Deploy(endpointID int, name, stackFileContent string, env []StackEnv) (*Stack, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	if err := writer.WriteField("Name", name); err != nil {
		return nil, fmt.Errorf("failed to write name field: %w", err)
	}

	if err := writer.WriteField("StackFileContent", stackFileContent); err != nil {
		return nil, fmt.Errorf("failed to write stack file content: %w", err)
	}

	if len(env) > 0 {
		envJSON, err := json.Marshal(env)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal env: %w", err)
		}
		if err := writer.WriteField("Env", string(envJSON)); err != nil {
			return nil, fmt.Errorf("failed to write env field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	path := fmt.Sprintf("stacks?type=2&method=string&endpointId=%d", endpointID)

	req, err := s.client.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Body = io.NopCloser(body)

	resp, err := s.client.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to deploy stack: %w", err)
	}
	defer resp.Body.Close()

	if err := checkResponse(resp); err != nil {
		return nil, err
	}

	var stack Stack
	if err := json.NewDecoder(resp.Body).Decode(&stack); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &stack, nil
}

func (s *StackService) Update(stackID, endpointID int, stackFileContent string, env []StackEnv) error {
	type updatePayload struct {
		StackFileContent string     `json:"stackFileContent"`
		Env              []StackEnv `json:"env,omitempty"`
	}

	payload := updatePayload{
		StackFileContent: stackFileContent,
	}

	if len(env) > 0 {
		payload.Env = env
	}

	path := fmt.Sprintf("stacks/%d?endpointId=%d", stackID, endpointID)

	return s.client.DoRequest(http.MethodPut, path, payload, nil)
}

func (s *StackService) Remove(stackID, endpointID int) error {
	path := fmt.Sprintf("stacks/%d?endpointId=%d", stackID, endpointID)

	if err := s.client.Delete(path); err != nil {
		return fmt.Errorf("failed to remove stack: %w", err)
	}
	return nil
}

func (s *StackService) GetFile(stackID int) (string, error) {
	path := fmt.Sprintf("stacks/%d/file", stackID)

	var response struct {
		StackFileContent string `json:"StackFileContent"`
	}

	if err := s.client.Get(path, &response); err != nil {
		return "", fmt.Errorf("failed to get stack file: %w", err)
	}

	return response.StackFileContent, nil
}

func (stack *Stack) TypeString() string {
	switch stack.Type {
	case StackTypeSwarm:
		return "Swarm"
	case StackTypeCompose:
		return "Compose"
	case StackTypeKubernetes:
		return "Kubernetes"
	default:
		return fmt.Sprintf("Unknown (%d)", stack.Type)
	}
}

func (stack *Stack) StatusString() string {
	switch stack.Status {
	case StackStatusActive:
		return "Active"
	case StackStatusInactive:
		return "Inactive"
	default:
		return "Unknown"
	}
}

func ParseStackFile(filePath string) (string, error) {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("stack file does not exist: %s", filePath)
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read stack file: %w", err)
	}

	ext := filepath.Ext(filePath)
	if ext != ".yml" && ext != ".yaml" {
		return "", fmt.Errorf("stack file must be a YAML file (.yml or .yaml)")
	}

	return string(content), nil
}
