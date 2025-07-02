package model

import (
	"time"

	"github.com/google/uuid"
)

type Invite struct {
	ID      uuid.UUID
	RoleID  uuid.UUID
	Message *Text
	At      time.Time
	Users   []uuid.UUID
}
