package service

import (
	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/servers"
	"github.com/mikrocloud/mikrocloud/internal/domain/servers/repository"
)

type ServersService struct {
	repo *repository.ServersRepository
}

func NewServersService(repo *repository.ServersRepository) *ServersService {
	return &ServersService{repo: repo}
}

func (s *ServersService) CreateServer(name, hostname, ipAddress string, port int, serverType servers.ServerType, organizationID uuid.UUID) (*servers.Server, error) {
	serverName, err := servers.NewServerName(name)
	if err != nil {
		return nil, err
	}

	server := servers.NewServer(serverName, hostname, ipAddress, port, serverType, organizationID)
	if err := s.repo.Create(server); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *ServersService) GetServer(id servers.ServerID) (*servers.Server, error) {
	return s.repo.GetByID(id)
}

func (s *ServersService) UpdateServer(server *servers.Server) error {
	return s.repo.Update(server)
}

func (s *ServersService) DeleteServer(id servers.ServerID) error {
	return s.repo.Delete(id)
}

func (s *ServersService) ListServersByOrganization(organizationID uuid.UUID) ([]*servers.Server, error) {
	return s.repo.ListByOrganization(organizationID)
}

func (s *ServersService) ListServersByType(organizationID uuid.UUID, serverType servers.ServerType) ([]*servers.Server, error) {
	return s.repo.ListByType(organizationID, serverType)
}

func (s *ServersService) GetServerByHostname(hostname string) (*servers.Server, error) {
	return s.repo.GetByHostname(hostname)
}
