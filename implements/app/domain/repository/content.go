package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type ContentRepository interface {
	Create(c context.Context, contents []model.Content, mention model.Mention) error
	DeleteAndCreate(c context.Context, contents []model.Content, mention model.Mention) error
	List(c context.Context, ids []uuid.UUID) ([]model.Content, error)
	ListByTopic(c context.Context, topicID uuid.UUID) ([]model.Content, error)
	ListByLine(c context.Context, lineID uuid.UUID) ([]model.Content, error)
	ListByPost(c context.Context, postID uuid.UUID) ([]model.Content, error)
	DeleteByResource(c context.Context, mention model.Mention) error
}
