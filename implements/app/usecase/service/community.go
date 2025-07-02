package service

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	dservice "app/domain/service"
	uerror "app/usecase/error"
	umodel "app/usecase/model"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"github.com/samber/lo"
)

type CommunityUsecase interface {
	Create(c context.Context, userID uuid.UUID, name string, invitation bool) (*uuid.UUID, error)
	Get(c context.Context, id uuid.UUID) (*umodel.Community, error)
	GetByMember(c context.Context, memberID uuid.UUID) (*umodel.Community, error)
	CanUpdate(c context.Context, id uuid.UUID, userID uuid.UUID) (*umodel.Community, error)
	Update(c context.Context, id uuid.UUID, userID uuid.UUID, name string) error
	CreateRole(c context.Context, communityID uuid.UUID, userID uuid.UUID, name string, action map[string][]string) error
	ListRole(c context.Context, communityID uuid.UUID) ([]umodel.Role, error)
	UpdateRole(c context.Context, communityID uuid.UUID, userID uuid.UUID, roleID uuid.UUID, name string, action map[string][]string) error
	DeleteRole(c context.Context, communityID uuid.UUID, userID uuid.UUID, roleID uuid.UUID) error
	Invite(c context.Context, communityID uuid.UUID, userID uuid.UUID, roleID uuid.UUID, mention []uuid.UUID, message *string) error
	ListInvite(c context.Context, communityID uuid.UUID, limit int, offset int) ([]umodel.CommunityInvite, error)
	DeleteInvite(c context.Context, communityID uuid.UUID, userID uuid.UUID, inviteID uuid.UUID) error
	ListMember(c context.Context, communityID uuid.UUID, limit int, offset int) ([]umodel.Member, error)
	CreateTopic(c context.Context, communityID uuid.UUID, userID uuid.UUID, name string, contents []umodel.Content) (*uuid.UUID, error)
	ListTopic(c context.Context, communityID uuid.UUID, limit int, offset int) ([]umodel.Topic, error)
	Post(c context.Context, communityID uuid.UUID, userID uuid.UUID, topicID uuid.UUID, contents []umodel.Content, mention []umodel.Mention, searchWord string) error
	Reply(c context.Context, communityID uuid.UUID, userID uuid.UUID, topicID uuid.UUID, threadID uuid.UUID, contents []umodel.Content, mention []umodel.Mention, searchWord string) error
	ListThread(c context.Context, communityID uuid.UUID, topicID uuid.UUID, limit int, offset int) ([]umodel.Thread, error)
	ListPost(c context.Context, communityID uuid.UUID, topicID uuid.UUID, threadID uuid.UUID, limit int, offset int) ([]umodel.Post, error)
	LikePost(c context.Context, communityID uuid.UUID, topicID uuid.UUID, threadID uuid.UUID, postID uuid.UUID, userID uuid.UUID, like bool, comment *string) error
	ListPostLike(c context.Context, communityID uuid.UUID, topicID uuid.UUID, threadID uuid.UUID, postID uuid.UUID, like bool, limit int, offset int) ([]umodel.Like, error)
}

type communityUsecase struct {
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

// GetByMember implements CommunityUsecase.
func (co *communityUsecase) GetByMember(c context.Context, memberID uuid.UUID) (*umodel.Community, error) {
	communityID, err := co.memberService.GetJoinedCommunityID(c, memberID)
	if err != nil {
		return nil, err
	} else if communityID == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("joined community not found. id=%v", memberID.String()), nil)
	}

	community, _, err := co.get(c, *communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("community not found. id=%v", memberID.String()), nil)
	}

	return &umodel.Community{
		ID:   community.ID,
		Name: community.Name.String(),
	}, nil
}

// ListPostLike implements CommunityUsecase.
func (co *communityUsecase) ListPostLike(c context.Context, communityID uuid.UUID, topicID uuid.UUID, threadID uuid.UUID, postID uuid.UUID, like bool, limit int, offset int) ([]umodel.Like, error) {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("community not found. id=%v", communityID.String()), nil)
	}

	dMention, err := dmodel.NewMention(postID.String(), dmodel.ResourcePost.String())
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse mention", err)
	}

	dLikes, err := co.activityService.ListMemberLikeActivity(c, *dMention, like, dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	uLikes := []umodel.Like{}
	for _, dLike := range dLikes {
		by, err := func(c context.Context, memberID uuid.UUID, roles []dmodel.Role) (*umodel.Member, error) {
			dMember, err := co.memberService.Get(c, dLike.Member)
			if err != nil {
				return nil, err
			} else if dMember == nil {
				return nil, nil
			}

			uMember, err := co.toMember(c, &roles, nil, dMember)
			if err != nil {
				return nil, err
			}

			return uMember, nil
		}(c, dLike.Member, roles)

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

// Like implements CommunityUsecase.
func (co *communityUsecase) LikePost(c context.Context, communityID uuid.UUID, topicID uuid.UUID, threadID uuid.UUID, postID uuid.UUID, userID uuid.UUID, like bool, comment *string) error {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return err
	} else if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	myMember, _, err := co.getMymemberAndRole(c, communityID, userID, roles)
	if err != nil {
		return err
	}

	dActivity, err := dfactory.NewMemberLikeActivity(time.Now(), myMember.ID.String(), postID.String(), dmodel.ResourcePost.String(), like, comment)

	if err != nil {
		return uerror.NewInvalidParameter("failed to parse member like activity", err)
	}

	if err := co.activityService.SaveMemberLikeActivity(c, *dActivity); err != nil {
		return errors.Wrapf(err, "failed to save member like activity. id=%v", myMember.ID.String())
	}

	return nil
}

// CreateTopic implements CommunityUsecase.
func (co *communityUsecase) CreateTopic(c context.Context, communityID uuid.UUID, userID uuid.UUID, name string, contents []umodel.Content) (*uuid.UUID, error) {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound("community not found", nil)
	}

	var myMemberID *string
	myMember, myRole, err := co.getMymemberAndRole(c, communityID, userID, roles)
	if err != nil {
		return nil, err
	} else if !myRole.CanCreate(dmodel.ResourceTopic) {
		return nil, uerror.NewNewPermissionDenied("cannot create", nil)
	} else {
		v := myMember.ID.String()
		myMemberID = &v
	}

	newTopicID := uuid.New()
	newTopic, err := dfactory.NewTopic(newTopicID.String(), name, myMemberID)
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse topic", err)
	}

	if err := co.topicService.Create(c, *newTopic, communityID); err != nil {
		return nil, errors.Wrapf(err, "failed to create topic. community_id=%v member_id=%v name=%v", communityID.String(), myMember.ID.String(), name)
	}

	newContents := []dmodel.Content{}
	for _, content := range contents {
		newContentID := uuid.New()
		newContent, err := dfactory.NewContent(newContentID.String(), content.Type, content.Bin)
		if err != nil {
			return nil, uerror.NewInvalidParameter(fmt.Sprintf("failed to parse content. type=%v", content.Type), err)
		}

		newContents = append(newContents, *newContent)
	}

	newMention, err := dmodel.NewMention(newTopicID.String(), dmodel.ResourceTopic.String())
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse mention", err)
	}

	if err := co.contentService.Create(c, newContents, *newMention); err != nil {
		return nil, errors.Wrapf(err, "failed to create content. community_id=%v member_id=%v name=%v", communityID.String(), myMember.ID.String(), name)
	}

	if err := co.saveIndexAndActivity(c, myMember.ID, newTopicID, name, dmodel.ResourceTopic, dmodel.OperationCreate); err != nil {
		return nil, err
	}

	return &newTopicID, nil
}

// ListPost implements CommunityUsecase.
func (co *communityUsecase) ListPost(c context.Context, communityID uuid.UUID, topicID uuid.UUID, threadID uuid.UUID, limit int, offset int) ([]umodel.Post, error) {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("community not found. id=%v", communityID.String()), nil)
	}

	dPosts, err := co.postService.ListByThread(c, threadID, dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	uPosts := []umodel.Post{}
	for _, dPost := range dPosts {
		uPost, err := co.toPost(c, dPost, roles)
		if err != nil {
			return nil, err
		}

		uPosts = append(uPosts, *uPost)
	}

	return uPosts, nil
}

// ListThread implements CommunityUsecase.
func (co *communityUsecase) ListThread(c context.Context, communityID uuid.UUID, topicID uuid.UUID, limit int, offset int) ([]umodel.Thread, error) {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("community not found. id=%v", communityID.String()), nil)
	}

	dThreads, err := co.threadService.ListByTopic(c, topicID, dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	uThreads := []umodel.Thread{}
	for _, dThread := range dThreads {
		dPosts, err := co.postService.ListByThread(c, dThread.ID, dmodel.Range{Limit: 2, Offset: 0})
		if err != nil {
			return nil, err
		}

		uPosts := []umodel.Post{}
		for _, dPost := range dPosts {
			uPost, err := co.toPost(c, dPost, roles)
			if err != nil {
				return nil, err
			}

			uPosts = append(uPosts, *uPost)
		}

		uThreads = append(uThreads, umodel.Thread{
			ID:    dThread.ID,
			Posts: uPosts,
		})
	}

	return uThreads, nil
}

// ListTopic implements CommunityUsecase.
func (co *communityUsecase) ListTopic(c context.Context, communityID uuid.UUID, limit int, offset int) ([]umodel.Topic, error) {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("community not found. id=%v", communityID.String()), nil)
	}

	dTopics, err := co.topicService.ListByCommunity(c, communityID, dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	uTopics := []umodel.Topic{}
	for _, dTopic := range dTopics {
		dContents, err := co.contentService.ListByTopic(c, dTopic.ID)
		if err != nil {
			return nil, err
		}

		uContents := lo.Map(dContents, func(dContent dmodel.Content, _ int) umodel.Content {
			return umodel.Content{
				Type: dContent.Type.String(),
				Bin:  dContent.Value,
			}
		})

		var uPost *umodel.Post
		if dLastPost, err := co.postService.Last(c, dTopic.ID); err != nil {
			return nil, err
		} else if dLastPost != nil {
			uPost, err = co.toPost(c, *dLastPost, roles)
			if err != nil {
				return nil, err
			}
		}

		var uCreated *umodel.Member
		if dTopic.Created != nil {
			dMember, err := co.memberService.Get(c, *dTopic.Created)
			if err != nil {
				return nil, err
			} else if dMember == nil {
				return nil, uerror.NewNotFound(fmt.Sprintf("member not found. id=%v", *dTopic.Created), nil)
			}

			created, err := co.toMember(c, &roles, nil, dMember)
			if err != nil {
				return nil, err
			}

			uCreated = created
		}

		uTopics = append(uTopics, umodel.Topic{
			ID:       dTopic.ID,
			Name:     dTopic.Name.String(),
			Contents: uContents,
			Created:  uCreated,
			LastPost: uPost,
		})
	}

	return uTopics, nil
}

// Post implements CommunityUsecase.
func (co *communityUsecase) Post(c context.Context, communityID uuid.UUID, userID uuid.UUID, topicID uuid.UUID, contents []umodel.Content, mention []umodel.Mention, searchWord string) error {
	return co.post(c, communityID, userID, topicID, nil, contents, mention, searchWord)
}

// Reply implements CommunityUsecase.
func (co *communityUsecase) Reply(c context.Context, communityID uuid.UUID, userID uuid.UUID, topicID uuid.UUID, threadID uuid.UUID, contents []umodel.Content, mention []umodel.Mention, searchWord string) error {
	return co.post(c, communityID, userID, topicID, &threadID, contents, mention, searchWord)
}

// ListMember implements CommunityUsecase.
func (co *communityUsecase) ListMember(c context.Context, communityID uuid.UUID, limit int, offset int) ([]umodel.Member, error) {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound("community not found", nil)
	}

	members, err := co.memberService.ListByCommunity(c, communityID, dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	users, err := co.userService.List(c, lo.Map(members, func(member dmodel.Member, _ int) uuid.UUID { return member.UserID }))
	if err != nil {
		return nil, err
	}

	uMembers := []umodel.Member{}
	for _, member := range members {
		uMember, err := co.toMember(c, &roles, &users, &member)
		if err != nil {
			return nil, err
		}

		uMembers = append(uMembers, *uMember)
	}

	return uMembers, nil
}

// DeleteInvite implements CommunityUsecase.
func (co *communityUsecase) DeleteInvite(c context.Context, communityID uuid.UUID, userID uuid.UUID, inviteID uuid.UUID) error {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return err
	} else if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	if _, myRole, err := co.getMymemberAndRole(c, communityID, userID, roles); err != nil {
		return err
	} else if !myRole.CanCreate(dmodel.ResourceMember) {
		return uerror.NewNewPermissionDenied("cannot de;ete", nil)
	}

	return co.inviteService.Delete(c, inviteID)
}

// ListInvite implements CommunityUsecase.
func (co *communityUsecase) ListInvite(c context.Context, communityID uuid.UUID, limit int, offset int) ([]umodel.CommunityInvite, error) {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound("community not found", nil)
	}

	invites, err := co.inviteService.ListByRole(c, lo.Map(roles, func(role dmodel.Role, _ int) uuid.UUID { return role.ID }), dmodel.Range{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}

	uInvites := []umodel.CommunityInvite{}
	for _, invite := range invites {
		role, ok := lo.Find(roles, func(role dmodel.Role) bool { return role.ID == invite.RoleID })
		if !ok {
			return nil, uerror.NewNotFound("role not found", nil)
		}

		users, err := co.userService.List(c, invite.Users)
		if err != nil {
			return nil, err
		}

		var message *string
		if invite.Message != nil {
			v := invite.Message.String()
			message = &v
		} else {
			message = nil
		}

		uInvites = append(uInvites, umodel.CommunityInvite{
			ID: invite.ID,
			Role: umodel.Role{
				ID:     role.ID,
				Name:   role.Name.String(),
				Action: role.Action.Strings(),
			},
			Users: lo.Map(users, func(user dmodel.User, _ int) umodel.User {
				var imageURL *string
				if user.ImageURL != nil {
					v := user.ImageURL.String()
					imageURL = &v
				} else {
					imageURL = nil
				}

				return umodel.User{
					ID:       user.ID,
					Subject:  user.Subject.String(),
					Email:    user.Email.String(),
					Name:     user.Name.String(),
					ImageUrl: imageURL,
				}
			}),
			At:      invite.At,
			Message: message,
		})
	}

	return uInvites, nil
}

// Invite implements CommunityUsecase.
func (co *communityUsecase) Invite(c context.Context, communityID uuid.UUID, userID uuid.UUID, roleID uuid.UUID, mention []uuid.UUID, message *string) error {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return err
	} else if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	if _, ok := lo.Find(roles, func(role dmodel.Role) bool { return role.ID.String() == roleID.String() }); !ok {
		return uerror.NewNotFound("role not found", nil)
	}

	if _, myRole, err := co.getMymemberAndRole(c, communityID, userID, roles); err != nil {
		return err
	} else if !myRole.CanCreate(dmodel.ResourceMember) {
		return uerror.NewNewPermissionDenied("cannot create", nil)
	}

	for _, mentionedUserID := range mention {
		if invite, err := co.inviteService.GetByRoleAndUser(c, roleID, mentionedUserID); err != nil {
			return err
		} else if invite != nil {
			return uerror.NewAlreadyExists(fmt.Sprintf("invite already exists. user_id=%v", mentionedUserID.String()), nil)
		}
	}

	if member, err := co.memberService.GetByCommunityAndUser(c, communityID, userID); err != nil {
		return err
	} else if member != nil {
		return uerror.NewAlreadyExists(fmt.Sprintf("member already exists. user_id=%v", userID.String()), nil)
	}

	inviteID := uuid.NewString()
	dInvite, err := dfactory.NewInvite(inviteID, roleID.String(), message, time.Now(),
		lo.Map(mention, func(userID uuid.UUID, _ int) string { return userID.String() }))
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse invite", err)
	}

	if err := co.inviteService.Create(c, *dInvite); err != nil {
		return errors.Wrapf(err, "failed to create invite. community_id=%v role_id=%v", communityID.String(), roleID.String())
	}

	return nil
}

// ListRole implements CommunityUsecase.
func (co *communityUsecase) ListRole(c context.Context, communityID uuid.UUID) ([]umodel.Role, error) {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound("community not found", nil)
	}

	uRoles := lo.Map(roles, func(role dmodel.Role, _ int) umodel.Role {
		return umodel.Role{
			ID:     role.ID,
			Name:   role.Name.String(),
			Action: role.Action.Strings(),
		}
	})

	return uRoles, nil
}

// DeleteRole implements CommunityUsecase.
func (co *communityUsecase) DeleteRole(c context.Context, communityID uuid.UUID, userID uuid.UUID, roleID uuid.UUID) error {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return err
	} else if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	if _, ok := lo.Find(roles, func(role dmodel.Role) bool { return role.ID.String() == roleID.String() }); !ok {
		return uerror.NewNotFound("role not found", nil)
	}

	myMember, myRole, err := co.getMymemberAndRole(c, communityID, userID, roles)
	if err != nil {
		return err
	} else if !myRole.CanDelete(dmodel.ResourceRole) {
		return uerror.NewNewPermissionDenied("cannot delete", nil)
	}

	if err := co.roleService.Delete(c, roleID); err != nil {
		return errors.Wrapf(err, "failed to delete role. id=%v", roleID.String())
	}

	if err := co.deleteIndexAndSaveActivity(c, myMember.ID, roleID, dmodel.ResourceRole); err != nil {
		return err
	}

	return nil
}

// UpdateRole implements CommunityUsecase.
func (co *communityUsecase) UpdateRole(c context.Context, communityID uuid.UUID, userID uuid.UUID, roleID uuid.UUID, name string, action map[string][]string) error {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return err
	} else if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	if _, ok := lo.Find(roles, func(role dmodel.Role) bool { return role.ID.String() == roleID.String() }); !ok {
		return uerror.NewNotFound("role not found", nil)
	}

	myMember, myRole, err := co.getMymemberAndRole(c, communityID, userID, roles)
	if err != nil {
		return err
	}

	if !myRole.CanUpdate(dmodel.ResourceRole) {
		return uerror.NewNewPermissionDenied("cannot update", nil)
	}

	dRole, err := dfactory.NewRole(roleID.String(), name, action)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse role", err)
	}

	if err := co.roleService.Update(c, *dRole); err != nil {
		return errors.Wrapf(err, "failed to update role. id=%v", dRole.ID.String())
	}

	if err := co.saveIndexAndActivity(c, myMember.ID, roleID, name, dmodel.ResourceRole, dmodel.OperationUpdate); err != nil {
		return err
	}

	return nil
}

// CreateRole implements CommunityUsecase.
func (co *communityUsecase) CreateRole(c context.Context, communityID uuid.UUID, userID uuid.UUID, name string, action map[string][]string) error {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return err
	} else if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	myMember, myRole, err := co.getMymemberAndRole(c, communityID, userID, roles)
	if err != nil {
		return err
	}

	if !myRole.CanCreate(dmodel.ResourceRole) {
		return uerror.NewNewPermissionDenied("cannot create", nil)
	}

	roleID := uuid.New()
	role, err := dfactory.NewRole(roleID.String(), name, action)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse role", err)
	}

	roleMention, err := dmodel.NewMention(communityID.String(), dmodel.ResourceCommunity.String())
	if err != nil {
		return err
	}

	if err := co.roleService.Create(c, *role, *roleMention); err != nil {
		return errors.Wrapf(err, "failed to create role. id=%v", role.ID.String())
	}

	if err := co.saveIndexAndActivity(c, myMember.ID, roleID, name, dmodel.ResourceRole, dmodel.OperationCreate); err != nil {
		return err
	}

	return nil
}

// Get implements CommunityUsecase.
func (co *communityUsecase) Get(c context.Context, id uuid.UUID) (*umodel.Community, error) {
	community, _, err := co.get(c, id)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound("community not found", nil)
	}

	return &umodel.Community{
		ID:   community.ID,
		Name: community.Name.String(),
	}, nil
}

// CanUpdate implements CommunityUsecase.
func (co *communityUsecase) CanUpdate(c context.Context, id uuid.UUID, userID uuid.UUID) (*umodel.Community, error) {
	community, roles, err := co.get(c, id)
	if err != nil {
		return nil, err
	} else if community == nil {
		return nil, uerror.NewNotFound("community not found", nil)
	}

	_, myRole, err := co.getMymemberAndRole(c, id, userID, roles)
	if err != nil {
		return nil, err
	}

	if !myRole.CanUpdate(dmodel.ResourceCommunity) {
		return nil, uerror.NewNewPermissionDenied("cannot update", nil)
	}

	return &umodel.Community{
		ID:   community.ID,
		Name: community.Name.String(),
	}, nil
}

// Update implements CommunityUsecase.
func (co *communityUsecase) Update(c context.Context, id uuid.UUID, userID uuid.UUID, name string) error {
	parsedName, err := dmodel.NewName(name)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse name", err)
	}

	community, roles, err := co.get(c, id)
	if err != nil {
		return err
	} else if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	myMember, myRole, err := co.getMymemberAndRole(c, id, userID, roles)
	if err != nil {
		return err
	}

	if !myRole.CanUpdate(dmodel.ResourceCommunity) {
		return uerror.NewNewPermissionDenied("cannot update", nil)
	}

	community.Name = *parsedName
	if err := co.communityService.Update(c, *community); err != nil {

	}

	if err := co.saveIndexAndActivity(c, myMember.ID, community.ID, name, dmodel.ResourceCommunity, dmodel.OperationUpdate); err != nil {
		return err
	}

	return nil
}

// Create implements CommunityUsecase.
func (co *communityUsecase) Create(c context.Context, userID uuid.UUID, name string, invitation bool) (*uuid.UUID, error) {
	communityID := uuid.New()
	community, err := dfactory.NewCommunity(communityID.String(), name, invitation)
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse community", err)
	}

	if err := co.communityService.Create(c, *community); err != nil {
		return nil, errors.Wrapf(err, "failed to create community. id=%v", community.ID.String())
	}

	ownerRoleID := uuid.New()
	ownerAction := dmodel.CommunityMemberAction
	ownerRole, err := dfactory.NewRole(ownerRoleID.String(), name, ownerAction.Strings())
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse role", err)
	}

	communityMention, err := dmodel.NewMention(communityID.String(), dmodel.ResourceCommunity.String())
	if err != nil {
		return nil, err
	}

	if err := co.roleService.Create(c, *ownerRole, *communityMention); err != nil {
		return nil, errors.Wrapf(err, "failed to create role. id=%v", ownerRole.ID.String())
	}

	ownerID := uuid.New()
	owner, err := dfactory.NewMember(ownerID.String(), userID.String(), ownerRoleID.String())
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse member", err)
	}

	if err := co.memberService.Create(c, *owner, *communityMention); err != nil {
		return nil, errors.Wrapf(err, "failed to create member. id=%v", owner.ID.String())
	}

	descriptionID := uuid.NewString()
	description, err := dfactory.NewNote(descriptionID)
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse description", err)
	}

	if err := co.noteService.Create(c, *description, *communityMention); err != nil {
		return nil, errors.Wrapf(err, "failed to create description. id=%v", description.ID.String())
	}

	if err := co.saveIndexAndActivity(c, ownerID, communityID, name, dmodel.ResourceCommunity, dmodel.OperationCreate); err != nil {
		return nil, err
	}

	return &community.ID, nil
}

func (co *communityUsecase) get(c context.Context, id uuid.UUID) (*dmodel.Community, []dmodel.Role, error) {
	community, err := co.communityService.Get(c, id)
	if err != nil {
		return nil, nil, err
	} else if community == nil {
		return nil, nil, nil
	}

	roles, err := co.roleService.ListByCommunity(c, community.ID)
	if err != nil {
		return nil, nil, err
	}

	return community, roles, nil
}

func (co *communityUsecase) getMymemberAndRole(c context.Context, communityID uuid.UUID, userID uuid.UUID, roles []dmodel.Role) (*dmodel.Member, *dmodel.Role, error) {
	member, err := co.memberService.GetByCommunityAndUser(c, communityID, userID)
	if err != nil {
		return nil, nil, err
	} else if member == nil {
		return nil, nil, uerror.NewNewPermissionDenied("member not found", nil)
	}

	role, ok := lo.Find(roles, func(role dmodel.Role) bool { return role.ID.String() == member.RoleID.String() })
	if !ok {
		return nil, nil, uerror.NewNewPermissionDenied("role not found", nil)
	}

	return member, &role, nil
}

func (co *communityUsecase) toMember(c context.Context, roles *[]dmodel.Role, users *[]dmodel.User, member *dmodel.Member) (*umodel.Member, error) {
	uUser, err := func(users *[]dmodel.User, id uuid.UUID) (*umodel.User, error) {
		var user dmodel.User

		if users != nil {
			var ok bool
			user, ok = lo.Find(*users, func(user dmodel.User) bool { return user.ID == id })
			if !ok {
				return nil, uerror.NewNotFound(fmt.Sprintf("user not found. id=%v", id), nil)
			}
		} else {
			user, err := co.userService.Get(c, id)
			if err != nil {
				return nil, err
			} else if user == nil {
				return nil, uerror.NewNotFound(fmt.Sprintf("user not found. id=%v", id), nil)
			}
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
	}(users, member.UserID)

	if err != nil {
		return nil, err
	}

	uRole, err := func(roles *[]dmodel.Role, id uuid.UUID) (*umodel.Role, error) {
		var role dmodel.Role

		if roles != nil {
			var ok bool
			role, ok = lo.Find(*roles, func(role dmodel.Role) bool { return role.ID == id })
			if !ok {
				return nil, nil
			}
		} else {
			role, err := co.roleService.Get(c, id)
			if err != nil {
				return nil, err
			} else if role == nil {
				return nil, nil
			}
		}

		return &umodel.Role{
			ID:     role.ID,
			Name:   role.Name.String(),
			Action: role.Action.Strings(),
		}, nil
	}(roles, member.RoleID)

	if err != nil {
		return nil, err
	}

	return &umodel.Member{
		ID:   member.ID,
		User: *uUser,
		Role: uRole,
	}, nil
}

func (co *communityUsecase) toPost(c context.Context, post dmodel.Post, roles []dmodel.Role) (*umodel.Post, error) {
	dContents, err := co.contentService.ListByPost(c, post.ID)
	if err != nil {
		return nil, err
	}

	uContents := lo.Map(dContents, func(dContent dmodel.Content, _ int) umodel.Content {
		return umodel.Content{
			Type: dContent.Type.String(),
			Bin:  dContent.Value,
		}
	})

	var uCreated *umodel.Member
	if post.From != nil {
		dMember, err := co.memberService.Get(c, *post.From)
		if err != nil {
			return nil, err
		} else if dMember == nil {
			return nil, uerror.NewNotFound(fmt.Sprintf("member not found. id=%v", post.From.String()), nil)
		}

		created, err := co.toMember(c, &roles, nil, dMember)
		if err != nil {
			return nil, err
		}

		uCreated = created
	}

	likeMention, err := dmodel.NewMention(post.ID.String(), dmodel.ResourcePost.String())
	if err != nil {
		return nil, err
	}

	likes, err := co.activityService.CountMemberLikeActivity(c, *likeMention, true)
	if err != nil {
		return nil, err
	}

	dislikes, err := co.activityService.CountMemberLikeActivity(c, *likeMention, false)
	if err != nil {
		return nil, err
	}

	return &umodel.Post{
		ID:       post.ID,
		At:       post.At.Int(),
		Contents: uContents,
		Created:  uCreated,
		Reaction: umodel.Reaction{
			Likes:    *likes,
			Dislikes: *dislikes,
		},
	}, nil
}

func (co *communityUsecase) post(c context.Context, communityID uuid.UUID, userID uuid.UUID, topicID uuid.UUID, threadID *uuid.UUID, contents []umodel.Content, mention []umodel.Mention, searchWord string) error {
	community, roles, err := co.get(c, communityID)
	if err != nil {
		return err
	} else if community == nil {
		return uerror.NewNotFound("community not found", nil)
	}

	var myMemberID *string
	myMember, myRole, err := co.getMymemberAndRole(c, communityID, userID, roles)
	if err != nil {
		return err
	} else if !myRole.CanCreate(dmodel.ResourcePost) {
		return uerror.NewNewPermissionDenied("cannot create", nil)
	} else {
		v := myMember.ID.String()
		myMemberID = &v
	}

	if threadID == nil {
		newThreadID := uuid.New()
		dThread, err := dfactory.NewThread(newThreadID.String())
		if err != nil {
			return uerror.NewInvalidParameter("failed to parse thread", err)
		}

		if err := co.threadService.Create(c, *dThread, topicID); err != nil {
			return errors.Wrapf(err, "failed to create thread. topic_id=%v", topicID.String())
		}

		threadID = &newThreadID
	}

	dMention := []dmodel.Mention{}
	for _, to := range mention {
		dTo, err := dmodel.NewMention(to.ID.String(), to.ResourceType)
		if err != nil {
			return uerror.NewInvalidParameter(fmt.Sprintf("failed to parse mention. id=%v, type=%v", to.ID.String(), to.ResourceType), err)
		}

		dMention = append(dMention, *dTo)
	}

	newPostID := uuid.New()
	dPost, err := dfactory.NewPost(newPostID.String(), myMemberID, dMention, int(time.Now().Unix()))
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse post", err)
	}

	if err := co.postService.Create(c, *dPost, topicID, *threadID); err != nil {
		return errors.Wrapf(err, "failed to create post. topic_id=%v", topicID.String())
	}

	newContents := []dmodel.Content{}
	for _, content := range contents {
		newContentID := uuid.New()
		newContent, err := dfactory.NewContent(newContentID.String(), content.Type, content.Bin)
		if err != nil {
			return uerror.NewInvalidParameter(fmt.Sprintf("failed to parse content. type=%v", content.Type), err)
		}

		newContents = append(newContents, *newContent)
	}

	newMention, err := dmodel.NewMention(newPostID.String(), dmodel.ResourcePost.String())
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse mention", err)
	}

	if err := co.contentService.Create(c, newContents, *newMention); err != nil {
		return errors.Wrapf(err, "failed to create content. community_id=%v member_id=%v topic_id=%v", communityID.String(), myMember.ID.String(), topicID.String())
	}

	if err := co.saveIndexAndActivity(c, myMember.ID, newPostID, searchWord, dmodel.ResourcePost, dmodel.OperationCreate); err != nil {
		return err
	}

	return nil
}

func (co *communityUsecase) deleteIndexAndSaveActivity(c context.Context, memberID uuid.UUID, resourceID uuid.UUID, resource dmodel.Resource) error {
	if err := co.deleteIndex(c, resourceID); err != nil {
		return nil
	}

	if err := co.saveMemberActivity(c, memberID, resourceID, resource, dmodel.OperationDelete); err != nil {
		return nil
	}

	return nil
}

func (co *communityUsecase) saveIndexAndActivity(c context.Context, memberID uuid.UUID, resourceID uuid.UUID, name string, resource dmodel.Resource, operation dmodel.Operation) error {
	if name != "" {
		var saveIndex func(c context.Context, resourceID uuid.UUID, name string, resource dmodel.Resource) error
		if operation == dmodel.OperationCreate {
			saveIndex = co.createIndex
		} else {
			saveIndex = co.updateIndex
		}

		if err := saveIndex(c, resourceID, name, resource); err != nil {
			return nil
		}
	}

	if err := co.saveMemberActivity(c, memberID, resourceID, resource, operation); err != nil {
		return nil
	}

	return nil
}

func (co *communityUsecase) createIndex(c context.Context, resourceID uuid.UUID, name string, resource dmodel.Resource) error {
	dIndex, err := dfactory.NewResourceSearchIndex(resourceID.String(), resource.String(), name)
	if err != nil {
		return errors.Wrapf(err, "failed to parse resource search index. id=%v", resourceID.String())
	}

	if err := co.resourceSearchIndexService.Create(c, *dIndex); err != nil {
		return errors.Wrapf(err, "failed to create resource search index. id=%v", resourceID.String())
	}

	return nil
}

func (co *communityUsecase) updateIndex(c context.Context, resourceID uuid.UUID, name string, resource dmodel.Resource) error {
	dIndex, err := dfactory.NewResourceSearchIndex(resourceID.String(), resource.String(), name)
	if err != nil {
		return errors.Wrapf(err, "failed to parse resource search index. id=%v", resourceID.String())
	}

	if err := co.resourceSearchIndexService.Update(c, *dIndex); err != nil {
		return errors.Wrapf(err, "failed to update resource search index. id=%v", resourceID.String())
	}

	return nil
}

func (co *communityUsecase) deleteIndex(c context.Context, resourceID uuid.UUID) error {
	if err := co.resourceSearchIndexService.Delete(c, resourceID); err != nil {
		return errors.Wrapf(err, "failed to delete resource search index. id=%v", resourceID.String())
	}

	return nil
}

func (co *communityUsecase) saveMemberActivity(c context.Context, memberID uuid.UUID, resourceID uuid.UUID, resource dmodel.Resource, operation dmodel.Operation) error {
	dActivity, err := dfactory.NewMemberActivity(time.Now(), memberID.String(), resourceID.String(), resource.String(), operation.String())
	if err != nil {
		return errors.Wrapf(err, "failed to parse member activity. id=%v", memberID.String())
	}

	if err := co.activityService.SaveMemberActivity(c, *dActivity); err != nil {
		return errors.Wrapf(err, "failed to save member activity. id=%v", memberID.String())
	}

	return nil
}

func NewCommunityUsecase(i *do.Injector) (CommunityUsecase, error) {
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
	return &communityUsecase{
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
