package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewTopic(id string, name string, created *string) (*model.Topic, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	parsedName, err := model.NewName(name)

	if err != nil {
		return nil, err
	}

	var dCreated *uuid.UUID
	if created != nil {
		parsedCreated, err := uuid.Parse(*created)
		if err != nil {
			return nil, err
		}

		dCreated = &parsedCreated
	}

	return &model.Topic{
		ID:      parsedID,
		Name:    *parsedName,
		Created: dCreated,
	}, nil
}
