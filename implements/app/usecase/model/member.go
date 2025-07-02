package model

import "github.com/google/uuid"

type Member struct {
	ID   uuid.UUID
	User User
	Role *Role
}
