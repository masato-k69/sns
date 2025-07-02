package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type ActivityService interface {
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

type activityService struct {
	activityRepository repository.ActivityRepository
}

// ListRecentMemberActivity implements ActivityService.
func (a *activityService) ListRecentMemberActivity(c context.Context, memberID uuid.UUID, page model.Range) ([]model.MemberActivity, error) {
	return a.activityRepository.ListRecentMemberActivity(c, memberID, page)
}

// ListMembersLikeActivity implements ActivityService.
func (a *activityService) ListMembersLikeActivity(c context.Context, memberIDs []uuid.UUID, page model.Range) ([]model.MemberLikeActivity, error) {
	return a.activityRepository.ListMembersLikeActivity(c, memberIDs, page)
}

// ListMembersActivity implements ActivityService.
func (a *activityService) ListMembersActivity(c context.Context, memberIDs []uuid.UUID, page model.Range) ([]model.MemberActivity, error) {
	return a.activityRepository.ListMembersActivity(c, memberIDs, page)
}

// ListUserLoginActivity implements ActivityService.
func (a *activityService) ListUserLoginActivity(c context.Context, userID uuid.UUID, page model.Range) ([]model.UserLoginActivity, error) {
	return a.activityRepository.ListUserLoginActivity(c, userID, page)
}

// ListMemberLikeActivity implements ActivityService.
func (a *activityService) ListMemberLikeActivity(c context.Context, target model.Mention, like bool, page model.Range) ([]model.MemberLikeActivity, error) {
	return a.activityRepository.ListMemberLikeActivity(c, target, like, page)
}

// CountMemberLikeActivity implements ActivityService.
func (a *activityService) CountMemberLikeActivity(c context.Context, target model.Mention, like bool) (*int, error) {
	return a.activityRepository.CountMemberLikeActivity(c, target, like)
}

// SaveMemberLikeActivity implements ActivityService.
func (a *activityService) SaveMemberLikeActivity(c context.Context, activity model.MemberLikeActivity) error {
	return a.activityRepository.SaveMemberLikeActivity(c, activity)
}

// SaveMemberActivity implements ActivityService.
func (a *activityService) SaveMemberActivity(c context.Context, activity model.MemberActivity) error {
	return a.activityRepository.SaveMemberActivity(c, activity)
}

// SaveUserLoginActivity implements ActivityService.
func (a *activityService) SaveUserLoginActivity(c context.Context, activity model.UserLoginActivity) error {
	return a.activityRepository.SaveUserLoginActivity(c, activity)
}

func NewActivityService(i *do.Injector) (ActivityService, error) {
	activityRepository := do.MustInvoke[repository.ActivityRepository](i)
	return &activityService{activityRepository: activityRepository}, nil
}
