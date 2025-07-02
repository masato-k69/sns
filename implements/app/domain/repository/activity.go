package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type ActivityRepository interface {
	SaveUserLoginActivity(c context.Context, activity model.UserLoginActivity) error
	SaveMemberActivity(c context.Context, activity model.MemberActivity) error
	SaveMemberLikeActivity(c context.Context, activity model.MemberLikeActivity) error
	ListMemberLikeActivity(c context.Context, target model.Mention, like bool, page model.Range) ([]model.MemberLikeActivity, error)
	CountMemberLikeActivity(c context.Context, target model.Mention, like bool) (*int, error)
	ListUserLoginActivity(c context.Context, userID uuid.UUID, page model.Range) ([]model.UserLoginActivity, error)
	ListMembersActivity(c context.Context, memberIDs []uuid.UUID, page model.Range) ([]model.MemberActivity, error)
	ListMembersLikeActivity(c context.Context, memberIDs []uuid.UUID, page model.Range) ([]model.MemberLikeActivity, error)
	ListRecentMemberActivity(c context.Context, memberID uuid.UUID, page model.Range) ([]model.MemberActivity, error)
}
