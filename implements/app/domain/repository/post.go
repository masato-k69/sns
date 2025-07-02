package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type PostRepository interface {
	Create(c context.Context, post model.Post, topicID uuid.UUID, threadID uuid.UUID) error
	Get(c context.Context, id uuid.UUID) (*model.Post, error)
	Last(c context.Context, topicID uuid.UUID) (*model.Post, error)
	ListByThread(c context.Context, threadID uuid.UUID, page model.Range) ([]model.Post, error)
}
