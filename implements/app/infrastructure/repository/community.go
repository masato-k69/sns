package repository

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	drepository "app/domain/repository"
	irdb "app/infrastructure/adapter/datastore/rdb"
	imodel "app/infrastructure/model"
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"

	"github.com/samber/do"
	"gorm.io/gorm"
)

type communityRepository struct {
	communityStoreConnection irdb.CommunityStoreConnection
}

// Update implements repository.CommunityRepository.
func (co *communityRepository) Update(c context.Context, community dmodel.Community) error {
	return co.communityStoreConnection.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Save(&imodel.Community{
				ID:   community.ID.String(),
				Name: community.Name.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to update community. id=%v", community.ID.String())
		}

		return nil
	})
}

// Get implements repository.CommunityRepository.
func (co *communityRepository) Get(c context.Context, id uuid.UUID) (*dmodel.Community, error) {
	return co.get(c, id.String())
}

// Create implements repository.CommunityRepository.
func (co *communityRepository) Create(c context.Context, community dmodel.Community) error {
	return co.communityStoreConnection.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Create(&imodel.Community{
				ID:   community.ID.String(),
				Name: community.Name.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create community. id=%v", community.ID.String())
		}

		return nil
	})
}

func (co *communityRepository) get(_ context.Context, id string) (*dmodel.Community, error) {
	community := imodel.Community{ID: id}
	if err := co.communityStoreConnection.Read().
		First(&community).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get community. id=%v", id)
	}

	return dfactory.NewCommunity(community.ID, community.Name, community.Invitation)
}

func NewCommunityRepository(i *do.Injector) (drepository.CommunityRepository, error) {
	communityStoreConnection := do.MustInvoke[irdb.CommunityStoreConnection](i)
	return &communityRepository{
		communityStoreConnection: communityStoreConnection,
	}, nil
}
