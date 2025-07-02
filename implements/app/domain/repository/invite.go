package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type InviteRepository interface {
	Create(c context.Context, invite model.Invite) error
	Get(c context.Context, id uuid.UUID) (*model.Invite, error)
	GetByRoleAndUser(c context.Context, roleID uuid.UUID, userID uuid.UUID) (*model.Invite, error)
	ListByRole(c context.Context, roleIDs []uuid.UUID, page model.Range) ([]model.Invite, error)
	ListByUser(c context.Context, userID uuid.UUID, page model.Range) ([]model.Invite, error)
	Delete(c context.Context, id uuid.UUID) error
	DeleteInvitedUser(c context.Context, id uuid.UUID, userID uuid.UUID) error
}
