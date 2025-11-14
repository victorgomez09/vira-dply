package projection

import (
	"time"

	"gorm.io/gorm"
)

type FailedProjection struct {
	ID          uint   `gorm:"primaryKey"`
	EventType   string `gorm:"index"`
	EventData   []byte
	Attempts    int
	MaxAttempts int
	ErrorMsg    string
	NextRetry   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type RetryStore struct {
	db *gorm.DB
}

func NewRetryStore(db *gorm.DB) *RetryStore {
	db.AutoMigrate(&FailedProjection{})
	return &RetryStore{db}
}

func (s *RetryStore) SaveFailed(eventType string, data []byte, attempts, max int, next time.Time, err error) error {
	record := FailedProjection{
		EventType:   eventType,
		EventData:   data,
		Attempts:    attempts,
		MaxAttempts: max,
		ErrorMsg:    err.Error(),
		NextRetry:   next,
	}
	return s.db.Create(&record).Error
}

func (s *RetryStore) GetDueRetries() ([]FailedProjection, error) {
	var list []FailedProjection
	err := s.db.
		Where("next_retry <= ?", time.Now()).
		Find(&list).Error
	return list, err
}

func (s *RetryStore) Delete(id uint) error {
	return s.db.Delete(&FailedProjection{}, id).Error
}
