package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewRole(id string, name string, action map[string][]string) (*model.Role, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	parsedName, err := model.NewName(name)

	if err != nil {
		return nil, err
	}

	parsedAction, err := model.NewAction(action)

	if err != nil {
		return nil, err
	}

	return &model.Role{
		ID:     parsedID,
		Name:   *parsedName,
		Action: *parsedAction,
	}, nil
}
