package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type UserService interface {
	Save(c context.Context, user model.User) error
	Get(c context.Context, id uuid.UUID) (*model.User, error)
	GetBySubject(c context.Context, subject model.Subject) (*model.User, error)
	List(c context.Context, ids []uuid.UUID) ([]model.User, error)
}

type userService struct {
	userRepository repository.UserRepository
}

// List implements UserService.
func (u *userService) List(c context.Context, ids []uuid.UUID) ([]model.User, error) {
	return u.userRepository.List(c, ids)
}

// Get implements UserService.
func (u *userService) Get(c context.Context, id uuid.UUID) (*model.User, error) {
	return u.userRepository.Get(c, id)
}

// GetBySubject implements UserService.
func (u *userService) GetBySubject(c context.Context, subject model.Subject) (*model.User, error) {
	return u.userRepository.GetBySubject(c, subject)
}

// Save implements UserService.
func (u *userService) Save(c context.Context, user model.User) error {
	return u.userRepository.Save(c, user)
}

func NewUserService(i *do.Injector) (UserService, error) {
	userRepository := do.MustInvoke[repository.UserRepository](i)
	return &userService{userRepository: userRepository}, nil
}
