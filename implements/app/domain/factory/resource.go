package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewResourceSearchIndex(resourceID string, resource string, keyword string) (*model.ResourceSearchIndex, error) {
	parsedResourceID, err := uuid.Parse(resourceID)
	if err != nil {
		return nil, err
	}

	parsedResource, err := model.NewResource(resource)
	if err != nil {
		return nil, err
	}

	parsedText, err := model.NewText(keyword)
	if err != nil {
		return nil, err
	}

	return &model.ResourceSearchIndex{
		ResourceID: parsedResourceID,
		Type:       *parsedResource,
		Keyword:    *parsedText,
	}, nil
}
