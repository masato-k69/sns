package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewThread(id string) (*model.Thread, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	return &model.Thread{
		ID: parsedID,
	}, nil
}
