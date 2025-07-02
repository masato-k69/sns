package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type ThreadStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}

type threadStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements threadStoreConnection.
func (u *threadStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements threadStoreConnection.
func (u *threadStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewThreadStoreConnection(i *do.Injector) (ThreadStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_THREAD_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_THREAD_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &threadStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
