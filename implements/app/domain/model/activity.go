package model

import (
	"time"

	"github.com/google/uuid"
)

type UserLoginActivity struct {
	At              time.Time
	UserID          uuid.UUID
	IPAddress       IPAddress
	OperationSystem OperationgSystem
	UserAgent       UserAgent
}

type MemberActivity struct {
	At        time.Time
	Member    uuid.UUID
	Target    uuid.UUID
	Resource  Resource
	Operation Operation
}

type MemberLikeActivity struct {
	At       time.Time
	Member   uuid.UUID
	Target   uuid.UUID
	Resource Resource
	Like     bool
	Comment  *ShortMessage
}
