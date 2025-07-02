package model

type AuthenticatedUser struct {
	Subject  string
	Email    string
	Issuer   string
	Name     string
	ImageURL *string
}
