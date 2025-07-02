package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Subject  string
	Email    string
	Name     string
	ImageUrl *string
}
