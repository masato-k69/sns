package model

import (
	"database/sql"
	"time"
)

type Invite struct {
	ID      string `gorm:"primaryKey"`
	RoleID  string
	Message sql.NullString
	At      time.Time
}

type InvitedUser struct {
	InviteID string `gorm:"primaryKey"`
	UserID   string `gorm:"primaryKey"`
}
