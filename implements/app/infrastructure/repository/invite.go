package repository

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	drepository "app/domain/repository"
	irdb "app/infrastructure/adapter/datastore/rdb"
	imodel "app/infrastructure/model"
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"github.com/samber/lo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type inviteRepository struct {
	inviteStoreConnectionRDB irdb.InviteStoreConnection
}

// GetByRoleAndUser implements repository.InviteRepository.
func (i *inviteRepository) GetByRoleAndUser(c context.Context, roleID uuid.UUID, userID uuid.UUID) (*dmodel.Invite, error) {
	invite := imodel.Invite{}
	if err := i.inviteStoreConnectionRDB.Read().
		Model(&imodel.Invite{}).
		Select("invites.id as id, invites.role_id as role_id, invites.message as message, invites.at as at").
		Joins("inner join invited_users on invites.id = invited_users.invite_id").
		Where("invites.role_id = ?", roleID.String()).
		Where("invited_users.user_id = ?", userID.String()).
		Scan(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get invite. role_id=%v user_id=%v", roleID.String(), userID.String())
	}

	invitedUsers := []imodel.InvitedUser{}
	if err := i.inviteStoreConnectionRDB.Read().
		Where("invite_id = ?", invite.ID).
		Find(&invitedUsers).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list invited users. id=%v", invite.ID)
	}

	invitedUserIds := lo.Map(invitedUsers, func(iu imodel.InvitedUser, _ int) string { return iu.UserID })

	var message *string
	if invite.Message.Valid {
		message = &invite.Message.String
	} else {
		message = nil
	}

	dInvite, err := dfactory.NewInvite(invite.ID, invite.RoleID, message, invite.At, invitedUserIds)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse invite. id=%v", invite.ID)
	}

	return dInvite, nil
}

// Get implements repository.InviteRepository.
func (i *inviteRepository) Get(c context.Context, id uuid.UUID) (*dmodel.Invite, error) {
	invite := imodel.Invite{ID: id.String()}
	if err := i.inviteStoreConnectionRDB.Read().
		First(&invite).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get invite. id=%v", id.String())
	}

	invitedUsers := []imodel.InvitedUser{}
	if err := i.inviteStoreConnectionRDB.Read().
		Where("invite_id = ?", invite.ID).
		Find(&invitedUsers).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list invited users. id=%v", invite.ID)
	}

	invitedUserIds := lo.Map(invitedUsers, func(iu imodel.InvitedUser, _ int) string { return iu.UserID })

	var message *string
	if invite.Message.Valid {
		message = &invite.Message.String
	} else {
		message = nil
	}

	dInvite, err := dfactory.NewInvite(invite.ID, invite.RoleID, message, invite.At, invitedUserIds)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse invite. id=%v", invite.ID)
	}

	return dInvite, nil
}

// Create implements repository.InviteRepository.
func (i *inviteRepository) Create(c context.Context, invite dmodel.Invite) error {
	return i.inviteStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		var message sql.NullString
		if invite.Message != nil {
			message = sql.NullString{
				Valid:  true,
				String: invite.Message.String(),
			}
		} else {
			message = sql.NullString{
				Valid: false,
			}
		}

		if err := tx.
			Create(&imodel.Invite{
				ID:      invite.ID.String(),
				RoleID:  invite.RoleID.String(),
				Message: message,
				At:      invite.At,
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create invite. id=%v", invite.ID.String())
		}

		invitedUsers := lo.Map(invite.Users, func(userID uuid.UUID, _ int) imodel.InvitedUser {
			return imodel.InvitedUser{
				InviteID: invite.ID.String(),
				UserID:   userID.String(),
			}
		})

		if err := tx.
			Create(&invitedUsers).Error; err != nil {
			return errors.Wrapf(err, "failed to create invited users. id=%v", invite.ID.String())
		}

		return nil
	})
}

// Delete implements repository.InviteRepository.
func (i *inviteRepository) Delete(c context.Context, id uuid.UUID) error {
	return i.inviteStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Where("invite_id = ?", id.String()).
			Delete(&imodel.InvitedUser{}).Error; err != nil {
			return errors.Wrapf(err, "failed to delete invited users. id=%v", id.String())
		}

		if err := tx.
			Delete(&imodel.Invite{
				ID: id.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to delete invite. id=%v", id.String())
		}

		return nil
	})
}

// DeleteInvitedUser implements repository.InviteRepository.
func (i *inviteRepository) DeleteInvitedUser(c context.Context, id uuid.UUID, userID uuid.UUID) error {
	return i.inviteStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		invite := imodel.Invite{
			ID: id.String(),
		}

		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			First(&invite).Error; err != nil {
			return errors.Wrapf(err, "fialed to lock invite. id=%v", id.String())
		}

		invitedUser := imodel.InvitedUser{
			InviteID: id.String(),
			UserID:   userID.String(),
		}

		if err := tx.
			Delete(&invitedUser).Error; err != nil {
			return errors.Wrapf(err, "failed to delete invited user. id=%v user_id=%v", id.String(), userID.String())
		}

		if err := tx.
			Where("invite_id = ?", id.String()).
			First(&imodel.InvitedUser{}).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				if err := tx.
					Delete(&invite).Error; err != nil {
					return errors.Wrapf(err, "failed to delete invite. id=%v", id.String())
				}
			} else {
				return errors.Wrapf(err, "fialed to get invited user. id=%v user_id=%v", id.String(), userID.String())
			}
		}

		return nil
	})
}

// ListByRole implements repository.InviteRepository.
func (i *inviteRepository) ListByRole(c context.Context, roleIDs []uuid.UUID, page dmodel.Range) ([]dmodel.Invite, error) {
	invites := []imodel.Invite{}
	if err := i.inviteStoreConnectionRDB.Read().
		Where("role_id in ?", lo.Map(roleIDs, func(roleID uuid.UUID, _ int) string { return roleID.String() })).
		Order("created_at asc").
		Limit(page.Limit).Offset(page.Offset).
		Find(&invites).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list invite. ids=%v", roleIDs)
	}

	dInvites := []dmodel.Invite{}
	for _, invite := range invites {
		invitedUsers := []imodel.InvitedUser{}
		if err := i.inviteStoreConnectionRDB.Read().
			Where("invite_id = ?", invite.ID).
			Find(&invitedUsers).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to list invited users. id=%v", invite.ID)
		}

		invitedUserIds := lo.Map(invitedUsers, func(iu imodel.InvitedUser, _ int) string { return iu.UserID })

		var message *string
		if invite.Message.Valid {
			message = &invite.Message.String
		} else {
			message = nil
		}

		dInvite, err := dfactory.NewInvite(invite.ID, invite.RoleID, message, invite.At, invitedUserIds)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse invite. id=%v", invite.ID)
		}

		dInvites = append(dInvites, *dInvite)
	}

	return dInvites, nil
}

// ListByUser implements repository.InviteRepository.
func (i *inviteRepository) ListByUser(c context.Context, userID uuid.UUID, page dmodel.Range) ([]dmodel.Invite, error) {
	invitedMyselfs := []imodel.InvitedUser{}
	if err := i.inviteStoreConnectionRDB.Read().
		Where("user_id = ?", userID.String()).
		Order("created_at asc").
		Limit(page.Limit).Offset(page.Offset).
		Find(&invitedMyselfs).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list invited users. user_id=%v", userID.String())
	}

	invites := []imodel.Invite{}
	if err := i.inviteStoreConnectionRDB.Read().
		Where("id in ?", lo.Map(invitedMyselfs, func(iu imodel.InvitedUser, _ int) string { return iu.InviteID })).
		Find(&invites).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list invite. user_id=%v", userID.String())
	}

	dInvites := []dmodel.Invite{}
	for _, invite := range invites {
		invitedUsers := []imodel.InvitedUser{}
		if err := i.inviteStoreConnectionRDB.Read().
			Where("invite_id = ?", invite.ID).
			Find(&invitedUsers).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to list invited users. id=%v", invite.ID)
		}

		invitedUserIds := lo.Map(invitedUsers, func(iu imodel.InvitedUser, _ int) string { return iu.UserID })

		var message *string
		if invite.Message.Valid {
			message = &invite.Message.String
		} else {
			message = nil
		}

		dInvite, err := dfactory.NewInvite(invite.ID, invite.RoleID, message, invite.At, invitedUserIds)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse invite. id=%v", invite.ID)
		}

		dInvites = append(dInvites, *dInvite)
	}

	return dInvites, nil
}

func NewInviteRepository(i *do.Injector) (drepository.InviteRepository, error) {
	inviteStoreConnectionRDB := do.MustInvoke[irdb.InviteStoreConnection](i)
	return &inviteRepository{
		inviteStoreConnectionRDB: inviteStoreConnectionRDB,
	}, nil
}
