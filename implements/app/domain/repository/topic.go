package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type TopicRepository interface {
	Create(c context.Context, topic model.Topic, communityID uuid.UUID) error
	Get(c context.Context, id uuid.UUID) (*model.Topic, error)
	ListByCommunity(c context.Context, communityID uuid.UUID, page model.Range) ([]model.Topic, error)
}
