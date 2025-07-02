package document

import (
	"encoding/json"
	"os"

	"github.com/samber/do"
	"go.mongodb.org/mongo-driver/mongo"
)

type RoleStoreConnection interface {
	DB() *mongo.Database
}

type roleStoreConnection struct {
	conn *mongo.Client
	db   string
}

// Conn implements roleStoreConnection.
func (r *roleStoreConnection) DB() *mongo.Database {
	return r.conn.Database(r.db)
}

func NewRoleStoreConnection(i *do.Injector) (RoleStoreConnection, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("MONGODB_ROLE")), &config); err != nil {
		return nil, err
	}

	conn, err := getConnection(config)

	if err != nil {
		return nil, err
	}

	return &roleStoreConnection{
		conn: conn,
		db:   config.DB,
	}, nil
}
