package mq

import (
	"context"
	"encoding/json"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/samber/do"
)

type ActivityStoreConnection interface {
	Publish(c context.Context, exchange ExchangeName, routingKey RoutingKey, v any) error
}
type activityStoreConnection struct {
	channel *amqp.Channel
}

var (
	ExchangeActivity             ExchangeName = ExchangeName(os.Getenv("RABBITMQ_PUBLISH_EXCHANGE_ACTIVITY"))
	RoutingKeyActivityUser       RoutingKey   = RoutingKey(os.Getenv("RABBITMQ_PUBLISH_ROUTINGKEY_ACTIVITY_USER"))
	RoutingKeyActivityMember     RoutingKey   = RoutingKey(os.Getenv("RABBITMQ_PUBLISH_ROUTINGKEY_ACTIVITY_MEMBER"))
	RoutingKeyActivityMemberLike RoutingKey   = RoutingKey(os.Getenv("RABBITMQ_PUBLISH_ROUTINGKEY_ACTIVITY_MEMBER_LIKE"))
)

// Publish implements ActivityStoreConnection.
func (r *activityStoreConnection) Publish(c context.Context, exchange ExchangeName, routingKey RoutingKey, v any) error {
	message, err := createMessage(c, v)
	if err != nil {
		return err
	}

	return r.channel.Publish(exchange.String(), routingKey.String(), false, false, *message)
}

func NewActivityStoreConnection(i *do.Injector) (ActivityStoreConnection, error) {
	var config ConnectionConfig
	if err := json.Unmarshal([]byte(os.Getenv("RABBITMQ_CONNECTION")), &config); err != nil {
		return nil, err
	}

	channel, err := getConnection(config)

	if err != nil {
		return nil, err
	}

	return &activityStoreConnection{
		channel: channel,
	}, nil
}
