package subscriber

import (
	lcontext "app/lib/context"
	llog "app/lib/log"
	"app/presentation/subscriber/implement"
	"app/presentation/subscriber/interfaces"
	"context"
	"encoding/json"
	"fmt"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type ExchangeName string

func (e ExchangeName) String() string {
	return string(e)
}

type QueueName string

func (q QueueName) String() string {
	return string(q)
}

type ConnectionConfig struct {
	Protocol  string `json:"protocol"`
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"user"`
	Password  string `json:"password"`
	Exchanges []struct {
		Name       ExchangeName           `json:"name"`
		Type       string                 `json:"type"`
		Durable    bool                   `json:"durable"`     // プロセス再起動時に定義を残すか否か
		AuthDelete bool                   `json:"auto_delete"` // すべてのBindingが無くなった時に削除するか否か
		Internal   bool                   `json:"internal"`    // true に設定すると Publisher から直接使用されることはなくなり、他の Exchange から使用されるだけとなる
		NoWait     bool                   `json:"no_wait"`     // MQからの応答を待たないか否か
		Arguments  map[string]interface{} `json:"arguments"`
		Queues     []struct {
			Name       QueueName              `json:"name"`
			Durable    bool                   `json:"durable"`     // プロセス再起動時に定義を残すか否か
			AuthDelete bool                   `json:"auto_delete"` // すべてのConsumerが無くなった時に削除するか否か
			Exclusive  bool                   `json:"exclusive"`   // 接続が切れた際に定義を残すか否か
			NoWait     bool                   `json:"no_wait"`     // MQからの応答を待たないか否か
			Arguments  map[string]interface{} `json:"arguments"`
		} `json:"queues"`
	} `json:"echanges"`
}

var (
	handlers = map[ExchangeName]map[QueueName]interfaces.Handler{
		"resource_search_index": {
			"create": implement.NewResourceSearchIndexCreateHandler(),
			"update": implement.NewResourceSearchIndexUpdateHandler(),
			"delete": implement.NewResourceSearchIndexDeleteHandler(),
		},
		"activity": {
			"user":        implement.NewUserLoginActivityHandler(),
			"member":      implement.NewMemberActivityHandler(),
			"member_like": implement.NewMemberLikeActivityHandler(),
		},
	}
)

func Start() {
	connectionConfig := ConnectionConfig{}
	if err := json.Unmarshal([]byte(os.Getenv("SUBSCRIBE_CONFIG")), &connectionConfig); err != nil {
		panic(err)
	}

	connection, err := amqp.Dial(fmt.Sprintf("%v://%v:%v@%v:%v/",
		connectionConfig.Protocol,
		connectionConfig.User,
		connectionConfig.Password,
		connectionConfig.Host,
		connectionConfig.Port),
	)
	if err != nil {
		panic(err)
	}

	defer connection.Close()

	channel, err := connection.Channel()

	if err != nil {
		panic(err)
	}

	defer channel.Close()

	forever := make(chan bool)

	for _, exchangeConfig := range connectionConfig.Exchanges {
		if err := channel.ExchangeDeclare(
			exchangeConfig.Name.String(),
			exchangeConfig.Type,
			exchangeConfig.Durable,
			exchangeConfig.AuthDelete,
			exchangeConfig.Internal,
			exchangeConfig.NoWait,
			exchangeConfig.Arguments,
		); err != nil {
			panic(err)
		}

		for _, queueConfig := range exchangeConfig.Queues {
			handler, ok := handlers[exchangeConfig.Name][queueConfig.Name]
			if !ok {
				panic(fmt.Errorf("consumer does not exist. queue=%v", queueConfig.Name))
			}

			if _, err := channel.QueueDeclare(
				queueConfig.Name.String(),
				queueConfig.Durable,
				queueConfig.AuthDelete,
				queueConfig.Exclusive,
				queueConfig.NoWait,
				queueConfig.Arguments,
			); err != nil {
				panic(err)
			}

			if err := channel.QueueBind(
				queueConfig.Name.String(),
				queueConfig.Name.String(),
				exchangeConfig.Name.String(),
				false,
				nil,
			); err != nil {
				panic(err)
			}

			messages, err := channel.Consume(queueConfig.Name.String(), "", true, false, false, false, nil)
			if err != nil {
				panic(err)
			}

			fmt.Printf("start consumer. exchange=%v queue=%v \n", exchangeConfig.Name, queueConfig.Name)
			go func() {
				for message := range messages {
					body := message.Body
					c := context.Background()

					requestID, ok := message.Headers[lcontext.ContextKeyRequestID.String()].(string)
					if !ok || requestID == "" {
						llog.Error(c, "failied to get request id. exchange=%v queue=%v body=%v", exchangeConfig.Name, queueConfig.Name.String(), string(body))
						continue
					}

					c = context.WithValue(c, lcontext.ContextKeyRequestID, requestID)
					if err := func() error {
						defer func() {
							if r := recover(); r != nil {
								err = fmt.Errorf("recovered. from=%v", r)
							}
						}()

						llog.Info(c, "handle. exchange=%v queue=%v body=%v", exchangeConfig.Name, queueConfig.Name.String(), string(body))
						return handler.Handle(c, body)
					}(); err != nil {
						llog.Error(c, "failed to handle. exchange=%v queue=%v body=%v err=%v", exchangeConfig.Name, queueConfig.Name.String(), string(body), err)
					}
				}
			}()
		}
	}
	<-forever
}
