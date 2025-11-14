package shared

import "github.com/google/uuid"

type ID string

func NewID() ID {
	return ID(uuid.New().String())
}

func (id ID) String() string {
	return string(id)
}
