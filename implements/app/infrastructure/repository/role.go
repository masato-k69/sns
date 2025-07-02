package repository

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	drepository "app/domain/repository"
	idocument "app/infrastructure/adapter/datastore/document"
	irdb "app/infrastructure/adapter/datastore/rdb"
	imodel "app/infrastructure/model"
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

type roleRepository struct {
	roleStoreConnectionRDB      irdb.RoleStoreConnection
	roleStoreConnectionDocument idocument.RoleStoreConnection
}

// GetDefaultByCommunity implements repository.RoleRepository.
func (r *roleRepository) GetDefaultByCommunity(c context.Context, communityID uuid.UUID) (*dmodel.Role, error) {
	communityRelation := imodel.RoleCommunityRelation{}
	if err := r.roleStoreConnectionRDB.Write().
		Where("community_id = ?", communityID.String()).
		Where("default = ?", true).
		First(&communityRelation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get community role relations. community_id=%v", communityID.String())
	}

	role := imodel.Role{ID: communityRelation.RoleID}
	if err := r.roleStoreConnectionRDB.Read().
		First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get role. id=%v", communityRelation.RoleID)
	}

	action := imodel.Action{
		RoleID: role.ID,
	}

	option, err := idocument.Unmarshal(&action)
	if err != nil {
		return nil, err
	}

	if err := r.roleStoreConnectionDocument.DB().
		Collection(action.Collection()).
		FindOne(c, option).
		Decode(&action); err != nil {
		return nil, errors.Wrapf(err, "failed to get action. role_id=%v", communityRelation.RoleID)
	}

	actions := map[string][]string{}
	for _, actionItem := range action.Items {
		actions[actionItem.Resource] = actionItem.Operation
	}

	return dfactory.NewRole(role.ID, role.Name, actions)
}

// GetRelatedCommunity implements repository.RoleRepository.
func (r *roleRepository) GetRelatedCommunity(c context.Context, id uuid.UUID) (*uuid.UUID, error) {
	communityRelation := imodel.RoleCommunityRelation{}
	if err := r.roleStoreConnectionRDB.Write().
		Where("role_id = ?", id.String()).
		First(&communityRelation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get community role relations. role_id=%v", id.String())
	}

	communityID, err := uuid.Parse(communityRelation.CommunityID)
	if err != nil {
		return nil, err
	}

	return &communityID, nil
}

// ListByCommunity implements repository.RoleRepository.
func (r *roleRepository) ListByCommunity(c context.Context, communityID uuid.UUID) ([]dmodel.Role, error) {
	roles := []imodel.Role{}
	if err := r.roleStoreConnectionRDB.Read().
		Model(&imodel.Role{}).
		Select("roles.id as id, roles.name as name").
		Joins("inner join role_community_relations on roles.id = role_community_relations.role_id").
		Where("role_community_relations.community_id = ?", communityID.String()).
		Order("roles.created_at asc").
		Scan(&roles).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get role. community_id=%v", communityID.String())
	}

	dRoles := []dmodel.Role{}
	for _, role := range roles {
		action := imodel.Action{
			RoleID: role.ID,
		}

		option, err := idocument.Unmarshal(&action)
		if err != nil {
			return nil, err
		}

		if err := r.roleStoreConnectionDocument.DB().
			Collection(action.Collection()).
			FindOne(c, option).
			Decode(&action); err != nil {
			return nil, errors.Wrapf(err, "failed to get action. role_id=%v", role.ID)
		}

		actions := map[string][]string{}
		for _, actionItem := range action.Items {
			actions[actionItem.Resource] = actionItem.Operation
		}

		dRole, err := dfactory.NewRole(role.ID, role.Name, actions)
		if err != nil {
			return nil, err
		}

		dRoles = append(dRoles, *dRole)
	}

	return dRoles, nil
}

// Delete implements repository.RoleRepository.
func (r *roleRepository) Delete(c context.Context, id uuid.UUID) error {
	return r.roleStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Delete(&imodel.Role{
				ID: id.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to delete role. id=%v", id.String())
		}

		action := &imodel.Action{
			RoleID: id.String(),
		}

		if _, err := r.roleStoreConnectionDocument.DB().
			Collection(action.Collection()).
			DeleteOne(c, action.CollectionKey()); err != nil {
			return errors.Wrapf(err, "failed to delete action. role_id=%v", id.String())
		}

		return nil
	})
}

// Update implements repository.RoleRepository.
func (r *roleRepository) Update(c context.Context, role dmodel.Role) error {
	return r.roleStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Updates(&imodel.Role{
				ID:   role.ID.String(),
				Name: role.Name.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to update role. id=%v", role.ID.String())
		}

		actionItem := []imodel.ActionItem{}
		for resource, operation := range role.Action.Strings() {
			actionItem = append(actionItem, imodel.ActionItem{
				Resource:  resource,
				Operation: operation,
			})
		}

		action := &imodel.Action{
			RoleID: role.ID.String(),
			Items:  actionItem,
		}

		if _, err := r.roleStoreConnectionDocument.DB().
			Collection(action.Collection()).
			DeleteOne(c, action.CollectionKey()); err != nil {
			return errors.Wrapf(err, "failed to delete action. role_id=%v", role.ID.String())
		}

		if _, err := r.roleStoreConnectionDocument.DB().
			Collection(action.Collection()).
			InsertOne(c, &action); err != nil {
			return errors.Wrapf(err, "failed to create action. role_id=%v", role.ID.String())
		}

		return nil
	})
}

// List implements repository.RoleRepository.
func (r *roleRepository) List(c context.Context, ids []uuid.UUID) ([]dmodel.Role, error) {
	roles := []imodel.Role{}
	if err := r.roleStoreConnectionRDB.Read().
		Where("id in ?", lo.Map(ids, func(id uuid.UUID, _ int) string { return id.String() })).
		Find(&roles).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get role. ids=%v", ids)
	}

	dRoles := []dmodel.Role{}
	for _, role := range roles {
		action := imodel.Action{
			RoleID: role.ID,
		}

		option, err := idocument.Unmarshal(&action)
		if err != nil {
			return nil, err
		}

		if err := r.roleStoreConnectionDocument.DB().
			Collection(action.Collection()).
			FindOne(c, option).
			Decode(&action); err != nil {
			return nil, errors.Wrapf(err, "failed to get action. role_id=%v", role.ID)
		}

		actions := map[string][]string{}
		for _, actionItem := range action.Items {
			actions[actionItem.Resource] = actionItem.Operation
		}

		dRole, err := dfactory.NewRole(role.ID, role.Name, actions)
		if err != nil {
			return nil, err
		}

		dRoles = append(dRoles, *dRole)
	}

	return dRoles, nil
}

// Get implements repository.RoleRepository.
func (r *roleRepository) Get(c context.Context, id uuid.UUID) (*dmodel.Role, error) {
	role := imodel.Role{ID: id.String()}
	if err := r.roleStoreConnectionRDB.Read().
		First(&role).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get role. id=%v", id.String())
	}

	action := imodel.Action{
		RoleID: role.ID,
	}

	option, err := idocument.Unmarshal(&action)
	if err != nil {
		return nil, err
	}

	if err := r.roleStoreConnectionDocument.DB().
		Collection(action.Collection()).
		FindOne(c, option).
		Decode(&action); err != nil {
		return nil, errors.Wrapf(err, "failed to get action. role_id=%v", id.String())
	}

	actions := map[string][]string{}
	for _, actionItem := range action.Items {
		actions[actionItem.Resource] = actionItem.Operation
	}

	return dfactory.NewRole(role.ID, role.Name, actions)
}

// Create implements repository.RoleRepository.
func (r *roleRepository) Create(c context.Context, role dmodel.Role, mention dmodel.Mention) error {
	return r.roleStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Create(&imodel.Role{
				ID:   role.ID.String(),
				Name: role.Name.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create role. id=%v", role.ID.String())
		}

		switch mention.Resource {
		case dmodel.ResourceCommunity:
			if err := tx.
				Create(&imodel.RoleCommunityRelation{
					RoleID:      role.ID.String(),
					CommunityID: mention.ID.String(),
				}).Error; err != nil {
				return errors.Wrapf(err, "failed to create community relation. id=%v", mention.ID.String())
			}
		}

		actionItem := []imodel.ActionItem{}
		for resource, operation := range role.Action.Strings() {
			actionItem = append(actionItem, imodel.ActionItem{
				Resource:  resource,
				Operation: operation,
			})
		}

		action := &imodel.Action{
			RoleID: role.ID.String(),
			Items:  actionItem,
		}

		if _, err := r.roleStoreConnectionDocument.DB().
			Collection(action.Collection()).
			InsertOne(c, &action); err != nil {
			return errors.Wrapf(err, "failed to create action. role_id=%v", role.ID.String())
		}

		return nil
	})
}

func NewRoleRepository(i *do.Injector) (drepository.RoleRepository, error) {
	roleStoreConnectionRDB := do.MustInvoke[irdb.RoleStoreConnection](i)
	roleStoreConnectionDocument := do.MustInvoke[idocument.RoleStoreConnection](i)
	return &roleRepository{
		roleStoreConnectionRDB:      roleStoreConnectionRDB,
		roleStoreConnectionDocument: roleStoreConnectionDocument,
	}, nil
}
