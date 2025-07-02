package model

import "github.com/google/uuid"

type Community struct {
	ID         uuid.UUID
	Name       string
	Invitation bool
}
