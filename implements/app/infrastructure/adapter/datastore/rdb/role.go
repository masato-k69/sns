package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type RoleStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}
type roleStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements roleStoreConnection.
func (u *roleStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements roleStoreConnection.
func (u *roleStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewRoleStoreConnection(i *do.Injector) (RoleStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_ROLE_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_ROLE_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &roleStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
