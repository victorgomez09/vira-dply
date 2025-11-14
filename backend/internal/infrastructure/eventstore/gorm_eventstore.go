package eventstore

import (
	"context"
	"encoding/json"
	"errors"

	"gorm.io/gorm"
)

type EventRecord struct {
	ID          uint   `gorm:"primaryKey"`
	AggregateID string `gorm:"index"`
	Version     int
	Payload     []byte
}

type GormEventStore struct {
	db *gorm.DB
}

func NewGormEventStore(db *gorm.DB) *GormEventStore {
	db.AutoMigrate(&EventRecord{})
	return &GormEventStore{db}
}

func (s *GormEventStore) Load(ctx context.Context, aggregateID string) ([]interface{}, int, error) {
	var records []EventRecord
	err := s.db.WithContext(ctx).Where("aggregate_id = ?", aggregateID).
		Order("version asc").Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	var events []interface{}
	for _, r := range records {
		var raw map[string]interface{}
		json.Unmarshal(r.Payload, &raw)
		events = append(events, raw)
	}

	return events, len(records), nil
}

func (s *GormEventStore) Append(ctx context.Context, aggregateID string, expectedVersion int, events []interface{}) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var current int
		tx.Raw("SELECT COALESCE(MAX(version), 0) FROM event_records WHERE aggregate_id = ?", aggregateID).
			Scan(&current)

		if current != expectedVersion {
			return errors.New("concurrency conflict")
		}

		for i, evt := range events {
			data, _ := json.Marshal(evt)
			record := EventRecord{
				AggregateID: aggregateID,
				Version:     expectedVersion + i + 1,
				Payload:     data,
			}
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
		}

		return nil
	})
}
