package readmodel

import (
	"context"

	"gorm.io/gorm"
)

type ReadModelRepository interface {
	InsertOrder(ctx context.Context, id string) error
	UpdateOrderPaid(ctx context.Context, id string) error
}

type GormReadModelRepo struct {
	db *gorm.DB
}

func NewReadModelRepo(db *gorm.DB) *GormReadModelRepo {
	return &GormReadModelRepo{db}
}

// Insert
func (r *GormReadModelRepo) InsertOrder(ctx context.Context, id string) error {
	model := OrderReadModel{
		ID:     id,
		Status: "CREATED",
	}
	return r.db.WithContext(ctx).Create(&model).Error
}

// Update
func (r *GormReadModelRepo) UpdateOrderPaid(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).
		Model(&OrderReadModel{}).
		Where("id = ?", id).
		Update("status", "PAID").Error
}
