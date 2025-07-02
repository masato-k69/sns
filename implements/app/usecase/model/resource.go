package model

import "github.com/google/uuid"

type ResourceSearchIndex struct {
	ResourceID uuid.UUID
	Type       string
	Keyword    string
}
