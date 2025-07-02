package searchengine

import (
	"github.com/elastic/go-elasticsearch/v8"
)

type ConnectionConfig struct {
	Nodes []string `json:"nodes"`
}

func getClient(config ConnectionConfig) (*elasticsearch.TypedClient, error) {
	return elasticsearch.NewTypedClient(elasticsearch.Config{
		Addresses: config.Nodes,
	})
}
