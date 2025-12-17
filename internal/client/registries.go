package client

import (
	"fmt"
)

type RegistryService struct {
	client *Client
}

func NewRegistryService(client *Client) *RegistryService {
	return &RegistryService{client: client}
}

func (s *RegistryService) List() ([]Registry, error) {
	var registries []Registry
	if err := s.client.Get("registries", &registries); err != nil {
		return nil, fmt.Errorf("failed to list registries: %w", err)
	}
	return registries, nil
}

func (s *RegistryService) Get(id int) (*Registry, error) {
	path := fmt.Sprintf("registries/%d", id)

	var registry Registry
	if err := s.client.Get(path, &registry); err != nil {
		return nil, fmt.Errorf("failed to get registry: %w", err)
	}
	return &registry, nil
}

func (s *RegistryService) Create(registry *Registry) (*Registry, error) {
	var result Registry
	if err := s.client.Post("registries", registry, &result); err != nil {
		return nil, fmt.Errorf("failed to create registry: %w", err)
	}
	return &result, nil
}

func (s *RegistryService) Update(id int, registry *Registry) (*Registry, error) {
	path := fmt.Sprintf("registries/%d", id)

	var result Registry
	if err := s.client.Put(path, registry, &result); err != nil {
		return nil, fmt.Errorf("failed to update registry: %w", err)
	}
	return &result, nil
}

func (s *RegistryService) Delete(id int) error {
	path := fmt.Sprintf("registries/%d", id)

	if err := s.client.Delete(path); err != nil {
		return fmt.Errorf("failed to delete registry: %w", err)
	}
	return nil
}
