package v1

import (
	v1 "app/gen/api/v1"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/samber/lo"
)

var (

	// Type 内容の種類
	// * text - テキスト
	// * heading - 見出し
	// * list - 箇条書き
	// * checkbox - チェックボックス
	// * radiobutton - ラジオボタン
	// * line - 水平線
	// * img - 画像
	// * mention - メンション
	supportedContents = map[v1.ContentType]func(v1.Content_Entity) (*string, []byte, error){
		v1.ContentTypeText: func(t v1.Content_Entity) (*string, []byte, error) {
			content, err := t.AsText()
			if err != nil {
				return nil, nil, err
			}

			contentValue := content.Value
			marshaled, err := t.Union().MarshalJSON()
			return &contentValue, marshaled, err
		},
		v1.ContentTypeHeading: func(t v1.Content_Entity) (*string, []byte, error) {
			content, err := t.AsHeading()
			if err != nil {
				return nil, nil, err
			}

			contentValue := content.Value.Value
			marshaled, err := t.Union().MarshalJSON()
			return &contentValue, marshaled, err
		},
		v1.ContentTypeList: func(t v1.Content_Entity) (*string, []byte, error) {
			content, err := t.AsList()
			if err != nil {
				return nil, nil, err
			}

			contentValue := strings.Join(lo.Map(content.Values, func(elm v1.ListElement, _ int) string { return elm.Value.Value }), ",")
			marshaled, err := t.Union().MarshalJSON()
			return &contentValue, marshaled, err
		},
		v1.ContentTypeCheckbox: func(t v1.Content_Entity) (*string, []byte, error) {
			content, err := t.AsCheckBox()
			if err != nil {
				return nil, nil, err
			}

			contentValue := strings.Join(lo.Map(content.Values, func(elm v1.CheckBoxElement, _ int) string { return elm.Value.Value }), ",")
			marshaled, err := t.Union().MarshalJSON()
			return &contentValue, marshaled, err
		},
		v1.ContentTypeRadiobutton: func(t v1.Content_Entity) (*string, []byte, error) {
			content, err := t.AsRadioButton()
			if err != nil {
				return nil, nil, err
			}

			contentValue := strings.Join(lo.Map(content.Values, func(elm v1.RadioButtonElement, _ int) string { return elm.Value.Value }), ",")
			marshaled, err := t.Union().MarshalJSON()
			return &contentValue, marshaled, err
		},
		v1.ContentTypeLine: func(t v1.Content_Entity) (*string, []byte, error) {
			_, err := t.AsHorizontalLine()
			if err != nil {
				return nil, nil, err
			}

			marshaled, err := t.Union().MarshalJSON()
			return nil, marshaled, err
		},
		v1.ContentTypeImg: func(t v1.Content_Entity) (*string, []byte, error) {
			_, err := t.AsImage()
			if err != nil {
				return nil, nil, err
			}

			marshaled, err := t.Union().MarshalJSON()
			return nil, marshaled, err
		},
		v1.ContentTypeMentions: func(t v1.Content_Entity) (*string, []byte, error) {
			content, err := t.AsMentions()
			if err != nil {
				return nil, nil, err
			}

			contentValue := strings.Join(lo.Map(content, func(elm v1.Mention, _ int) string { return elm.Id.String() }), ",")
			marshaled, err := t.Union().MarshalJSON()
			return &contentValue, marshaled, err
		},
	}

	// Type 送信されるメッセージの種類
	// * current - 現在の全行
	supportedMessagesToSends = map[v1.SendType]func(v1.MessagesToSend_Entity) error{
		v1.Current: func(t v1.MessagesToSend_Entity) error {
			if _, err := t.AsCurrentLinesMessage(); err != nil {
				return err
			}
			return nil
		},
	}

	supportedResources = []v1.Resource{
		v1.ResourceChoose,
		v1.ResourceCommunity,
		v1.ResourceElection,
		v1.ResourceLike,
		v1.ResourceMember,
		v1.ResourceMilestone,
		v1.ResourcePost,
		v1.ResourceProject,
		v1.ResourceRole,
		v1.ResourceTag,
		v1.ResourceTask,
		v1.ResourceThread,
		v1.ResourceTopic,
		v1.ResourceUser,
	}

	supportedOperations = []v1.Operation{
		v1.OperationCreate,
		v1.OperationDelete,
		v1.OperationUpdate,
	}
)

func NewMessagesToSend(entity any) (*v1.MessagesToSend, error) {
	entityBin, err := json.Marshal(&entity)
	if err != nil {
		return nil, err
	}

	switch entity.(type) {
	case v1.CurrentLinesMessage:
		message := v1.NewMessageToSend(v1.Current, entityBin)
		return &message, nil
	}

	return nil, fmt.Errorf("unsupported message. v=%v", entity)
}

func NewSendType(v string) (*v1.SendType, error) {
	result := v1.SendType(v)

	for sendType := range supportedMessagesToSends {
		if result == sendType {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("invalid argument. v=%v", v)
}

func NewContent(contentType string, bytes []byte) (*v1.Content, error) {
	parsedContentType, err := NewContentType(contentType)

	if err != nil {
		return nil, err
	}

	supportedContentAs, ok := supportedContents[*parsedContentType]

	if !ok {
		return nil, fmt.Errorf("unsupported type. v=%v", parsedContentType)
	}

	entity := v1.NewContentEntity(bytes)

	if _, _, err := supportedContentAs(entity); err != nil {
		return nil, err
	}

	return &v1.Content{
		Type:   *parsedContentType,
		Entity: entity,
	}, nil
}

func NewContentType(v string) (*v1.ContentType, error) {
	result := v1.ContentType(v)

	for contentType := range supportedContents {
		if result == contentType {
			return &result, nil
		}
	}

	return nil, fmt.Errorf("invalid argument. v=%v", v)
}

type MessageReciever func(v1.MessagesToRecieve_Entity) error

func HandleRecievedMessage(bytes []byte, consumers map[v1.RecieveType]MessageReciever) error {
	var recievedMessage v1.MessagesToRecieve
	if err := json.Unmarshal(bytes, &recievedMessage); err != nil {
		return err
	}

	consumer, ok := consumers[recievedMessage.Type]
	if !ok {
		return fmt.Errorf("consumer not found. v=%v", recievedMessage.Type)
	}

	return consumer(recievedMessage.Entity)
}

func ToMetaAndBin(content v1.Content) (*string, []byte, error) {
	supportedContentAs, ok := supportedContents[content.Type]

	if !ok {
		return nil, nil, fmt.Errorf("unsupported type. v=%v", content.Type)
	}

	return supportedContentAs(content.Entity)
}

func ListMention(contents []v1.Content) (v1.Mentions, error) {
	mentions := v1.Mentions{}
	for _, content := range contents {
		if content.Type != v1.ContentTypeMentions {
			continue
		}

		content, err := content.Entity.AsMentions()
		if err != nil {
			return nil, err
		}

		mentions = slices.Concat(content)
	}

	return mentions, nil
}

func ToText(contents []v1.Content) (*string, error) {
	text := ""

	for _, content := range contents {
		supportedContentAs, ok := supportedContents[content.Type]

		if !ok {
			return nil, fmt.Errorf("unsupported type. v=%v", content.Type)
		}

		valueString, _, err := supportedContentAs(content.Entity)
		if err != nil {
			return nil, err
		}

		if valueString == nil {
			continue
		}

		text += *valueString
	}

	return &text, nil
}
