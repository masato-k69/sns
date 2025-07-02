package v1

import (
	"encoding/json"
)

func (m *Content_Entity) Union() json.RawMessage {
	return m.union
}

func NewMessageToSend(sendType SendType, union []byte) MessagesToSend {
	return MessagesToSend{
		Type:   sendType,
		Entity: MessagesToSend_Entity{union: union},
	}
}

func NewContentEntity(union []byte) Content_Entity {
	return Content_Entity{union: union}
}
