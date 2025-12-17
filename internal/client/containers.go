package client

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type ContainerService struct {
	client *Client
}

type Container struct {
	Id              string            `json:"Id"`
	Names           []string          `json:"Names"`
	Image           string            `json:"Image"`
	ImageID         string            `json:"ImageID"`
	Command         string            `json:"Command"`
	Created         int64             `json:"Created"`
	State           string            `json:"State"`
	Status          string            `json:"Status"`
	Ports           []Port            `json:"Ports"`
	Labels          map[string]string `json:"Labels"`
	SizeRw          int64             `json:"SizeRw,omitempty"`
	SizeRootFs      int64             `json:"SizeRootFs,omitempty"`
	HostConfig      HostConfig        `json:"HostConfig,omitempty"`
	NetworkSettings NetworkSettings   `json:"NetworkSettings,omitempty"`
	Mounts          []Mount           `json:"Mounts"`
}

type ContainerDetails struct {
	Id              string                 `json:"Id"`
	Created         string                 `json:"Created"`
	Path            string                 `json:"Path"`
	Args            []string               `json:"Args"`
	State           ContainerState         `json:"State"`
	Image           string                 `json:"Image"`
	ResolvConfPath  string                 `json:"ResolvConfPath"`
	HostnamePath    string                 `json:"HostnamePath"`
	HostsPath       string                 `json:"HostsPath"`
	LogPath         string                 `json:"LogPath"`
	Name            string                 `json:"Name"`
	RestartCount    int                    `json:"RestartCount"`
	Driver          string                 `json:"Driver"`
	Platform        string                 `json:"Platform"`
	MountLabel      string                 `json:"MountLabel"`
	ProcessLabel    string                 `json:"ProcessLabel"`
	AppArmorProfile string                 `json:"AppArmorProfile"`
	Config          ContainerConfig        `json:"Config"`
	NetworkSettings ContainerNetworkSettings `json:"NetworkSettings"`
	Mounts          []Mount                `json:"Mounts"`
}

type ContainerState struct {
	Status     string `json:"Status"`
	Running    bool   `json:"Running"`
	Paused     bool   `json:"Paused"`
	Restarting bool   `json:"Restarting"`
	OOMKilled  bool   `json:"OOMKilled"`
	Dead       bool   `json:"Dead"`
	Pid        int    `json:"Pid"`
	ExitCode   int    `json:"ExitCode"`
	Error      string `json:"Error"`
	StartedAt  string `json:"StartedAt"`
	FinishedAt string `json:"FinishedAt"`
}

type ContainerConfig struct {
	Hostname     string            `json:"Hostname"`
	Domainname   string            `json:"Domainname"`
	User         string            `json:"User"`
	AttachStdin  bool              `json:"AttachStdin"`
	AttachStdout bool              `json:"AttachStdout"`
	AttachStderr bool              `json:"AttachStderr"`
	Tty          bool              `json:"Tty"`
	OpenStdin    bool              `json:"OpenStdin"`
	StdinOnce    bool              `json:"StdinOnce"`
	Env          []string          `json:"Env"`
	Cmd          []string          `json:"Cmd"`
	Image        string            `json:"Image"`
	Volumes      map[string]struct{} `json:"Volumes"`
	WorkingDir   string            `json:"WorkingDir"`
	Entrypoint   []string          `json:"Entrypoint"`
	Labels       map[string]string `json:"Labels"`
}

type ContainerNetworkSettings struct {
	Bridge                 string                       `json:"Bridge"`
	SandboxID              string                       `json:"SandboxID"`
	HairpinMode            bool                         `json:"HairpinMode"`
	LinkLocalIPv6Address   string                       `json:"LinkLocalIPv6Address"`
	LinkLocalIPv6PrefixLen int                          `json:"LinkLocalIPv6PrefixLen"`
	Ports                  map[string][]PortBinding     `json:"Ports"`
	SandboxKey             string                       `json:"SandboxKey"`
	SecondaryIPAddresses   []string                     `json:"SecondaryIPAddresses"`
	SecondaryIPv6Addresses []string                     `json:"SecondaryIPv6Addresses"`
	EndpointID             string                       `json:"EndpointID"`
	Gateway                string                       `json:"Gateway"`
	GlobalIPv6Address      string                       `json:"GlobalIPv6Address"`
	GlobalIPv6PrefixLen    int                          `json:"GlobalIPv6PrefixLen"`
	IPAddress              string                       `json:"IPAddress"`
	IPPrefixLen            int                          `json:"IPPrefixLen"`
	IPv6Gateway            string                       `json:"IPv6Gateway"`
	MacAddress             string                       `json:"MacAddress"`
	Networks               map[string]EndpointSettings  `json:"Networks"`
}

type Port struct {
	IP          string `json:"IP,omitempty"`
	PrivatePort int    `json:"PrivatePort"`
	PublicPort  int    `json:"PublicPort,omitempty"`
	Type        string `json:"Type"`
}

type PortBinding struct {
	HostIP   string `json:"HostIp"`
	HostPort string `json:"HostPort"`
}

type HostConfig struct {
	NetworkMode string `json:"NetworkMode,omitempty"`
}

type NetworkSettings struct {
	Networks map[string]EndpointSettings `json:"Networks,omitempty"`
}

type EndpointSettings struct {
	IPAMConfig          *EndpointIPAMConfig `json:"IPAMConfig,omitempty"`
	Links               []string            `json:"Links"`
	Aliases             []string            `json:"Aliases"`
	NetworkID           string              `json:"NetworkID"`
	EndpointID          string              `json:"EndpointID"`
	Gateway             string              `json:"Gateway"`
	IPAddress           string              `json:"IPAddress"`
	IPPrefixLen         int                 `json:"IPPrefixLen"`
	IPv6Gateway         string              `json:"IPv6Gateway"`
	GlobalIPv6Address   string              `json:"GlobalIPv6Address"`
	GlobalIPv6PrefixLen int                 `json:"GlobalIPv6PrefixLen"`
	MacAddress          string              `json:"MacAddress"`
}

type EndpointIPAMConfig struct {
	IPv4Address string `json:"IPv4Address,omitempty"`
	IPv6Address string `json:"IPv6Address,omitempty"`
}

type Mount struct {
	Type        string `json:"Type"`
	Name        string `json:"Name,omitempty"`
	Source      string `json:"Source"`
	Destination string `json:"Destination"`
	Driver      string `json:"Driver,omitempty"`
	Mode        string `json:"Mode"`
	RW          bool   `json:"RW"`
	Propagation string `json:"Propagation"`
}

func NewContainerService(client *Client) *ContainerService {
	return &ContainerService{client: client}
}

func (s *ContainerService) List(endpointID int, all bool) ([]Container, error) {
	path := fmt.Sprintf("endpoints/%d/docker/containers/json", endpointID)
	if all {
		path += "?all=true"
	}

	var containers []Container
	if err := s.client.Get(path, &containers); err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}
	return containers, nil
}

func (s *ContainerService) Inspect(endpointID int, containerID string) (*ContainerDetails, error) {
	path := fmt.Sprintf("endpoints/%d/docker/containers/%s/json", endpointID, containerID)

	var container ContainerDetails
	if err := s.client.Get(path, &container); err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}
	return &container, nil
}

func (s *ContainerService) Logs(endpointID int, containerID string, follow bool, tail int, stdout, stderr bool) (io.ReadCloser, error) {
	params := url.Values{}
	params.Set("stdout", fmt.Sprintf("%t", stdout))
	params.Set("stderr", fmt.Sprintf("%t", stderr))
	params.Set("follow", fmt.Sprintf("%t", follow))
	if tail > 0 {
		params.Set("tail", fmt.Sprintf("%d", tail))
	} else {
		params.Set("tail", "all")
	}
	params.Set("timestamps", "true")

	path := fmt.Sprintf("endpoints/%d/docker/containers/%s/logs?%s", endpointID, containerID, params.Encode())

	req, err := s.client.newRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create logs request: %w", err)
	}

	resp, err := s.client.do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("failed to get logs: HTTP %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (s *ContainerService) Start(endpointID int, containerID string) error {
	path := fmt.Sprintf("endpoints/%d/docker/containers/%s/start", endpointID, containerID)
	return s.client.Post(path, nil, nil)
}

func (s *ContainerService) Stop(endpointID int, containerID string) error {
	path := fmt.Sprintf("endpoints/%d/docker/containers/%s/stop", endpointID, containerID)
	return s.client.Post(path, nil, nil)
}

func (s *ContainerService) Restart(endpointID int, containerID string) error {
	path := fmt.Sprintf("endpoints/%d/docker/containers/%s/restart", endpointID, containerID)
	return s.client.Post(path, nil, nil)
}

func (s *ContainerService) Remove(endpointID int, containerID string, force bool) error {
	path := fmt.Sprintf("endpoints/%d/docker/containers/%s", endpointID, containerID)
	if force {
		path += "?force=true"
	}
	return s.client.Delete(path)
}

func (c *Container) GetName() string {
	if len(c.Names) > 0 {
		name := c.Names[0]
		return strings.TrimPrefix(name, "/")
	}
	return c.Id[:12]
}

func (c *Container) GetShortID() string {
	if len(c.Id) > 12 {
		return c.Id[:12]
	}
	return c.Id
}

func (c *Container) GetPorts() string {
	if len(c.Ports) == 0 {
		return "-"
	}

	var ports []string
	for _, port := range c.Ports {
		if port.PublicPort > 0 {
			ports = append(ports, fmt.Sprintf("%d->%d/%s", port.PublicPort, port.PrivatePort, port.Type))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", port.PrivatePort, port.Type))
		}
	}
	return strings.Join(ports, ", ")
}

func (c *Container) GetStatus() string {
	return c.Status
}

func (c *Container) IsRunning() bool {
	return c.State == "running"
}
