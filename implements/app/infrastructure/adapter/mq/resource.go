package mq

import (
	"context"
	"encoding/json"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/samber/do"
)

type ResourceSearchIndexStoreConnection interface {
	Publish(c context.Context, exchange ExchangeName, routingKey RoutingKey, v any) error
}
type resourceSearchIndexStoreConnection struct {
	channel *amqp.Channel
}

var (
	ExchangeResource         ExchangeName = ExchangeName(os.Getenv("RABBITMQ_PUBLISH_EXCHANGE_RESOURCE"))
	RoutingKeyResourceCreate RoutingKey   = RoutingKey(os.Getenv("RABBITMQ_PUBLISH_ROUTINGKEY_RESOURCE_CREATE"))
	RoutingKeyResourceUpdate RoutingKey   = RoutingKey(os.Getenv("RABBITMQ_PUBLISH_ROUTINGKEY_RESOURCE_UPDATE"))
	RoutingKeyResourceDelete RoutingKey   = RoutingKey(os.Getenv("RABBITMQ_PUBLISH_ROUTINGKEY_RESOURCE_DELETE"))
)

// Publish implements ResourceSearchIndexStoreConnection.
func (r *resourceSearchIndexStoreConnection) Publish(c context.Context, exchange ExchangeName, routingKey RoutingKey, v any) error {
	message, err := createMessage(c, v)
	if err != nil {
		return err
	}

	return r.channel.Publish(exchange.String(), routingKey.String(), false, false, *message)
}

func NewResourceSearchIndexStoreConnection(i *do.Injector) (ResourceSearchIndexStoreConnection, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("RABBITMQ_CONNECTION")), &config); err != nil {
		return nil, err
	}

	channel, err := getConnection(config)

	if err != nil {
		return nil, err
	}

	return &resourceSearchIndexStoreConnection{
		channel: channel,
	}, nil
}
