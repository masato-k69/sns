package service

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	dservice "app/domain/service"
	uerror "app/usecase/error"
	umodel "app/usecase/model"
	"fmt"
	"time"

	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"github.com/samber/lo"
)

type ActivityUsecase interface {
	SaveUserLoginActivity(c context.Context, at time.Time, userID string, ipAddress string, operationSystem string, userAgent string) error
	SaveMemberActivity(c context.Context, at time.Time, member string, target string, resource string, operation string) error
	SaveMemberLikeActivity(c context.Context, at time.Time, member string, target string, resource string, like bool, comment *string) error
	ListUserLoginActivity(c context.Context, userID uuid.UUID, limit int, offset int) ([]umodel.Login, error)
	ListUsersMemberActivity(c context.Context, userID uuid.UUID, limit int, offset int) ([]umodel.Activity, error)
	ListUsersMemberLikeActivity(c context.Context, userID uuid.UUID, limit int, offset int) ([]umodel.Like, error)
	ListRecentMemberActivity(c context.Context, memberID uuid.UUID) ([]umodel.Activity, error)
}

type activityUsecase struct {
	roleService                dservice.RoleService
	memberService              dservice.MemberService
	noteService                dservice.NoteService
	communityService           dservice.CommunityService
	resourceSearchIndexService dservice.ResourceSearchIndexService
	activityService            dservice.ActivityService
	inviteService              dservice.InviteService
	userService                dservice.UserService
	topicService               dservice.TopicService
	threadService              dservice.ThreadService
	postService                dservice.PostService
	contentService             dservice.ContentService
}

// ListRecentMemberActivity implements ActivityUsecase.
func (a *activityUsecase) ListRecentMemberActivity(c context.Context, memberID uuid.UUID) ([]umodel.Activity, error) {
	dActivities, err := a.activityService.ListRecentMemberActivity(c, memberID, dmodel.Range{Limit: 10, Offset: 0})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list member activity. id=%v", memberID.String())
	}

	uActivities := []umodel.Activity{}
	for _, dActivity := range dActivities {
		uActivities = append(uActivities, umodel.Activity{
			At:        dActivity.At,
			Me:        dActivity.Member,
			Target:    dActivity.Target,
			Resource:  dActivity.Resource.String(),
			Operation: dActivity.Operation.String(),
		})
	}

	return uActivities, nil
}

// ListUsersMemberLikeActivity implements ActivityUsecase.
func (a *activityUsecase) ListUsersMemberLikeActivity(c context.Context, userID uuid.UUID, limit int, offset int) ([]umodel.Like, error) {
	dMembers, err := a.memberService.ListByUser(c, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list member. user_id=%v", userID.String())
	}

	dLikes, err := a.activityService.ListMembersLikeActivity(c, lo.Map(dMembers, func(member dmodel.Member, _ int) uuid.UUID { return member.ID }), dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	uLikes := []umodel.Like{}
	for _, dLike := range dLikes {
		by, err := func(c context.Context, memberID uuid.UUID) (*umodel.Member, error) {
			dMember, err := a.memberService.Get(c, dLike.Member)
			if err != nil {
				return nil, err
			} else if dMember == nil {
				return nil, nil
			}

			uMember, err := a.toMember(c, dMember)
			if err != nil {
				return nil, err
			}

			return uMember, nil
		}(c, dLike.Member)

		if err != nil {
			return nil, err
		}

		var comment *string
		if dLike.Comment != nil {
			v := dLike.Comment.String()
			comment = &v
		} else {
			comment = nil
		}

		uLikes = append(uLikes, umodel.Like{
			Like:    dLike.Like,
			Comment: comment,
			By:      by,
		})
	}

	return uLikes, nil
}

// ListUsersMemberActivity implements ActivityUsecase.
func (a *activityUsecase) ListUsersMemberActivity(c context.Context, userID uuid.UUID, limit int, offset int) ([]umodel.Activity, error) {
	dMembers, err := a.memberService.ListByUser(c, userID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list member. user_id=%v", userID.String())
	}

	if len(dMembers) < 1 {
		return []umodel.Activity{}, nil
	}

	dActivities, err := a.activityService.ListMembersActivity(c, lo.Map(dMembers, func(member dmodel.Member, _ int) uuid.UUID { return member.ID }), dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list member activity. user_id=%v", userID.String())
	}

	uActivities := []umodel.Activity{}
	for _, dActivity := range dActivities {
		uActivities = append(uActivities, umodel.Activity{
			At:        dActivity.At,
			Me:        dActivity.Member,
			Target:    dActivity.Target,
			Resource:  dActivity.Resource.String(),
			Operation: dActivity.Operation.String(),
		})
	}

	return uActivities, nil
}

// ListUserLoginActivity implements ActivityUsecase.
func (a *activityUsecase) ListUserLoginActivity(c context.Context, userID uuid.UUID, limit int, offset int) ([]umodel.Login, error) {
	dActivities, err := a.activityService.ListUserLoginActivity(c, userID, dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list user login activity. id=%v", userID.String())
	}

	uActivities := []umodel.Login{}
	for _, dActivity := range dActivities {
		uActivities = append(uActivities, umodel.Login{
			At:              dActivity.At,
			UserID:          dActivity.UserID,
			IPAddress:       dActivity.IPAddress.String(),
			OperationSystem: dActivity.OperationSystem.String(),
			UserAgent:       dActivity.UserAgent.String(),
		})
	}

	return uActivities, nil
}

// SaveMemberLikeActivity implements ActivityUsecase.
func (a *activityUsecase) SaveMemberLikeActivity(c context.Context, at time.Time, member string, target string, resource string, like bool, comment *string) error {
	dActivity, err := dfactory.NewMemberLikeActivity(at, member, target, resource, like, comment)

	if err != nil {
		return uerror.NewInvalidParameter("failed to parse member like activity", err)
	}

	if err := a.activityService.SaveMemberLikeActivity(c, *dActivity); err != nil {
		return errors.Wrapf(err, "failed to save member like activity. id=%v", member)
	}

	return nil
}

// SaveMemberActivity implements ActivityUsecase.
func (a *activityUsecase) SaveMemberActivity(c context.Context, at time.Time, member string, target string, resource string, operation string) error {
	dActivity, err := dfactory.NewMemberActivity(at, member, target, resource, operation)

	if err != nil {
		return uerror.NewInvalidParameter("failed to parse member activity", err)
	}

	if err := a.activityService.SaveMemberActivity(c, *dActivity); err != nil {
		return errors.Wrapf(err, "failed to save member activity. id=%v", member)
	}

	return nil
}

// SaveUserLoginActivity implements ActivityUsecase.
func (a *activityUsecase) SaveUserLoginActivity(c context.Context, at time.Time, userID string, ipAddress string, operationSystem string, userAgent string) error {
	dActivity, err := dfactory.NewUserLoginActivity(at, userID, ipAddress, operationSystem, userAgent)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse user login activity", err)
	}
	if err := a.activityService.SaveUserLoginActivity(c, *dActivity); err != nil {
		return errors.Wrapf(err, "failed to save user login activity. id=%v", userID)
	}
	return nil
}

func (a *activityUsecase) toMember(c context.Context, member *dmodel.Member) (*umodel.Member, error) {
	uUser, err := func(id uuid.UUID) (*umodel.User, error) {
		user, err := a.userService.Get(c, id)
		if err != nil {
			return nil, err
		} else if user == nil {
			return nil, uerror.NewNotFound(fmt.Sprintf("user not found. id=%v", id), nil)
		}

		var imageURL *string
		if user.ImageURL != nil {
			v := user.ImageURL.String()
			imageURL = &v
		} else {
			imageURL = nil
		}

		return &umodel.User{
			ID:       user.ID,
			Name:     user.Name.String(),
			ImageUrl: imageURL,
		}, nil
	}(member.UserID)

	if err != nil {
		return nil, err
	}

	uRole, err := func(id uuid.UUID) (*umodel.Role, error) {
		role, err := a.roleService.Get(c, id)
		if err != nil {
			return nil, err
		} else if role == nil {
			return nil, nil
		}

		return &umodel.Role{
			ID:     role.ID,
			Name:   role.Name.String(),
			Action: role.Action.Strings(),
		}, nil
	}(member.RoleID)

	if err != nil {
		return nil, err
	}

	return &umodel.Member{
		ID:   member.ID,
		User: *uUser,
		Role: uRole,
	}, nil
}

func NewActivityUsecase(i *do.Injector) (ActivityUsecase, error) {
	roleService := do.MustInvoke[dservice.RoleService](i)
	memberService := do.MustInvoke[dservice.MemberService](i)
	noteService := do.MustInvoke[dservice.NoteService](i)
	communityService := do.MustInvoke[dservice.CommunityService](i)
	resourceSearchIndexService := do.MustInvoke[dservice.ResourceSearchIndexService](i)
	activityService := do.MustInvoke[dservice.ActivityService](i)
	inviteService := do.MustInvoke[dservice.InviteService](i)
	userService := do.MustInvoke[dservice.UserService](i)
	topicService := do.MustInvoke[dservice.TopicService](i)
	threadService := do.MustInvoke[dservice.ThreadService](i)
	postService := do.MustInvoke[dservice.PostService](i)
	contentService := do.MustInvoke[dservice.ContentService](i)
	return &activityUsecase{
		roleService:                roleService,
		memberService:              memberService,
		noteService:                noteService,
		communityService:           communityService,
		resourceSearchIndexService: resourceSearchIndexService,
		activityService:            activityService,
		inviteService:              inviteService,
		userService:                userService,
		topicService:               topicService,
		threadService:              threadService,
		postService:                postService,
		contentService:             contentService,
	}, nil
}
