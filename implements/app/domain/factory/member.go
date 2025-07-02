package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewMember(id string, userID string, roleID string) (*model.Member, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	parsedUserID, err := uuid.Parse(userID)

	if err != nil {
		return nil, err
	}

	parsedRoleID, err := uuid.Parse(roleID)

	if err != nil {
		return nil, err
	}

	return &model.Member{
		ID:     parsedID,
		UserID: parsedUserID,
		RoleID: parsedRoleID,
	}, nil
}
