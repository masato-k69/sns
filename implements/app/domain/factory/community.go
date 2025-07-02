package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewCommunity(id string, name string, invitation bool) (*model.Community, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	parsedName, err := model.NewName(name)

	if err != nil {
		return nil, err
	}

	return &model.Community{
		ID:         parsedID,
		Name:       *parsedName,
		Invitation: invitation,
	}, nil
}
