package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type CommunityRepository interface {
	Create(c context.Context, community model.Community) error
	Get(c context.Context, id uuid.UUID) (*model.Community, error)
	Update(c context.Context, community model.Community) error
}
