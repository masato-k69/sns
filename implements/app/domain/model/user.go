package model

import "github.com/google/uuid"

type User struct {
	ID       uuid.UUID
	Subject  Subject
	Email    EmailAddress
	Issuer   string
	Name     Name
	ImageURL *URL
}
