package model

import (
	"fmt"

	"github.com/google/uuid"
)

type Content struct {
	ID    uuid.UUID
	Type  ContentType
	Value ContentValue
}

type ContentType string

const (
	ContentTypeCheckbox    ContentType = "checkbox"
	ContentTypeHeading     ContentType = "heading"
	ContentTypeImg         ContentType = "img"
	ContentTypeLine        ContentType = "line"
	ContentTypeList        ContentType = "list"
	ContentTypeMention     ContentType = "mention"
	ContentTypeRadiobutton ContentType = "radiobutton"
	ContentTypeText        ContentType = "text"
)

func (m *ContentType) String() string {
	return string(*m)
}

func NewContentType(v string) (*ContentType, error) {
	t := ContentType(v)

	switch t {
	case
		ContentTypeCheckbox,
		ContentTypeHeading,
		ContentTypeImg,
		ContentTypeLine,
		ContentTypeList,
		ContentTypeMention,
		ContentTypeRadiobutton,
		ContentTypeText:
		return &t, nil
	}

	return nil, fmt.Errorf("invalid argument. v=%v", v)
}

type ContentValue []byte

func NewContentValue(b []byte) (*ContentValue, error) {
	result := ContentValue(b)

	return &result, nil
}
