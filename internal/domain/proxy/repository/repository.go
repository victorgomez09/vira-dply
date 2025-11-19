package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/mikrocloud/mikrocloud/internal/domain/proxy"
)

type ProxyRepository interface {
	Create(ctx context.Context, config *proxy.ProxyConfig) error
	GetByID(ctx context.Context, id proxy.ProxyConfigID) (*proxy.ProxyConfig, error)
	GetByContainerID(ctx context.Context, containerID string) (*proxy.ProxyConfig, error)
	GetByServiceName(ctx context.Context, projectID uuid.UUID, serviceName string) (*proxy.ProxyConfig, error)
	ListByProject(ctx context.Context, projectID uuid.UUID) ([]*proxy.ProxyConfig, error)
	ListAll(ctx context.Context) ([]*proxy.ProxyConfig, error)
	ListByStatus(ctx context.Context, status proxy.ProxyStatus) ([]*proxy.ProxyConfig, error)
	Update(ctx context.Context, config *proxy.ProxyConfig) error
	Delete(ctx context.Context, id proxy.ProxyConfigID) error
	DeleteByContainerID(ctx context.Context, containerID string) error
	Exists(ctx context.Context, id proxy.ProxyConfigID) (bool, error)
	ExistsByHostname(ctx context.Context, hostname string) (bool, error)
}

type TraefikConfigRepository interface {
	Create(ctx context.Context, config *proxy.TraefikGlobalConfig) error
	GetCurrent(ctx context.Context) (*proxy.TraefikGlobalConfig, error)
	Update(ctx context.Context, config *proxy.TraefikGlobalConfig) error
	Delete(ctx context.Context, id proxy.TraefikConfigID) error
	Exists(ctx context.Context) (bool, error)
}
