package model

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	Like    bool
	Comment *string
	By      *Member
}

type Login struct {
	At              time.Time
	UserID          uuid.UUID
	IPAddress       string
	OperationSystem string
	UserAgent       string
}

type Activity struct {
	At        time.Time
	Me        uuid.UUID
	Target    uuid.UUID
	Resource  string
	Operation string
}
