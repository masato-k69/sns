package model

import "github.com/google/uuid"

type Member struct {
	ID     uuid.UUID
	UserID uuid.UUID
	RoleID uuid.UUID
}
