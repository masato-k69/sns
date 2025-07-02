package model

import "github.com/google/uuid"

type Post struct {
	ID       uuid.UUID
	At       int
	Contents []Content
	Created  *Member
	Reaction Reaction
}
