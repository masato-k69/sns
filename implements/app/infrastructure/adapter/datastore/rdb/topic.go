package rdb

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type TopicStoreConnection interface {
	Read() *gorm.DB
	Write() *gorm.DB
}

type topicStoreConnection struct {
	connRead  *gorm.DB
	connWrite *gorm.DB
}

// Read implements topicStoreConnection.
func (u *topicStoreConnection) Read() *gorm.DB {
	return u.connRead
}

// Write implements topicStoreConnection.
func (u *topicStoreConnection) Write() *gorm.DB {
	return u.connWrite
}

func NewTopicStoreConnection(i *do.Injector) (TopicStoreConnection, error) {
	var configRead ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_TOPIC_READ")), &configRead); err != nil {
		return nil, err
	}

	read, err := getConnection(configRead)

	if err != nil {
		return nil, err
	}

	var configWrite ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MYSQL_TOPIC_WRITE")), &configWrite); err != nil {
		return nil, err
	}

	write, err := getConnection(configWrite)

	if err != nil {
		return nil, err
	}

	return &topicStoreConnection{
		connRead:  read,
		connWrite: write,
	}, nil
}
