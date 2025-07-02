package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type CommunityService interface {
	Create(c context.Context, community model.Community) error
	Get(c context.Context, id uuid.UUID) (*model.Community, error)
	Update(c context.Context, community model.Community) error
}

type communityService struct {
	communityRepository repository.CommunityRepository
}

// Update implements CommunityService.
func (co *communityService) Update(c context.Context, community model.Community) error {
	return co.communityRepository.Update(c, community)
}

// Get implements CommunityService.
func (co *communityService) Get(c context.Context, id uuid.UUID) (*model.Community, error) {
	return co.communityRepository.Get(c, id)
}

// Create implements CommunityService.
func (co *communityService) Create(c context.Context, community model.Community) error {
	return co.communityRepository.Create(c, community)
}

func NewCommunityService(i *do.Injector) (CommunityService, error) {
	communityRepository := do.MustInvoke[repository.CommunityRepository](i)
	return &communityService{communityRepository: communityRepository}, nil
}
