package model

type Community struct {
	ID         string `gorm:"primaryKey"`
	Name       string
	Invitation bool
}
