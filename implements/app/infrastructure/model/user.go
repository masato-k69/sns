package model

import "database/sql"

type User struct {
	ID       string `gorm:"primaryKey"`
	Subject  string `gorm:"unique"`
	Email    string `gorm:"unique"`
	Issuer   string
	Name     string
	ImageUrl sql.NullString
}
