package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type InviteService interface {
	Create(c context.Context, invite model.Invite) error
	Get(c context.Context, id uuid.UUID) (*model.Invite, error)
	GetByRoleAndUser(c context.Context, roleID uuid.UUID, userID uuid.UUID) (*model.Invite, error)
	ListByRole(c context.Context, roleIDs []uuid.UUID, page model.Range) ([]model.Invite, error)
	ListByUser(c context.Context, userID uuid.UUID, page model.Range) ([]model.Invite, error)
	Delete(c context.Context, id uuid.UUID) error
	DeleteInvitedUser(c context.Context, id uuid.UUID, userID uuid.UUID) error
}

type inviteService struct {
	inviteRepository repository.InviteRepository
}

// GetByRoleAndUser implements InviteService.
func (i *inviteService) GetByRoleAndUser(c context.Context, roleID uuid.UUID, userID uuid.UUID) (*model.Invite, error) {
	return i.inviteRepository.GetByRoleAndUser(c, roleID, userID)
}

// Get implements InviteService.
func (i *inviteService) Get(c context.Context, id uuid.UUID) (*model.Invite, error) {
	return i.inviteRepository.Get(c, id)
}

// Create implements InviteService.
func (i *inviteService) Create(c context.Context, invite model.Invite) error {
	return i.inviteRepository.Create(c, invite)
}

// Delete implements InviteService.
func (i *inviteService) Delete(c context.Context, id uuid.UUID) error {
	return i.inviteRepository.Delete(c, id)
}

// DeleteInvitedUser implements InviteService.
func (i *inviteService) DeleteInvitedUser(c context.Context, id uuid.UUID, userID uuid.UUID) error {
	return i.inviteRepository.DeleteInvitedUser(c, id, userID)
}

// ListByRole implements InviteService.
func (i *inviteService) ListByRole(c context.Context, roleIDs []uuid.UUID, page model.Range) ([]model.Invite, error) {
	return i.inviteRepository.ListByRole(c, roleIDs, page)
}

// ListByUser implements InviteService.
func (i *inviteService) ListByUser(c context.Context, userID uuid.UUID, page model.Range) ([]model.Invite, error) {
	return i.inviteRepository.ListByUser(c, userID, page)
}

func NewInviteService(i *do.Injector) (InviteService, error) {
	inviteRepository := do.MustInvoke[repository.InviteRepository](i)
	return &inviteService{inviteRepository: inviteRepository}, nil
}
