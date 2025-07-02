package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type MemberStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}
type memberStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements memberStoreConnection.
func (u *memberStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements memberStoreConnection.
func (u *memberStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewMemberStoreConnection(i *do.Injector) (MemberStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_MEMBER_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_MEMBER_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &memberStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
