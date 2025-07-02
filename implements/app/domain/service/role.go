package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type RoleService interface {
	Create(c context.Context, role model.Role, mention model.Mention) error
	Get(c context.Context, id uuid.UUID) (*model.Role, error)
	GetDefaultByCommunity(c context.Context, communityID uuid.UUID) (*model.Role, error)
	GetRelatedCommunity(c context.Context, id uuid.UUID) (*uuid.UUID, error)
	List(c context.Context, ids []uuid.UUID) ([]model.Role, error)
	ListByCommunity(c context.Context, communityID uuid.UUID) ([]model.Role, error)
	Update(c context.Context, role model.Role) error
	Delete(c context.Context, id uuid.UUID) error
}

type roleService struct {
	roleRepository repository.RoleRepository
}

// GetDefaultByCommunity implements RoleService.
func (r *roleService) GetDefaultByCommunity(c context.Context, communityID uuid.UUID) (*model.Role, error) {
	return r.roleRepository.GetDefaultByCommunity(c, communityID)
}

// GetRelatedCommunity implements RoleService.
func (r *roleService) GetRelatedCommunity(c context.Context, id uuid.UUID) (*uuid.UUID, error) {
	return r.roleRepository.GetRelatedCommunity(c, id)
}

// ListByCommunity implements RoleService.
func (r *roleService) ListByCommunity(c context.Context, communityID uuid.UUID) ([]model.Role, error) {
	return r.roleRepository.ListByCommunity(c, communityID)
}

// Delete implements RoleService.
func (r *roleService) Delete(c context.Context, id uuid.UUID) error {
	return r.roleRepository.Delete(c, id)
}

// Update implements RoleService.
func (r *roleService) Update(c context.Context, role model.Role) error {
	return r.roleRepository.Update(c, role)
}

// List implements RoleService.
func (r *roleService) List(c context.Context, ids []uuid.UUID) ([]model.Role, error) {
	return r.roleRepository.List(c, ids)
}

// Get implements RoleService.
func (r *roleService) Get(c context.Context, id uuid.UUID) (*model.Role, error) {
	return r.roleRepository.Get(c, id)
}

// Create implements RoleService.
func (r *roleService) Create(c context.Context, role model.Role, mention model.Mention) error {
	return r.roleRepository.Create(c, role, mention)
}

func NewRoleService(i *do.Injector) (RoleService, error) {
	roleRepository := do.MustInvoke[repository.RoleRepository](i)
	return &roleService{roleRepository: roleRepository}, nil
}
