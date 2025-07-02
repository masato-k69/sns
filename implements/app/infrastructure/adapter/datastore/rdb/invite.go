package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type InviteStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}

type inviteStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements inviteStoreConnection.
func (u *inviteStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements inviteStoreConnection.
func (u *inviteStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewInviteStoreConnection(i *do.Injector) (InviteStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_INVITE_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_INVITE_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &inviteStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
