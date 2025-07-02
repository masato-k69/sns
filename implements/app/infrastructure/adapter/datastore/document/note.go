package document

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"go.mongodb.org/mongo-driver/mongo"
)

type NoteStoreConnection interface {
	DB() *mongo.Database
}

type noteStoreConnection struct {
	conn *mongo.Client
	db   string
}

// Conn implements noteStoreConnection.
func (u *noteStoreConnection) DB() *mongo.Database {
	return u.conn.Database(u.db)
}

func NewNoteStoreConnection(i *do.Injector) (NoteStoreConnection, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MONGODB_NOTE")), &config); err != nil {
		return nil, err
	}

	conn, err := getConnection(config)

	if err != nil {
		return nil, err
	}

	return &noteStoreConnection{
		conn: conn,
		db:   config.DB,
	}, nil
}
