package factory

import (
	"app/domain/model"
	"fmt"

	"github.com/google/uuid"
)

func NewContent(id string, contentType string, value []byte) (*model.Content, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, fmt.Errorf("invalid argument. v=%v", id)
	}

	parsedContentType, err := model.NewContentType(contentType)

	if err != nil {
		return nil, fmt.Errorf("invalid argument. v=%v", contentType)
	}

	parsedContentValue, err := model.NewContentValue(value)

	if err != nil {
		return nil, fmt.Errorf("invalid argument. v=%v", value)
	}

	return &model.Content{
		ID:    parsedID,
		Type:  *parsedContentType,
		Value: *parsedContentValue,
	}, nil
}
