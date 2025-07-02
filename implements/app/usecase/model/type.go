package model

import "github.com/google/uuid"

type Mention struct {
	ID           uuid.UUID
	ResourceType string
}

type Reaction struct {
	Likes    int
	Dislikes int
}
