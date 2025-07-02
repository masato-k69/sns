package model

import "github.com/google/uuid"

type Topic struct {
	ID      uuid.UUID
	Name    Name
	Created *uuid.UUID
}
