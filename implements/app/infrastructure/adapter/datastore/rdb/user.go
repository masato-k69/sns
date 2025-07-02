package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type UserStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}

type userStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements userStoreConnection.
func (u *userStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements userStoreConnection.
func (u *userStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewUserStoreConnection(i *do.Injector) (UserStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_USER_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_USER_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &userStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
