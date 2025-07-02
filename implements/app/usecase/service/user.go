package service

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	dservice "app/domain/service"
	uerror "app/usecase/error"
	umodel "app/usecase/model"
	"time"

	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
)

type UserUsecase interface {
	Get(c context.Context, id uuid.UUID) (*umodel.User, error)
	Save(c context.Context, subject string, email string, name string, issuer string, imageURL *string) (*uuid.UUID, error)
	ListInvite(c context.Context, userID uuid.UUID, limit int, offset int) ([]umodel.UserInvite, error)
	ReplyInvite(c context.Context, userID uuid.UUID, inviteID uuid.UUID, agree bool) error
}

type userUsecase struct {
	userService                dservice.UserService
	noteService                dservice.NoteService
	resourceSearchIndexService dservice.ResourceSearchIndexService
	inviteService              dservice.InviteService
	activityService            dservice.ActivityService
	communityService           dservice.CommunityService
	roleService                dservice.RoleService
	memberService              dservice.MemberService
}

// ReplyInvite implements UserUsecase.
func (u *userUsecase) ReplyInvite(c context.Context, userID uuid.UUID, inviteID uuid.UUID, agree bool) error {
	invite, err := u.inviteService.Get(c, inviteID)
	if err != nil {
		return err
	}

	if invite == nil {
		return uerror.NewNotFound("invite not found", nil)
	}

	role, err := u.roleService.Get(c, invite.RoleID)
	if err != nil {
		return err
	}

	if role == nil {
		return uerror.NewNotFound("role not found", nil)
	}

	if !agree {
		return u.inviteService.DeleteInvitedUser(c, inviteID, userID)
	}

	communityID, err := u.roleService.GetRelatedCommunity(c, role.ID)
	if err != nil {
		return err
	}

	if communityID == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	community, err := u.communityService.Get(c, *communityID)
	if err != nil {
		return err
	}

	if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	if member, err := u.memberService.GetByCommunityAndUser(c, community.ID, userID); err != nil {
		return err
	} else if member != nil {
		return uerror.NewAlreadyExists("member already exists", nil)
	}

	memberID := uuid.NewString()
	member, err := dfactory.NewMember(memberID, userID.String(), role.ID.String())
	if err != nil {
		return err
	}

	memberMention, err := dmodel.NewMention(community.ID.String(), dmodel.ResourceCommunity.String())
	if err != nil {
		return err
	}

	if err := u.memberService.Create(c, *member, *memberMention); err != nil {
		return err
	}

	if err := u.inviteService.DeleteInvitedUser(c, inviteID, userID); err != nil {
		return err
	}

	dActivity, err := dfactory.NewMemberActivity(time.Now(), memberID, member.ID.String(), dmodel.ResourceMember.String(), dmodel.OperationCreate.String())
	if err != nil {
		return errors.Wrapf(err, "failed to parse member activity. id=%v", memberID)
	}

	if err := u.activityService.SaveMemberActivity(c, *dActivity); err != nil {
		return errors.Wrapf(err, "failed to save member activity. id=%v", memberID)
	}

	return nil
}

// ListInvite implements UserUsecase.
func (u *userUsecase) ListInvite(c context.Context, userID uuid.UUID, limit int, offset int) ([]umodel.UserInvite, error) {
	invites, err := u.inviteService.ListByUser(c, userID, dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	uInvites := []umodel.UserInvite{}
	for _, invite := range invites {
		role, err := u.roleService.Get(c, invite.RoleID)
		if err != nil {
			return nil, err
		}

		if role == nil {
			return nil, uerror.NewNotFound("role not found", nil)
		}

		communityID, err := u.roleService.GetRelatedCommunity(c, role.ID)
		if err != nil {
			return nil, err
		}

		if communityID == nil {
			return nil, uerror.NewNotFound("community not found", nil)
		}

		community, err := u.communityService.Get(c, *communityID)
		if err != nil {
			return nil, err
		}

		if community == nil {
			return nil, uerror.NewNotFound("community not found", nil)
		}

		var message *string
		if invite.Message != nil {
			v := invite.Message.String()
			message = &v
		} else {
			message = nil
		}

		uInvites = append(uInvites, umodel.UserInvite{
			ID: invite.ID,
			Community: umodel.Community{
				ID:   community.ID,
				Name: community.Name.String(),
			},
			Role: umodel.Role{
				ID:     role.ID,
				Name:   role.Name.String(),
				Action: role.Action.Strings(),
			},
			At:      invite.At,
			Message: message,
		})
	}

	return uInvites, nil
}

// Get implements UserUsecase.
func (u *userUsecase) Get(c context.Context, id uuid.UUID) (*umodel.User, error) {
	user, err := u.userService.Get(c, id)

	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user. id=%v", id)
	}

	if user == nil {
		return nil, uerror.NewNotFound("user not found", err)
	}

	var imageURL *string
	if user.ImageURL == nil {
		imageURL = nil
	} else {
		imageURL = (*string)(user.ImageURL)
	}

	return &umodel.User{
		ID:       user.ID,
		Subject:  user.Subject.String(),
		Email:    user.Email.String(),
		Name:     user.Email.String(),
		ImageUrl: imageURL,
	}, nil
}

// Save implements UserUsecase.
func (u *userUsecase) Save(c context.Context, subject string, email string, name string, issuer string, imageURL *string) (*uuid.UUID, error) {
	dSubject, err := dmodel.NewSubject(subject)
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse subject", err)
	}

	user, err := u.userService.GetBySubject(c, *dSubject)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user. subject=%v", subject)
	}

	var userID uuid.UUID

	if user == nil {
		userID = uuid.New()
		profileID := uuid.New().String()

		dNote, err := dfactory.NewNote(profileID)
		if err != nil {
			return nil, err
		}

		mention, err := dmodel.NewMention(profileID, dmodel.ResourceUser.String())
		if err != nil {
			return nil, err
		}

		if err := u.noteService.Create(c, *dNote, *mention); err != nil {
			return nil, err
		}

		dIndex, err := dfactory.NewResourceSearchIndex(userID.String(), dmodel.ResourceUser.String(), name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse resource search index. id=%v", userID.String())
		}

		if err := u.resourceSearchIndexService.Create(c, *dIndex); err != nil {
			return nil, errors.Wrapf(err, "failed to create resource search index. id=%v", userID.String())
		}
	} else {
		userID = user.ID

		dIndex, err := dfactory.NewResourceSearchIndex(userID.String(), dmodel.ResourceUser.String(), name)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse resource search index. id=%v", userID.String())
		}

		if err := u.resourceSearchIndexService.Update(c, *dIndex); err != nil {
			return nil, errors.Wrapf(err, "failed to update resource search index. id=%v", userID.String())
		}
	}

	saveUser, err := dfactory.NewUser(userID.String(), subject, email, issuer, name, imageURL)

	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse user", err)
	}

	if err := u.userService.Save(c, *saveUser); err != nil {
		return nil, errors.Wrapf(err, "failed to save user. id=%v", userID.String())
	}

	return &userID, nil
}

func NewUserUsecase(i *do.Injector) (UserUsecase, error) {
	userService := do.MustInvoke[dservice.UserService](i)
	noteService := do.MustInvoke[dservice.NoteService](i)
	resourceSearchIndexService := do.MustInvoke[dservice.ResourceSearchIndexService](i)
	activityService := do.MustInvoke[dservice.ActivityService](i)
	inviteService := do.MustInvoke[dservice.InviteService](i)
	communityService := do.MustInvoke[dservice.CommunityService](i)
	roleService := do.MustInvoke[dservice.RoleService](i)
	memberService := do.MustInvoke[dservice.MemberService](i)
	return &userUsecase{
		userService:                userService,
		noteService:                noteService,
		resourceSearchIndexService: resourceSearchIndexService,
		activityService:            activityService,
		inviteService:              inviteService,
		communityService:           communityService,
		roleService:                roleService,
		memberService:              memberService,
	}, nil
}
