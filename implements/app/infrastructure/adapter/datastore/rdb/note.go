package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type NoteStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}
type noteStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements noteStoreConnection.
func (u *noteStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements noteStoreConnection.
func (u *noteStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewNoteStoreConnection(i *do.Injector) (NoteStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_NOTE_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_NOTE_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &noteStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
