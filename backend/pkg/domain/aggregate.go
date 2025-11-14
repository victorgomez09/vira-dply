package domain

import (
	"github.com/victorgomez09/vira-dply/pkg/infrastructure/shared"
)

type AggregateRoot interface {
	GetID() shared.ID
	GetType() string
	GetVersion() int
	GetUncommittedChanges() []Event
	MarkChangesCommitted()
	ApplyChange(Event)
}
