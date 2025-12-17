package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type ImageService struct {
	client *Client
}

type Image struct {
	Id          string            `json:"Id"`
	RepoTags    []string          `json:"RepoTags"`
	RepoDigests []string          `json:"RepoDigests"`
	Parent      string            `json:"Parent"`
	Comment     string            `json:"Comment"`
	Created     int64             `json:"Created"`
	Container   string            `json:"Container"`
	DockerVersion string          `json:"DockerVersion"`
	Author      string            `json:"Author"`
	Architecture string           `json:"Architecture"`
	Os          string            `json:"Os"`
	Size        int64             `json:"Size"`
	VirtualSize int64             `json:"VirtualSize"`
	Labels      map[string]string `json:"Labels"`
}

type ImageDetails struct {
	Id              string                 `json:"Id"`
	RepoTags        []string               `json:"RepoTags"`
	RepoDigests     []string               `json:"RepoDigests"`
	Parent          string                 `json:"Parent"`
	Comment         string                 `json:"Comment"`
	Created         string                 `json:"Created"`
	Container       string                 `json:"Container"`
	ContainerConfig *ContainerConfig       `json:"ContainerConfig"`
	DockerVersion   string                 `json:"DockerVersion"`
	Author          string                 `json:"Author"`
	Config          *ContainerConfig       `json:"Config"`
	Architecture    string                 `json:"Architecture"`
	Os              string                 `json:"Os"`
	Size            int64                  `json:"Size"`
	VirtualSize     int64                  `json:"VirtualSize"`
	GraphDriver     GraphDriver            `json:"GraphDriver"`
	RootFS          RootFS                 `json:"RootFS"`
	Metadata        map[string]interface{} `json:"Metadata"`
}

type GraphDriver struct {
	Name string            `json:"Name"`
	Data map[string]string `json:"Data"`
}

type RootFS struct {
	Type   string   `json:"Type"`
	Layers []string `json:"Layers"`
}

type ImagePullRequest struct {
	Image    string `json:"Image"`
	Registry string `json:"Registry,omitempty"`
}

type Registry struct {
	Id                      int                  `json:"Id"`
	Type                    int                  `json:"Type"`
	Name                    string               `json:"Name"`
	URL                     string               `json:"URL"`
	Authentication          bool                 `json:"Authentication"`
	Username                string               `json:"Username"`
	Password                string               `json:"Password,omitempty"`
	RegistryAccesses        *RegistryAccesses    `json:"RegistryAccesses,omitempty"`
	Gitlab                  *GitlabRegistryData  `json:"Gitlab,omitempty"`
	Quay                    *QuayRegistryData    `json:"Quay,omitempty"`
	ManagementConfiguration *ManagementConfig    `json:"ManagementConfiguration,omitempty"`
}

type RegistryAccesses struct {
	UserAccessPolicies map[int]RegistryAccessPolicy `json:"UserAccessPolicies"`
	TeamAccessPolicies map[int]RegistryAccessPolicy `json:"TeamAccessPolicies"`
	Namespaces         []string                     `json:"Namespaces"`
}

type RegistryAccessPolicy struct {
	RoleId int `json:"RoleId"`
}

type GitlabRegistryData struct {
	ProjectId   int    `json:"ProjectId"`
	InstanceURL string `json:"InstanceURL"`
	ProjectPath string `json:"ProjectPath"`
}

type QuayRegistryData struct {
	UseOrganisation  bool   `json:"UseOrganisation"`
	OrganisationName string `json:"OrganisationName"`
}

type ManagementConfig struct {
	Type           int    `json:"Type"`
	Authentication bool   `json:"Authentication"`
	Username       string `json:"Username"`
	Password       string `json:"Password,omitempty"`
	TLSConfig      TLSConfiguration `json:"TLSConfig"`
}

const (
	RegistryTypeQuay           = 1
	RegistryTypeAzure          = 2
	RegistryTypeCustom         = 3
	RegistryTypeGitlab         = 4
	RegistryTypeProGet         = 5
	RegistryTypeDockerHub      = 6
	RegistryTypeECR            = 7
)

func NewImageService(client *Client) *ImageService {
	return &ImageService{client: client}
}

func (s *ImageService) List(endpointID int) ([]Image, error) {
	path := fmt.Sprintf("endpoints/%d/docker/images/json", endpointID)
	
	var images []Image
	if err := s.client.Get(path, &images); err != nil {
		return nil, fmt.Errorf("failed to list images: %w", err)
	}
	return images, nil
}

func (s *ImageService) Inspect(endpointID int, imageID string) (*ImageDetails, error) {
	path := fmt.Sprintf("endpoints/%d/docker/images/%s/json", endpointID, url.PathEscape(imageID))
	
	var image ImageDetails
	if err := s.client.Get(path, &image); err != nil {
		return nil, fmt.Errorf("failed to inspect image: %w", err)
	}
	return &image, nil
}

func (s *ImageService) Pull(endpointID int, imageName string, registryID int) error {
	path := fmt.Sprintf("endpoints/%d/docker/images/create?fromImage=%s", endpointID, url.QueryEscape(imageName))
	
	if registryID > 0 {
		path += fmt.Sprintf("&X-Registry-Auth=%d", registryID)
	}

	req, err := s.client.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.do(req)
	if err != nil {
		return fmt.Errorf("failed to pull image: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

func (s *ImageService) Remove(endpointID int, imageID string, force bool) error {
	path := fmt.Sprintf("endpoints/%d/docker/images/%s?force=%t", endpointID, url.PathEscape(imageID), force)
	
	if err := s.client.Delete(path); err != nil {
		return fmt.Errorf("failed to remove image: %w", err)
	}
	return nil
}

func (s *ImageService) Tag(endpointID int, imageID, repo, tag string) error {
	path := fmt.Sprintf("endpoints/%d/docker/images/%s/tag?repo=%s&tag=%s",
		endpointID, url.PathEscape(imageID), url.QueryEscape(repo), url.QueryEscape(tag))
	
	req, err := s.client.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.do(req)
	if err != nil {
		return fmt.Errorf("failed to tag image: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

func (s *ImageService) Push(endpointID int, imageName string, registryID int) error {
	path := fmt.Sprintf("endpoints/%d/docker/images/%s/push", endpointID, url.PathEscape(imageName))
	
	if registryID > 0 {
		path += fmt.Sprintf("?X-Registry-Auth=%d", registryID)
	}

	req, err := s.client.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.do(req)
	if err != nil {
		return fmt.Errorf("failed to push image: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

func (s *ImageService) Prune(endpointID int, dangling bool) error {
	filters := map[string][]string{}
	if dangling {
		filters["dangling"] = []string{"true"}
	}

	filtersJSON, err := json.Marshal(filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	path := fmt.Sprintf("endpoints/%d/docker/images/prune?filters=%s", endpointID, url.QueryEscape(string(filtersJSON)))
	
	req, err := s.client.newRequest(http.MethodPost, path, nil)
	if err != nil {
		return err
	}

	resp, err := s.client.do(req)
	if err != nil {
		return fmt.Errorf("failed to prune images: %w", err)
	}
	defer resp.Body.Close()

	return checkResponse(resp)
}

func (image *Image) GetShortID() string {
	if strings.HasPrefix(image.Id, "sha256:") {
		return image.Id[7:19]
	}
	if len(image.Id) > 12 {
		return image.Id[:12]
	}
	return image.Id
}

func (image *Image) GetPrimaryTag() string {
	if len(image.RepoTags) > 0 && image.RepoTags[0] != "<none>:<none>" {
		return image.RepoTags[0]
	}
	if len(image.RepoDigests) > 0 {
		return image.RepoDigests[0]
	}
	return "<none>"
}

func (image *Image) GetRepository() string {
	tag := image.GetPrimaryTag()
	if tag == "<none>" {
		return "<none>"
	}
	parts := strings.Split(tag, ":")
	if len(parts) > 0 {
		return parts[0]
	}
	return tag
}

func (image *Image) GetTag() string {
	tag := image.GetPrimaryTag()
	if tag == "<none>" {
		return "<none>"
	}
	parts := strings.Split(tag, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return "latest"
}

func (r *Registry) TypeString() string {
	switch r.Type {
	case RegistryTypeQuay:
		return "Quay"
	case RegistryTypeAzure:
		return "Azure"
	case RegistryTypeCustom:
		return "Custom"
	case RegistryTypeGitlab:
		return "GitLab"
	case RegistryTypeProGet:
		return "ProGet"
	case RegistryTypeDockerHub:
		return "DockerHub"
	case RegistryTypeECR:
		return "ECR"
	default:
		return fmt.Sprintf("Unknown (%d)", r.Type)
	}
}
