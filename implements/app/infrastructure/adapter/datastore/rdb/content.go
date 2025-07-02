package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type ContentStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}

type contentStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements contentStoreConnection.
func (u *contentStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements contentStoreConnection.
func (u *contentStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewContentStoreConnection(i *do.Injector) (ContentStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_CONTENT_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_CONTENT_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &contentStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
