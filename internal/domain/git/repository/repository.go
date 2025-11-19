package repository

import (
	"context"

	"github.com/mikrocloud/mikrocloud/internal/domain/git"
)

type GitRepository interface {
	Create(ctx context.Context, source *git.GitSource) error
	GetByID(ctx context.Context, id string) (*git.GitSource, error)
	GetByUserID(ctx context.Context, userID string) ([]*git.GitSource, error)
	GetByOrgID(ctx context.Context, orgID string) ([]*git.GitSource, error)
	Update(ctx context.Context, id string, source *git.GitSource) error
	Delete(ctx context.Context, id string) error
}
