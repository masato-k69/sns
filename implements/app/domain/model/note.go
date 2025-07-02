package model

import (
	"fmt"

	"github.com/google/uuid"
)

type Note struct {
	ID uuid.UUID
}

type Line struct {
	ID       uuid.UUID
	NoteID   uuid.UUID
	Order    OrderNumber
	Property *LineProperty
}

type LineProperty struct {
	Type LinePropertyType
}

type LinePropertyType string

const (
	LinePropertyTypeToggle     LinePropertyType = "toggle"
	LinePropertyTypeBlockquota LinePropertyType = "blockquote"
	LinePropertyTypeCallout    LinePropertyType = "callout"
)

func (m *LinePropertyType) String() string {
	return string(*m)
}

func NewLinePropertyType(v string) (*LinePropertyType, error) {
	t := LinePropertyType(v)

	switch t {
	case
		LinePropertyTypeToggle,
		LinePropertyTypeBlockquota,
		LinePropertyTypeCallout:
		return &t, nil
	}

	return nil, fmt.Errorf("invalid argument. v=%v", v)
}
