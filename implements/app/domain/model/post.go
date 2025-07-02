package model

import (
	"github.com/google/uuid"
)

type Post struct {
	ID   uuid.UUID
	At   UnixTime
	From *uuid.UUID
	To   []Mention
}
