package service

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	dservice "app/domain/service"
	uerror "app/usecase/error"
	umodel "app/usecase/model"
	"fmt"

	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
)

type ResourceSearchIndexUsecase interface {
	Create(c context.Context, resourceID uuid.UUID, resourceType string, keyword string) error
	List(c context.Context, resourceTypes []string, freeword string, limit int, offset int) ([]umodel.ResourceSearchIndex, error)
	Update(c context.Context, resourceID uuid.UUID, resourceType string, keyword string) error
	Delete(c context.Context, resourceID uuid.UUID) error
}

type resourceSearchIndexUsecase struct {
	resourceSearchIndexService dservice.ResourceSearchIndexService
}

// Delete implements ResourceSearchIndexUsecase.
func (r *resourceSearchIndexUsecase) Delete(c context.Context, resourceID uuid.UUID) error {
	return r.resourceSearchIndexService.Delete(c, resourceID)
}

// Create implements ResourceSearchIndexUsecase.
func (r *resourceSearchIndexUsecase) Create(c context.Context, resourceID uuid.UUID, resourceType string, keyword string) error {
	dIndex, err := dfactory.NewResourceSearchIndex(resourceID.String(), resourceType, keyword)
	if err != nil {
		return uerror.NewInvalidParameter(fmt.Sprintf("failed to parse resource search index. id=%v", resourceID.String()), err)
	}

	if err := r.resourceSearchIndexService.Create(c, *dIndex); err != nil {
		return errors.Wrapf(err, "failed to create resource search index. id=%v", resourceID.String())
	}

	return nil
}

// List implements ResourceSearchIndexUsecase.
func (r *resourceSearchIndexUsecase) List(c context.Context, resourceTypes []string, freeword string, limit int, offset int) ([]umodel.ResourceSearchIndex, error) {
	dResourceTypes := []dmodel.Resource{}
	for _, resourceType := range resourceTypes {
		dResourceType, err := dmodel.NewResource(resourceType)
		if err != nil {
			return nil, uerror.NewInvalidParameter(fmt.Sprintf("failed to parse resource type. v=%v", resourceType), err)
		}

		dResourceTypes = append(dResourceTypes, *dResourceType)
	}

	dRange, err := dmodel.NewRange(limit, offset)
	if err != nil {
		return nil, uerror.NewInvalidParameter(fmt.Sprintf("failed to parse range. limit=%v offset=%v", limit, offset), err)
	}

	dIndexes, err := r.resourceSearchIndexService.List(c, dResourceTypes, freeword, *dRange)
	if err != nil {
		return nil, err
	}

	indexes := []umodel.ResourceSearchIndex{}
	for _, dIndex := range dIndexes {
		indexes = append(indexes, umodel.ResourceSearchIndex{
			ResourceID: dIndex.ResourceID,
			Type:       dIndex.Type.String(),
			Keyword:    dIndex.Keyword.String(),
		})
	}

	return indexes, nil
}

// Update implements ResourceSearchIndexUsecase.
func (r *resourceSearchIndexUsecase) Update(c context.Context, resourceID uuid.UUID, resourceType string, keyword string) error {
	dIndex, err := dfactory.NewResourceSearchIndex(resourceID.String(), resourceType, keyword)
	if err != nil {
		return uerror.NewInvalidParameter(fmt.Sprintf("failed to parse resource search index. id=%v", resourceID.String()), err)
	}

	if err := r.resourceSearchIndexService.Update(c, *dIndex); err != nil {
		return errors.Wrapf(err, "failed to update resource search index. id=%v", resourceID.String())
	}

	return nil
}

func NewResourceSearchIndexUsecase(i *do.Injector) (ResourceSearchIndexUsecase, error) {
	resourceSearchIndexService := do.MustInvoke[dservice.ResourceSearchIndexService](i)
	return &resourceSearchIndexUsecase{
		resourceSearchIndexService: resourceSearchIndexService,
	}, nil
}
