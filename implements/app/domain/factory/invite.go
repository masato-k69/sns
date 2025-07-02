package factory

import (
	"app/domain/model"
	"time"

	"github.com/google/uuid"
)

func NewInvite(id string, roleID string, message *string, at time.Time, users []string) (*model.Invite, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	parsedRoleID, err := uuid.Parse(roleID)

	if err != nil {
		return nil, err
	}

	var parsedMessage *model.Text
	if message != nil {
		parsedMessage, err = model.NewText(*message)

		if err != nil {
			return nil, err
		}
	} else {
		parsedMessage = nil
	}

	parsedUsers := []uuid.UUID{}
	for _, userID := range users {
		parsedUserID, err := uuid.Parse(userID)

		if err != nil {
			return nil, err
		}

		parsedUsers = append(parsedUsers, parsedUserID)
	}

	return &model.Invite{
		ID:      parsedID,
		RoleID:  parsedRoleID,
		Message: parsedMessage,
		At:      at,
		Users:   parsedUsers,
	}, nil
}
