package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/networking"
)

type NetworkRepository interface {
	Create(ctx context.Context, network *networking.Network) error
	GetByID(ctx context.Context, id networking.NetworkID) (*networking.Network, error)
	GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*networking.Network, error)
	Update(ctx context.Context, network *networking.Network) error
	Delete(ctx context.Context, id networking.NetworkID) error
}

type NetworkRuleRepository interface {
	Create(ctx context.Context, rule *networking.NetworkRule) error
	GetByID(ctx context.Context, id networking.NetworkRuleID) (*networking.NetworkRule, error)
	GetByNetworkID(ctx context.Context, networkID networking.NetworkID) ([]*networking.NetworkRule, error)
	Update(ctx context.Context, rule *networking.NetworkRule) error
	Delete(ctx context.Context, id networking.NetworkRuleID) error
}
