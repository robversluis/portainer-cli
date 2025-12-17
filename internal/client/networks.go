package client

import (
	"fmt"
	"net/url"
)

type NetworkService struct {
	client *Client
}

type Network struct {
	Name       string                      `json:"Name"`
	Id         string                      `json:"Id"`
	Created    string                      `json:"Created"`
	Scope      string                      `json:"Scope"`
	Driver     string                      `json:"Driver"`
	EnableIPv6 bool                        `json:"EnableIPv6"`
	IPAM       IPAM                        `json:"IPAM"`
	Internal   bool                        `json:"Internal"`
	Attachable bool                        `json:"Attachable"`
	Ingress    bool                        `json:"Ingress"`
	ConfigFrom NetworkConfigReference      `json:"ConfigFrom,omitempty"`
	ConfigOnly bool                        `json:"ConfigOnly"`
	Containers map[string]NetworkContainer `json:"Containers,omitempty"`
	Options    map[string]string           `json:"Options"`
	Labels     map[string]string           `json:"Labels"`
}

type IPAM struct {
	Driver  string            `json:"Driver"`
	Options map[string]string `json:"Options"`
	Config  []IPAMConfig      `json:"Config"`
}

type IPAMConfig struct {
	Subnet     string            `json:"Subnet,omitempty"`
	IPRange    string            `json:"IPRange,omitempty"`
	Gateway    string            `json:"Gateway,omitempty"`
	AuxAddress map[string]string `json:"AuxAddress,omitempty"`
}

type NetworkConfigReference struct {
	Network string `json:"Network"`
}

type NetworkContainer struct {
	Name        string `json:"Name"`
	EndpointID  string `json:"EndpointID"`
	MacAddress  string `json:"MacAddress"`
	IPv4Address string `json:"IPv4Address"`
	IPv6Address string `json:"IPv6Address"`
}

type NetworkCreateRequest struct {
	Name           string            `json:"Name"`
	CheckDuplicate bool              `json:"CheckDuplicate,omitempty"`
	Driver         string            `json:"Driver,omitempty"`
	Internal       bool              `json:"Internal,omitempty"`
	Attachable     bool              `json:"Attachable,omitempty"`
	Ingress        bool              `json:"Ingress,omitempty"`
	EnableIPv6     bool              `json:"EnableIPv6,omitempty"`
	IPAM           *IPAM             `json:"IPAM,omitempty"`
	Options        map[string]string `json:"Options,omitempty"`
	Labels         map[string]string `json:"Labels,omitempty"`
}

type NetworkCreateResponse struct {
	Id      string `json:"Id"`
	Warning string `json:"Warning,omitempty"`
}

func NewNetworkService(client *Client) *NetworkService {
	return &NetworkService{client: client}
}

func (s *NetworkService) List(endpointID int) ([]Network, error) {
	path := fmt.Sprintf("endpoints/%d/docker/networks", endpointID)

	var networks []Network
	if err := s.client.Get(path, &networks); err != nil {
		return nil, fmt.Errorf("failed to list networks: %w", err)
	}
	return networks, nil
}

func (s *NetworkService) Inspect(endpointID int, networkID string) (*Network, error) {
	path := fmt.Sprintf("endpoints/%d/docker/networks/%s", endpointID, url.PathEscape(networkID))

	var network Network
	if err := s.client.Get(path, &network); err != nil {
		return nil, fmt.Errorf("failed to inspect network: %w", err)
	}
	return &network, nil
}

func (s *NetworkService) Create(endpointID int, req *NetworkCreateRequest) (*NetworkCreateResponse, error) {
	path := fmt.Sprintf("endpoints/%d/docker/networks/create", endpointID)

	var response NetworkCreateResponse
	if err := s.client.Post(path, req, &response); err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}
	return &response, nil
}

func (s *NetworkService) Remove(endpointID int, networkID string) error {
	path := fmt.Sprintf("endpoints/%d/docker/networks/%s", endpointID, url.PathEscape(networkID))

	if err := s.client.Delete(path); err != nil {
		return fmt.Errorf("failed to remove network: %w", err)
	}
	return nil
}

func (s *NetworkService) Prune(endpointID int) error {
	path := fmt.Sprintf("endpoints/%d/docker/networks/prune", endpointID)

	if err := s.client.Post(path, nil, nil); err != nil {
		return fmt.Errorf("failed to prune networks: %w", err)
	}
	return nil
}

func (n *Network) GetShortID() string {
	if len(n.Id) > 12 {
		return n.Id[:12]
	}
	return n.Id
}
