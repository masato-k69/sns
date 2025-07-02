package searchengine

import (
	"encoding/json"
	"os"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/samber/do"
)

type ResourceSearchIndexStoreConnection interface {
	Client() *elasticsearch.TypedClient
}

type resourceSearchIndexStoreConnection struct {
	client *elasticsearch.TypedClient
}

// Client implements ResourceSearchIndexStoreConnection.
func (r *resourceSearchIndexStoreConnection) Client() *elasticsearch.TypedClient {
	return r.client
}

func NewResourceSearchIndexStoreConnection(i *do.Injector) (ResourceSearchIndexStoreConnection, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("ELASTICSEARCH_ERSOURCE_CONNECTION")), &config); err != nil {
		return nil, err
	}

	client, err := getClient(config)

	if err != nil {
		return nil, err
	}

	return &resourceSearchIndexStoreConnection{
		client: client,
	}, nil
}
