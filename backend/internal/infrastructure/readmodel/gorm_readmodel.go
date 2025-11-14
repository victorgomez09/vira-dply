package readmodel

import (
	"context"

	"gorm.io/gorm"
)

type OrderReadModel struct {
	ID     string `gorm:"primaryKey"`
	Status string
}

type GormReadModel struct {
	db *gorm.DB
}

func NewGormReadModel(db *gorm.DB) *GormReadModel {
	db.AutoMigrate(&OrderReadModel{})
	return &GormReadModel{db}
}

func (r *GormReadModel) InsertOrder(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Create(&OrderReadModel{
		ID:     id,
		Status: "CREATED",
	}).Error
}

func (r *GormReadModel) UpdateOrderPaid(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&OrderReadModel{}).
		Where("id = ?", id).Update("status", "PAID").Error
}
