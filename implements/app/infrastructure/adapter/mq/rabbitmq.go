package mq

import (
	lcontext "app/lib/context"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/labstack/echo/v4"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConnectionConfig struct {
	Protocol string `json:"protocol"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
}

type ExchangeName string

func (e ExchangeName) String() string {
	return string(e)
}

type RoutingKey string

func (r RoutingKey) String() string {
	return string(r)
}

func getConnection(config ConnectionConfig) (*amqp.Channel, error) {
	connection, err := amqp.Dial(fmt.Sprintf("%v://%v:%v@%v:%v/",
		config.Protocol,
		config.User,
		config.Password,
		config.Host,
		config.Port),
	)
	if err != nil {
		return nil, err
	}

	return connection.Channel()
}

func createMessage(c context.Context, body any) (*amqp.Publishing, error) {
	bytes, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	requestID, ok := c.Value(lcontext.ContextKeyRequestID).(string)
	if !ok || requestID == "" {
		return nil, errors.New("failied to get request id")
	}

	return &amqp.Publishing{
		ContentType: echo.MIMETextPlain,
		Headers: amqp.Table{
			lcontext.ContextKeyRequestID.String(): requestID,
		},
		Body: bytes,
	}, nil
}
