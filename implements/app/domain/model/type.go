package model

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

type Subject string

func (m *Subject) String() string {
	return string(*m)
}

func NewSubject(v string) (*Subject, error) {
	result := Subject(v)
	return &result, nil
}

type EmailAddress string

func (m *EmailAddress) String() string {
	return string(*m)
}

func NewEmailAddress(v string) (*EmailAddress, error) {
	re, err := regexp.Compile(`^[a-zA-Z0-9_\-]+(\.[a-zA-Z0-9_\-]+)*@([a-zA-Z0-9][a-zA-Z0-9\-]*[a-zA-Z0-9]\.)+[a-zA-Z]{2,}$`)

	if err != nil {
		return nil, err
	}

	if !re.Match([]byte(v)) {
		return nil, fmt.Errorf("invalid argument. v=%v", v)
	}

	result := EmailAddress(v)
	return &result, nil
}

type Name string

func (m *Name) String() string {
	return string(*m)
}

func NewName(v string) (*Name, error) {
	if len(v) > 128 {
		return nil, fmt.Errorf("invalid argument. v=%v", v)
	}

	result := Name(v)
	return &result, nil
}

type URL string

func (m *URL) String() string {
	return string(*m)
}

func NewURL(v string) (*URL, error) {
	if len(v) > 4096 {
		return nil, fmt.Errorf("invalid argument. v=%v", v)
	}

	re, err := regexp.Compile(`^https?://[\w/:%#\$&\?\(\)~\.=\+\-]+$`)

	if err != nil {
		return nil, err
	}

	if !re.Match([]byte(v)) {
		return nil, fmt.Errorf("invalid argument. v=%v", v)
	}

	result := URL(v)
	return &result, nil
}

type OrderNumber int

func (m *OrderNumber) Int() int {
	return int(*m)
}

func NewOrderNumber(v int) (*OrderNumber, error) {
	if v < 1 {
		return nil, fmt.Errorf("invalid argument. v=%v", v)
	}

	result := OrderNumber(v)
	return &result, nil
}

type IPAddress string

func (m *IPAddress) String() string {
	return string(*m)
}

func NewIPAddress(v string) (*IPAddress, error) {
	reV4, err := regexp.Compile(`^((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])\.){3}(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])$`)

	if err != nil {
		return nil, err
	}

	reV4Match := reV4.Match([]byte(v))

	reV6, err := regexp.Compile(`^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))$`)

	if err != nil {
		return nil, err
	}

	reV6Match := reV6.Match([]byte(v))

	if !reV4Match && !reV6Match {
		return nil, fmt.Errorf("invalid argument. v=%v", v)
	}

	result := IPAddress(v)
	return &result, nil
}

type OperationgSystem string

func (m *OperationgSystem) String() string {
	return string(*m)
}

func NewOperationgSystem(v string) (*OperationgSystem, error) {
	result := OperationgSystem(strings.ReplaceAll(v, "\"", ""))
	return &result, nil
}

type UserAgent string

func (m *UserAgent) String() string {
	return string(*m)
}

func NewUserAgent(v string) (*UserAgent, error) {
	result := UserAgent(strings.ReplaceAll(v, "\"", ""))
	return &result, nil
}

type Text string

func (m *Text) String() string {
	return string(*m)
}

func NewText(v string) (*Text, error) {
	var t string
	if len(v) > 4096 {
		t = v[:4096]
	} else {
		t = v
	}

	result := Text(t)
	return &result, nil
}

type Range struct {
	Limit  int
	Offset int
}

func NewRange(limit int, offset int) (*Range, error) {
	if limit < 0 {
		return nil, fmt.Errorf("invalid argument. v=%v", limit)
	}

	if offset < 0 {
		return nil, fmt.Errorf("invalid argument. v=%v", offset)
	}

	return &Range{
		Limit:  limit,
		Offset: offset,
	}, nil
}

type Mention struct {
	ID       uuid.UUID
	Resource Resource
}

func NewMention(id string, resource string) (*Mention, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid argument. id=%v", id)
	}

	parsedResource, err := NewResource(resource)
	if err != nil {
		return nil, fmt.Errorf("invalid argument. resource=%v", resource)
	}

	return &Mention{
		ID:       parsedID,
		Resource: *parsedResource,
	}, nil
}

type UnixTime int

func (m *UnixTime) Int() int {
	return int(*m)
}

func NewUnixTime(v int) (*UnixTime, error) {
	if v < 0 {
		return nil, fmt.Errorf("invalid argument. v=%v", v)
	}

	if v > 9999999999 {
		return nil, fmt.Errorf("invalid argument. v=%v", v)
	}

	result := UnixTime(v)
	return &result, nil
}

type ShortMessage string

func (m *ShortMessage) String() string {
	return string(*m)
}

func NewShortMessage(v string) (*ShortMessage, error) {
	if len(v) > 140 {
		return nil, fmt.Errorf("invalid argument. id=%v", v)
	}

	result := ShortMessage(v)
	return &result, nil
}
