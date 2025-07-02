package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type ThreadRepository interface {
	Create(c context.Context, thread model.Thread, topicID uuid.UUID) error
	Get(c context.Context, id uuid.UUID) (*model.Thread, error)
	ListByTopic(c context.Context, topicID uuid.UUID, page model.Range) ([]model.Thread, error)
}
