package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Save(c context.Context, user model.User) error
	Get(c context.Context, id uuid.UUID) (*model.User, error)
	GetBySubject(c context.Context, subject model.Subject) (*model.User, error)
	List(c context.Context, ids []uuid.UUID) ([]model.User, error)
}
