package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/networking"
	"github.com/mikrocloud/mikrocloud/internal/domain/networking/repository"
)

type NetworkService struct {
	networkRepo repository.NetworkRepository
	ruleRepo    repository.NetworkRuleRepository
}

func NewNetworkService(
	networkRepo repository.NetworkRepository,
	ruleRepo repository.NetworkRuleRepository,
) *NetworkService {
	return &NetworkService{
		networkRepo: networkRepo,
		ruleRepo:    ruleRepo,
	}
}

func (s *NetworkService) CreateNetwork(
	ctx context.Context,
	name networking.NetworkName,
	projectID uuid.UUID,
	cidr string,
	isolated bool,
) (*networking.Network, error) {
	network, err := networking.NewNetwork(name, projectID, cidr, isolated)
	if err != nil {
		return nil, fmt.Errorf("failed to create network: %w", err)
	}

	if err := s.networkRepo.Create(ctx, network); err != nil {
		return nil, fmt.Errorf("failed to save network: %w", err)
	}

	return network, nil
}

func (s *NetworkService) GetNetwork(ctx context.Context, id networking.NetworkID) (*networking.Network, error) {
	return s.networkRepo.GetByID(ctx, id)
}

func (s *NetworkService) GetNetworksByProject(ctx context.Context, projectID uuid.UUID) ([]*networking.Network, error) {
	return s.networkRepo.GetByProjectID(ctx, projectID)
}

func (s *NetworkService) UpdateNetworkStatus(ctx context.Context, id networking.NetworkID, status networking.NetworkStatus) error {
	network, err := s.networkRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	network.ChangeStatus(status)
	return s.networkRepo.Update(ctx, network)
}

func (s *NetworkService) DeleteNetwork(ctx context.Context, id networking.NetworkID) error {
	rules, err := s.ruleRepo.GetByNetworkID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to check network rules: %w", err)
	}

	if len(rules) > 0 {
		return fmt.Errorf("cannot delete network: %d rules still exist", len(rules))
	}

	return s.networkRepo.Delete(ctx, id)
}

func (s *NetworkService) CreateNetworkRule(
	ctx context.Context,
	networkID networking.NetworkID,
	protocol networking.Protocol,
	port *int,
	portRange *networking.PortRange,
	source, target string,
	action networking.RuleAction,
	description string,
) (*networking.NetworkRule, error) {
	rule, err := networking.NewNetworkRule(networkID, protocol, port, portRange, source, target, action, description)
	if err != nil {
		return nil, fmt.Errorf("failed to create network rule: %w", err)
	}

	if err := s.ruleRepo.Create(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to save network rule: %w", err)
	}

	return rule, nil
}

func (s *NetworkService) GetNetworkRule(ctx context.Context, id networking.NetworkRuleID) (*networking.NetworkRule, error) {
	return s.ruleRepo.GetByID(ctx, id)
}

func (s *NetworkService) GetNetworkRules(ctx context.Context, networkID networking.NetworkID) ([]*networking.NetworkRule, error) {
	return s.ruleRepo.GetByNetworkID(ctx, networkID)
}

func (s *NetworkService) DeleteNetworkRule(ctx context.Context, id networking.NetworkRuleID) error {
	return s.ruleRepo.Delete(ctx, id)
}
