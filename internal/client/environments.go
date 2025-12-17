package client

import (
	"encoding/json"
	"fmt"
)

type EnvironmentService struct {
	client *Client
}

type Environment struct {
	Id                  int              `json:"Id"`
	Name                string           `json:"Name"`
	Type                int              `json:"Type"`
	URL                 string           `json:"URL"`
	PublicURL           string           `json:"PublicURL,omitempty"`
	GroupId             int              `json:"GroupId"`
	Status              int              `json:"Status"`
	Snapshots           []Snapshot       `json:"Snapshots,omitempty"`
	UserAccessPolicies  map[string]int   `json:"UserAccessPolicies,omitempty"`
	TeamAccessPolicies  map[string]int   `json:"TeamAccessPolicies,omitempty"`
	EdgeID              string           `json:"EdgeID,omitempty"`
	EdgeKey             string           `json:"EdgeKey,omitempty"`
	EdgeCheckinInterval int              `json:"EdgeCheckinInterval,omitempty"`
	TagIds              []int            `json:"TagIds,omitempty"`
	TLSConfig           TLSConfiguration `json:"TLSConfig,omitempty"`
	AzureCredentials    AzureCredentials `json:"AzureCredentials,omitempty"`
	KubernetesConfig    KubernetesConfig `json:"Kubernetes,omitempty"`
	Agent               AgentInfo        `json:"Agent,omitempty"`
	SecuritySettings    SecuritySettings `json:"SecuritySettings,omitempty"`
}

type Snapshot struct {
	Time                    int64           `json:"Time"`
	DockerSnapshotRaw       json.RawMessage `json:"DockerSnapshotRaw,omitempty"`
	KubernetesSnapshot      json.RawMessage `json:"KubernetesSnapshot,omitempty"`
	Swarm                   bool            `json:"Swarm"`
	TotalCPU                int             `json:"TotalCPU"`
	TotalMemory             int64           `json:"TotalMemory"`
	RunningContainerCount   int             `json:"RunningContainerCount"`
	StoppedContainerCount   int             `json:"StoppedContainerCount"`
	HealthyContainerCount   int             `json:"HealthyContainerCount"`
	UnhealthyContainerCount int             `json:"UnhealthyContainerCount"`
	VolumeCount             int             `json:"VolumeCount"`
	ImageCount              int             `json:"ImageCount"`
	ServiceCount            int             `json:"ServiceCount"`
	StackCount              int             `json:"StackCount"`
}

type TLSConfiguration struct {
	TLS           bool   `json:"TLS"`
	TLSSkipVerify bool   `json:"TLSSkipVerify"`
	TLSCACertPath string `json:"TLSCACert,omitempty"`
	TLSCertPath   string `json:"TLSCert,omitempty"`
	TLSKeyPath    string `json:"TLSKey,omitempty"`
}

type AzureCredentials struct {
	ApplicationID     string `json:"ApplicationID"`
	TenantID          string `json:"TenantID"`
	AuthenticationKey string `json:"AuthenticationKey"`
}

type KubernetesConfig struct {
	Configuration KubeConfiguration `json:"Configuration,omitempty"`
}

type KubeConfiguration struct {
	UseLoadBalancer  bool     `json:"UseLoadBalancer"`
	UseServerMetrics bool     `json:"UseServerMetrics"`
	StorageClasses   []string `json:"StorageClasses,omitempty"`
	IngressClasses   []string `json:"IngressClasses,omitempty"`
}

type AgentInfo struct {
	Version string `json:"Version,omitempty"`
}

type SecuritySettings struct {
	AllowVolumeBrowserForRegularUsers bool `json:"allowVolumeBrowserForRegularUsers"`
	EnableHostManagementFeatures      bool `json:"enableHostManagementFeatures"`
}

const (
	EnvironmentTypeDockerLocal           = 1
	EnvironmentTypeAgentOnDocker         = 2
	EnvironmentTypeAzure                 = 3
	EnvironmentTypeAgentOnKubernetes     = 4
	EnvironmentTypeEdgeAgentOnDocker     = 5
	EnvironmentTypeEdgeAgentOnKubernetes = 6
	EnvironmentTypeKubeLocal             = 7
)

const (
	EnvironmentStatusUp   = 1
	EnvironmentStatusDown = 2
)

func NewEnvironmentService(client *Client) *EnvironmentService {
	return &EnvironmentService{client: client}
}

func (s *EnvironmentService) List() ([]Environment, error) {
	var environments []Environment
	if err := s.client.Get("endpoints", &environments); err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}
	return environments, nil
}

func (s *EnvironmentService) Get(id int) (*Environment, error) {
	var environment Environment
	path := fmt.Sprintf("endpoints/%d", id)
	if err := s.client.Get(path, &environment); err != nil {
		return nil, fmt.Errorf("failed to get environment %d: %w", id, err)
	}
	return &environment, nil
}

func (s *EnvironmentService) GetByName(name string) (*Environment, error) {
	environments, err := s.List()
	if err != nil {
		return nil, err
	}

	for _, env := range environments {
		if env.Name == name {
			return &env, nil
		}
	}

	return nil, fmt.Errorf("environment '%s' not found", name)
}

func (s *EnvironmentService) Delete(id int) error {
	path := fmt.Sprintf("endpoints/%d", id)
	if err := s.client.Delete(path); err != nil {
		return fmt.Errorf("failed to delete environment %d: %w", id, err)
	}
	return nil
}

func (env *Environment) TypeString() string {
	switch env.Type {
	case EnvironmentTypeDockerLocal:
		return "Docker (Local)"
	case EnvironmentTypeAgentOnDocker:
		return "Docker (Agent)"
	case EnvironmentTypeAzure:
		return "Azure"
	case EnvironmentTypeAgentOnKubernetes:
		return "Kubernetes (Agent)"
	case EnvironmentTypeEdgeAgentOnDocker:
		return "Docker (Edge)"
	case EnvironmentTypeEdgeAgentOnKubernetes:
		return "Kubernetes (Edge)"
	case EnvironmentTypeKubeLocal:
		return "Kubernetes (Local)"
	default:
		return fmt.Sprintf("Unknown (%d)", env.Type)
	}
}

func (env *Environment) StatusString() string {
	switch env.Status {
	case EnvironmentStatusUp:
		return "Up"
	case EnvironmentStatusDown:
		return "Down"
	default:
		return "Unknown"
	}
}

func (env *Environment) GetLatestSnapshot() *Snapshot {
	if len(env.Snapshots) == 0 {
		return nil
	}

	latest := &env.Snapshots[0]
	for i := range env.Snapshots {
		if env.Snapshots[i].Time > latest.Time {
			latest = &env.Snapshots[i]
		}
	}
	return latest
}
