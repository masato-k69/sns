package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type ResourceSearchIndexRepository interface {
	Create(c context.Context, index model.ResourceSearchIndex) error
	List(c context.Context, resourceTypes []model.Resource, freeword string, page model.Range) ([]model.ResourceSearchIndex, error)
	Update(c context.Context, index model.ResourceSearchIndex) error
	Delete(c context.Context, id uuid.UUID) error
}
