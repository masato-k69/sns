package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type ResourceSearchIndexService interface {
	Create(c context.Context, index model.ResourceSearchIndex) error
	List(c context.Context, resourceTypes []model.Resource, freeword string, page model.Range) ([]model.ResourceSearchIndex, error)
	Update(c context.Context, index model.ResourceSearchIndex) error
	Delete(c context.Context, id uuid.UUID) error
}

type resourceSearchIndexService struct {
	resourceSearchIndexRepository repository.ResourceSearchIndexRepository
}

// Delete implements ResourceSearchIndexService.
func (r *resourceSearchIndexService) Delete(c context.Context, id uuid.UUID) error {
	return r.resourceSearchIndexRepository.Delete(c, id)
}

// Update implements ResourceSearchIndexService.
func (r *resourceSearchIndexService) Update(c context.Context, index model.ResourceSearchIndex) error {
	return r.resourceSearchIndexRepository.Update(c, index)
}

// Create implements ResourceSearchIndexService.
func (r *resourceSearchIndexService) Create(c context.Context, index model.ResourceSearchIndex) error {
	return r.resourceSearchIndexRepository.Create(c, index)
}

// List implements ResourceSearchIndexService.
func (r *resourceSearchIndexService) List(c context.Context, resourceTypes []model.Resource, freeword string, page model.Range) ([]model.ResourceSearchIndex, error) {
	return r.resourceSearchIndexRepository.List(c, resourceTypes, freeword, page)
}

func NewResourceSearchIndexService(i *do.Injector) (ResourceSearchIndexService, error) {
	resourceSearchIndexRepository := do.MustInvoke[repository.ResourceSearchIndexRepository](i)
	return &resourceSearchIndexService{resourceSearchIndexRepository: resourceSearchIndexRepository}, nil
}
