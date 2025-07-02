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

type memberRepository struct {
	memberStoreConnectionRDB irdb.MemberStoreConnection
}

// GetJoinedCommunityID implements repository.MemberRepository.
func (m *memberRepository) GetJoinedCommunityID(c context.Context, memberID uuid.UUID) (*uuid.UUID, error) {
	relation := imodel.MemberCommunityRelation{}
	if err := m.memberStoreConnectionRDB.Read().
		Where("member_id = ?", memberID.String()).
		First(&relation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get community_relation. member_id=%v", memberID.String())
	}

	communityID, err := uuid.Parse(relation.CommunityID)
	if err != nil {
		return nil, err
	}

	return &communityID, nil
}

// ListByUser implements repository.MemberRepository.
func (m *memberRepository) ListByUser(c context.Context, userID uuid.UUID) ([]dmodel.Member, error) {
	members := []imodel.Member{}
	if err := m.memberStoreConnectionRDB.Read().
		Where("user_id = ?", userID.String()).
		Find(&members).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list member. user_id=%v", userID.String())
	}

	dMembers := []dmodel.Member{}
	for _, member := range members {
		dMember, err := dfactory.NewMember(member.ID, member.UserID, member.RoleID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse member. id=%v", member.ID)
		}

		dMembers = append(dMembers, *dMember)
	}

	return dMembers, nil
}

// GetByCommunityAndUser implements repository.MemberRepository.
func (m *memberRepository) GetByCommunityAndUser(c context.Context, communityID uuid.UUID, userID uuid.UUID) (*dmodel.Member, error) {
	member := imodel.Member{}
	if err := m.memberStoreConnectionRDB.Read().
		Model(&imodel.Member{}).
		Select("members.id as id, members.user_id as user_id, members.role_id as role_id").
		Joins("inner join member_community_relations on members.id = member_community_relations.member_id").
		Where("member_community_relations.community_id = ?", communityID.String()).
		Where("members.user_id = ?", userID.String()).
		Scan(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to list member. community_id=%v", communityID.String())
	}

	dMember, err := dfactory.NewMember(member.ID, member.UserID, member.RoleID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse member. id=%v", member.ID)
	}

	return dMember, nil
}

// Get implements repository.MemberRepository.
func (m *memberRepository) Get(c context.Context, id uuid.UUID) (*dmodel.Member, error) {
	member := imodel.Member{ID: id.String()}

	if err := m.memberStoreConnectionRDB.Read().
		First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get member. id=%v", id.String())
	}

	return dfactory.NewMember(member.RoleID, member.UserID, member.RoleID)
}

// ListByCommunity implements repository.MemberRepository.
func (m *memberRepository) ListByCommunity(c context.Context, communityID uuid.UUID, page dmodel.Range) ([]dmodel.Member, error) {
	members := []imodel.Member{}
	if err := m.memberStoreConnectionRDB.Read().
		Model(&imodel.Member{}).
		Select("members.id as id, members.user_id as user_id, members.role_id as role_id").
		Joins("inner join member_community_relations on members.id = member_community_relations.member_id").
		Where("member_community_relations.community_id = ?", communityID.String()).
		Order("members.created_at asc").
		Limit(page.Limit).Offset(page.Offset).
		Scan(&members).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list member. community_id=%v", communityID.String())
	}

	dMembers := []dmodel.Member{}
	for _, member := range members {
		dMember, err := dfactory.NewMember(member.ID, member.UserID, member.RoleID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse member. id=%v", member.ID)
		}

		dMembers = append(dMembers, *dMember)
	}

	return dMembers, nil
}

// Create implements repository.MemberRepository.
func (m *memberRepository) Create(c context.Context, member dmodel.Member, mention dmodel.Mention) error {
	return m.memberStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Create(&imodel.Member{
				ID:     member.ID.String(),
				UserID: member.UserID.String(),
				RoleID: member.RoleID.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create member. id=%v", member.ID.String())
		}

		switch mention.Resource {
		case dmodel.ResourceCommunity:
			if err := tx.
				Create(&imodel.MemberCommunityRelation{
					MemberID:    member.ID.String(),
					CommunityID: mention.ID.String(),
				}).Error; err != nil {
				return errors.Wrapf(err, "failed to create community_relation. id=%v", member.ID.String())
			}
		}

		return nil

	})
}

func NewMemberRepository(i *do.Injector) (drepository.MemberRepository, error) {
	memberStoreConnectionRDB := do.MustInvoke[irdb.MemberStoreConnection](i)
	return &memberRepository{
		memberStoreConnectionRDB: memberStoreConnectionRDB,
	}, nil
}
