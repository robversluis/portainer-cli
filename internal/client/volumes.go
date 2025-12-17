package client

import (
	"fmt"
	"net/url"
)

type VolumeService struct {
	client *Client
}

type Volume struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver"`
	Mountpoint string            `json:"Mountpoint"`
	CreatedAt  string            `json:"CreatedAt"`
	Status     map[string]string `json:"Status,omitempty"`
	Labels     map[string]string `json:"Labels"`
	Scope      string            `json:"Scope"`
	Options    map[string]string `json:"Options"`
}

type VolumeListResponse struct {
	Volumes  []Volume `json:"Volumes"`
	Warnings []string `json:"Warnings"`
}

type VolumeDetails struct {
	Name       string                 `json:"Name"`
	Driver     string                 `json:"Driver"`
	Mountpoint string                 `json:"Mountpoint"`
	CreatedAt  string                 `json:"CreatedAt"`
	Status     map[string]interface{} `json:"Status,omitempty"`
	Labels     map[string]string      `json:"Labels"`
	Scope      string                 `json:"Scope"`
	Options    map[string]string      `json:"Options"`
	UsageData  *VolumeUsageData       `json:"UsageData,omitempty"`
}

type VolumeUsageData struct {
	Size     int64 `json:"Size"`
	RefCount int64 `json:"RefCount"`
}

type VolumeCreateRequest struct {
	Name       string            `json:"Name"`
	Driver     string            `json:"Driver,omitempty"`
	DriverOpts map[string]string `json:"DriverOpts,omitempty"`
	Labels     map[string]string `json:"Labels,omitempty"`
}

func NewVolumeService(client *Client) *VolumeService {
	return &VolumeService{client: client}
}

func (s *VolumeService) List(endpointID int) ([]Volume, error) {
	path := fmt.Sprintf("endpoints/%d/docker/volumes", endpointID)

	var response VolumeListResponse
	if err := s.client.Get(path, &response); err != nil {
		return nil, fmt.Errorf("failed to list volumes: %w", err)
	}
	return response.Volumes, nil
}

func (s *VolumeService) Inspect(endpointID int, volumeName string) (*VolumeDetails, error) {
	path := fmt.Sprintf("endpoints/%d/docker/volumes/%s", endpointID, url.PathEscape(volumeName))

	var volume VolumeDetails
	if err := s.client.Get(path, &volume); err != nil {
		return nil, fmt.Errorf("failed to inspect volume: %w", err)
	}
	return &volume, nil
}

func (s *VolumeService) Create(endpointID int, req *VolumeCreateRequest) (*Volume, error) {
	path := fmt.Sprintf("endpoints/%d/docker/volumes/create", endpointID)

	var volume Volume
	if err := s.client.Post(path, req, &volume); err != nil {
		return nil, fmt.Errorf("failed to create volume: %w", err)
	}
	return &volume, nil
}

func (s *VolumeService) Remove(endpointID int, volumeName string, force bool) error {
	path := fmt.Sprintf("endpoints/%d/docker/volumes/%s?force=%t", endpointID, url.PathEscape(volumeName), force)

	if err := s.client.Delete(path); err != nil {
		return fmt.Errorf("failed to remove volume: %w", err)
	}
	return nil
}

func (s *VolumeService) Prune(endpointID int) error {
	path := fmt.Sprintf("endpoints/%d/docker/volumes/prune", endpointID)

	if err := s.client.Post(path, nil, nil); err != nil {
		return fmt.Errorf("failed to prune volumes: %w", err)
	}
	return nil
}
