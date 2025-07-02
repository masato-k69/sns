package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type PostStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}

type postStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements postStoreConnection.
func (u *postStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements postStoreConnection.
func (u *postStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewPostStoreConnection(i *do.Injector) (PostStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_POST_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_POST_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &postStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
