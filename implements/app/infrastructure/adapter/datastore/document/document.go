package document

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ConnectionConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DB       string `json:"db"`
}

func getConnection(config ConnectionConfig) (*mongo.Client, error) {
	dsn := fmt.Sprintf("mongodb://%v:%v@%v:%v", config.User, config.Password, config.Host, config.Port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dsn))
	if err != nil {
		return nil, err
	}

	_, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}

	fmt.Printf("connection established. %v:%v \n", config.Host, config.Port)

	return client, nil
}

func Unmarshal(v interface{}) (doc *bson.D, err error) {
	var vMap map[string]interface{}

	indirect, err := bson.Marshal(&v)
	if err != nil {
		return nil, err
	}

	if err := bson.Unmarshal(indirect, &vMap); err != nil {
		return nil, err
	}

	for k := range vMap {
		if vMap[k] == nil || vMap[k] == "" {
			delete(vMap, k)
		}
	}

	data, err := bson.Marshal(&vMap)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}
