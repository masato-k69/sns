package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type RoleRepository interface {
	Create(c context.Context, role model.Role, mention model.Mention) error
	Get(c context.Context, id uuid.UUID) (*model.Role, error)
	GetDefaultByCommunity(c context.Context, communityID uuid.UUID) (*model.Role, error)
	GetRelatedCommunity(c context.Context, id uuid.UUID) (*uuid.UUID, error)
	List(c context.Context, ids []uuid.UUID) ([]model.Role, error)
	ListByCommunity(c context.Context, communityID uuid.UUID) ([]model.Role, error)
	Update(c context.Context, role model.Role) error
	Delete(c context.Context, id uuid.UUID) error
}
