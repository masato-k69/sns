package model

import (
	"time"

	"github.com/google/uuid"
)

type CommunityInvite struct {
	ID      uuid.UUID
	Role    Role
	Users   []User
	Message *string
	At      time.Time
}

type UserInvite struct {
	ID        uuid.UUID
	Role      Role
	Community Community
	Message   *string
	At        time.Time
}
