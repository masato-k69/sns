package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type CommunityStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}

type communityStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements communityStoreConnection.
func (u *communityStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements communityStoreConnection.
func (u *communityStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewCommunityStoreConnection(i *do.Injector) (CommunityStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_COMMUNITY_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_COMMUNITY_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &communityStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
