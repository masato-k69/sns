package model

import "github.com/google/uuid"

type Topic struct {
	ID       uuid.UUID
	Name     string
	Contents []Content
	Created  *Member
	LastPost *Post
}
