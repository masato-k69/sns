package v1

import (
	v1 "app/gen/api/v1"
	lsession "app/lib/echo/session"
	llock "app/lib/lock"
	llog "app/lib/log"
	uerror "app/usecase/error"
	umodel "app/usecase/model"
	uservice "app/usecase/service"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/samber/lo"
)

type Handler struct {
	upgrader         websocket.Upgrader
	noteUsecase      uservice.NoteUsecase
	userUsecase      uservice.UserUsecase
	communityUsecase uservice.CommunityUsecase
	roleUsecase      uservice.RoleUsecase
	activityUsecase  uservice.ActivityUsecase
	memberUsecase    uservice.MemberUsecase
	topicUsecase     uservice.TopicUsecase
	postUsecase      uservice.PostUsecase
}

// GetCommunityMember implements v1.ServerInterface.
func (h *Handler) GetCommunityMember(ctx echo.Context, communityId uuid.UUID, memberId uuid.UUID) error {
	member, err := h.memberUsecase.Get(ctx.Request().Context(), memberId)
	if err != nil {
		return h.handle(err)
	}

	var role *v1.Role
	if member.Role != nil {
		actions := []v1.Action{}
		for resource, operation := range member.Role.Action {
			actions = append(actions, v1.Action{
				Resource:   v1.Resource(resource),
				Operations: lo.Map(operation, func(op string, _ int) v1.Operation { return v1.Operation(op) }),
			})
		}
		role = &v1.Role{
			Id:      member.Role.ID,
			Name:    member.Role.Name,
			Actions: actions,
		}
	} else {
		role = nil
	}

	uActivities, err := h.activityUsecase.ListRecentMemberActivity(ctx.Request().Context(), memberId)
	if err != nil {
		return h.handle(err)
	}

	pActivities, err := h.listActivity(ctx, uActivities)
	if err != nil {
		return h.handle(err)
	}

	return ctx.JSON(http.StatusOK, v1.GetCommunityMemberResponse{
		Member: v1.Member{
			Id:   member.ID,
			Role: role,
			User: v1.User{
				Id:    member.User.ID,
				Name:  member.User.Name,
				Image: member.User.ImageUrl,
			},
		},
		RecentActivities: pActivities,
	})
}

// ListUserActivity implements v1.ServerInterface.
func (h *Handler) ListUserActivity(ctx echo.Context, userId uuid.UUID, params v1.ListUserActivityParams) error {
	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	uActivities, err := h.activityUsecase.ListUsersMemberActivity(ctx.Request().Context(), loggedInUser.ID, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pActivities, err := h.listActivity(ctx, uActivities)
	if err != nil {
		return h.handle(err)
	}

	return ctx.JSON(http.StatusOK, v1.ListUserActivityResponse{
		Activities: pActivities,
	})
}

// ListUserLoginActivity implements v1.ServerInterface.
func (h *Handler) ListUserLoginActivity(ctx echo.Context, params v1.ListUserLoginActivityParams) error {
	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	uActivities, err := h.activityUsecase.ListUserLoginActivity(ctx.Request().Context(), loggedInUser.ID, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pActivities := []v1.Activity{}
	for _, uActivity := range uActivities {
		when := int(uActivity.At.Unix())

		did := v1.Experience_Resource{}
		if err := did.FromLogin(v1.Login{
			IpAddress:       uActivity.IPAddress,
			OperationSystem: uActivity.OperationSystem,
			UserAgent:       uActivity.UserAgent,
		}); err != nil {
			return err
		}

		pActivities = append(pActivities, v1.Activity{
			When: &when,
			Did: v1.Experience{
				Resource:  did,
				Operation: v1.OperationCreate,
			},
		})
	}

	return ctx.JSON(http.StatusOK, v1.ListUserLoginActivityResponse{
		Activities: pActivities,
	})
}

// ListPostLike implements v1.ServerInterface.
func (h *Handler) ListPostLike(ctx echo.Context, communityId uuid.UUID, topicId uuid.UUID, threadId uuid.UUID, postId uuid.UUID, params v1.ListPostLikeParams) error {
	likes, err := h.communityUsecase.ListPostLike(ctx.Request().Context(), communityId, topicId, threadId, postId, params.Like, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pLikes := []v1.Like{}
	for _, like := range likes {
		var comment *v1.ShortMessage
		if like.Comment != nil {
			comment = like.Comment
		} else {
			comment = nil
		}

		var pMember *v1.Member
		if like.By != nil {
			pMember = h.buildMember(*like.By)
		}

		pLikes = append(pLikes, v1.Like{
			Like:    like.Like,
			Comment: comment,
			By:      pMember,
		})
	}

	return ctx.JSON(http.StatusOK, v1.ListPostLikeResponse{
		Likes: pLikes,
	})
}

// LikePost implements v1.ServerInterface.
func (h *Handler) LikePost(ctx echo.Context, communityId uuid.UUID, topicId uuid.UUID, threadId uuid.UUID, postId uuid.UUID) error {
	var body v1.LikeRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	if err := h.communityUsecase.LikePost(ctx.Request().Context(), communityId, topicId, threadId, postId, loggedInUser.ID, body.Like, body.Comment); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// CreateCommunityPost implements v1.ServerInterface.
func (h *Handler) CreateCommunityPost(ctx echo.Context, communityId uuid.UUID, topicId uuid.UUID, threadId uuid.UUID) error {
	var body v1.CreatePostRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	contents := []umodel.Content{}
	for _, content := range body.Contents {
		_, bin, err := ToMetaAndBin(content)
		if err != nil {
			return err
		}

		contents = append(contents, umodel.Content{
			Type: string(content.Type),
			Bin:  bin,
		})
	}

	mentions, err := ListMention(body.Contents)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	uMention := lo.Map(mentions, func(mention v1.Mention, _ int) umodel.Mention {
		return umodel.Mention{ID: mention.Id, ResourceType: string(mention.Resource)}
	})

	searchWord, err := ToText(body.Contents)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	if err := h.communityUsecase.Reply(ctx.Request().Context(), communityId, loggedInUser.ID, topicId, threadId, contents, uMention, *searchWord); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

// CreateCommunityThread implements v1.ServerInterface.
func (h *Handler) CreateCommunityThread(ctx echo.Context, communityId uuid.UUID, topicId uuid.UUID) error {
	var body v1.CreateThreadRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	contents := []umodel.Content{}
	for _, content := range body.Contents {
		_, bin, err := ToMetaAndBin(content)
		if err != nil {
			return err
		}

		contents = append(contents, umodel.Content{
			Type: string(content.Type),
			Bin:  bin,
		})
	}

	mentions, err := ListMention(body.Contents)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	uMention := lo.Map(mentions, func(mention v1.Mention, _ int) umodel.Mention {
		return umodel.Mention{ID: mention.Id, ResourceType: string(mention.Resource)}
	})

	searchWord, err := ToText(body.Contents)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	if err := h.communityUsecase.Post(ctx.Request().Context(), communityId, loggedInUser.ID, topicId, contents, uMention, *searchWord); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

// CreateCommunityTopic implements v1.ServerInterface.
func (h *Handler) CreateCommunityTopic(ctx echo.Context, communityId uuid.UUID) error {
	var body v1.CreateTopicRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	contents := []umodel.Content{}
	for _, content := range body.Contents {
		_, bin, err := ToMetaAndBin(content)
		if err != nil {
			return err
		}

		contents = append(contents, umodel.Content{
			Type: string(content.Type),
			Bin:  bin,
		})
	}

	topicID, err := h.communityUsecase.CreateTopic(ctx.Request().Context(), communityId, loggedInUser.ID, body.Name, contents)
	if err != nil {
		return h.handle(err)
	}

	return ctx.JSON(http.StatusCreated, v1.CreateTopicResponse{
		Id: *topicID,
	})
}

// ListCommunityPost implements v1.ServerInterface.
func (h *Handler) ListCommunityPost(ctx echo.Context, communityId uuid.UUID, topicId uuid.UUID, threadId uuid.UUID, params v1.ListCommunityPostParams) error {
	posts, err := h.communityUsecase.ListPost(ctx.Request().Context(), communityId, topicId, threadId, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pPosts := []v1.Post{}
	for _, post := range posts {
		pContents := []v1.Content{}
		for _, content := range post.Contents {
			pContent, err := NewContent(content.Type, content.Bin)
			if err != nil {
				return err
			}

			pContents = append(pContents, *pContent)
		}

		var pMember *v1.Member
		if post.Created != nil {
			pMember = h.buildMember(*post.Created)
		}

		pPosts = append(pPosts, v1.Post{
			Id:       post.ID,
			At:       post.At,
			Contents: pContents,
			Created:  pMember,
			Reaction: v1.Reaction{
				Likes:    post.Reaction.Likes,
				Dislikes: post.Reaction.Dislikes,
			},
		})
	}

	return ctx.JSON(http.StatusOK, v1.ListPostResponse{
		Posts: pPosts,
	})
}

// ListCommunityThread implements v1.ServerInterface.
func (h *Handler) ListCommunityThread(ctx echo.Context, communityId uuid.UUID, topicId uuid.UUID, params v1.ListCommunityThreadParams) error {
	threads, err := h.communityUsecase.ListThread(ctx.Request().Context(), communityId, topicId, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pThreads := []v1.Thread{}
	for _, thread := range threads {
		firstPost, exists := lo.First(thread.Posts)
		if !exists {
			return echo.NewHTTPError(http.StatusInternalServerError, fmt.Errorf("post does not exists. thread_id=%v", thread.ID.String()))
		}

		pContents := []v1.Content{}
		for _, content := range firstPost.Contents {
			pContent, err := NewContent(content.Type, content.Bin)
			if err != nil {
				return err
			}

			pContents = append(pContents, *pContent)
		}

		var pMember *v1.Member
		if firstPost.Created != nil {
			pMember = h.buildMember(*firstPost.Created)
		}

		pThreads = append(pThreads, v1.Thread{
			Id: thread.ID,
			FirstPost: v1.Post{
				Id:       firstPost.ID,
				At:       firstPost.At,
				Contents: pContents,
				Created:  pMember,
				Reaction: v1.Reaction{
					Likes:    firstPost.Reaction.Likes,
					Dislikes: firstPost.Reaction.Dislikes,
				},
			},
			Reply: len(thread.Posts) > 1,
		})
	}

	return ctx.JSON(http.StatusOK, v1.ListThreadResponse{
		Threads: pThreads,
	})
}

// ListCommunityTopic implements v1.ServerInterface.
func (h *Handler) ListCommunityTopic(ctx echo.Context, communityId uuid.UUID, params v1.ListCommunityTopicParams) error {
	topics, err := h.communityUsecase.ListTopic(ctx.Request().Context(), communityId, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pTopics := []v1.Topic{}
	for _, topic := range topics {
		pContents := []v1.Content{}
		for _, content := range topic.Contents {
			pContent, err := NewContent(content.Type, content.Bin)
			if err != nil {
				return err
			}

			pContents = append(pContents, *pContent)
		}

		var pMember *v1.Member
		if topic.Created != nil {
			pMember = h.buildMember(*topic.Created)
		}

		var pLastPost *v1.Post
		if topic.LastPost != nil {
			pContents := []v1.Content{}
			for _, content := range topic.LastPost.Contents {
				pContent, err := NewContent(content.Type, content.Bin)
				if err != nil {
					return err
				}

				pContents = append(pContents, *pContent)
			}

			var pMember *v1.Member
			if topic.Created != nil {
				pMember = h.buildMember(*topic.LastPost.Created)
			}

			pLastPost = &v1.Post{
				Id:       topic.LastPost.ID,
				At:       topic.LastPost.At,
				Contents: pContents,
				Created:  pMember,
				Reaction: v1.Reaction{
					Likes:    topic.LastPost.Reaction.Likes,
					Dislikes: topic.LastPost.Reaction.Dislikes,
				},
			}
		}

		pTopics = append(pTopics, v1.Topic{
			Id:       topic.ID,
			Name:     topic.Name,
			Contents: pContents,
			Created:  pMember,
			LastPost: pLastPost,
		})
	}

	return ctx.JSON(http.StatusOK, v1.ListTopicResponse{
		Topics: pTopics,
	})
}

// ListCommunityMember implements v1.ServerInterface.
func (h *Handler) ListCommunityMember(ctx echo.Context, communityId uuid.UUID, params v1.ListCommunityMemberParams) error {
	members, err := h.communityUsecase.ListMember(ctx.Request().Context(), communityId, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pMembers := []v1.Member{}
	for _, member := range members {
		var role *v1.Role
		if member.Role != nil {
			actions := []v1.Action{}
			for resource, operation := range member.Role.Action {
				actions = append(actions, v1.Action{
					Resource:   v1.Resource(resource),
					Operations: lo.Map(operation, func(op string, _ int) v1.Operation { return v1.Operation(op) }),
				})
			}
			role = &v1.Role{
				Id:      member.Role.ID,
				Name:    member.Role.Name,
				Actions: actions,
			}
		} else {
			role = nil
		}

		pMembers = append(pMembers, v1.Member{
			Id:   member.ID,
			Role: role,
			User: v1.User{
				Id:    member.User.ID,
				Name:  member.User.Name,
				Image: member.User.ImageUrl,
			},
		})
	}

	return ctx.JSON(http.StatusOK, &v1.ListCommunityMemberResponse{
		Members: pMembers,
	})
}

// DeleteCommunityInvite implements v1.ServerInterface.
func (h *Handler) DeleteCommunityInvite(ctx echo.Context, communityId uuid.UUID, inviteId uuid.UUID) error {
	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	if err := h.communityUsecase.DeleteInvite(ctx.Request().Context(), communityId, loggedInUser.ID, inviteId); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// InviteCommunityRole implements v1.ServerInterface.
func (h *Handler) InviteCommunityRole(ctx echo.Context, communityId uuid.UUID, roleId uuid.UUID) error {
	var body v1.InviteCommunityRoleRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	userIDs := lo.Map(body.Mention, func(mention v1.Mention, _ int) uuid.UUID { return mention.Id })
	if err := h.communityUsecase.Invite(ctx.Request().Context(), communityId, loggedInUser.ID, roleId, userIDs, &body.Message); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

// ListCommunityInvite implements v1.ServerInterface.
func (h *Handler) ListCommunityInvite(ctx echo.Context, communityId uuid.UUID, params v1.ListCommunityInviteParams) error {
	invites, err := h.communityUsecase.ListInvite(ctx.Request().Context(), communityId, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pInvites := []v1.CommunityInvite{}
	for _, invite := range invites {
		actions := []v1.Action{}
		for resource, operation := range invite.Role.Action {
			actions = append(actions, v1.Action{
				Resource:   v1.Resource(resource),
				Operations: lo.Map(operation, func(op string, _ int) v1.Operation { return v1.Operation(op) }),
			})
		}

		var message *v1.ShortMessage
		if invite.Message != nil {
			message = invite.Message
		} else {
			message = nil
		}

		pInvites = append(pInvites, v1.CommunityInvite{
			Id: invite.ID,
			Role: v1.Role{
				Id:      invite.Role.ID,
				Name:    invite.Role.Name,
				Actions: actions,
			},
			Users: lo.Map(invite.Users, func(user umodel.User, _ int) v1.User {
				return v1.User{
					Id:    user.ID,
					Name:  user.Name,
					Image: user.ImageUrl,
				}
			}),
			At:      int(invite.At.Unix()),
			Message: message,
		})
	}

	return ctx.JSON(http.StatusOK, &v1.ListCommunityInviteResponse{
		Invites: pInvites,
	})
}

// ListUserInvite implements v1.ServerInterface.
func (h *Handler) ListUserInvite(ctx echo.Context, params v1.ListUserInviteParams) error {
	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	invites, err := h.userUsecase.ListInvite(ctx.Request().Context(), loggedInUser.ID, params.Limit, params.Offset)
	if err != nil {
		return h.handle(err)
	}

	pInvites := []v1.UserInvite{}
	for _, invite := range invites {
		actions := []v1.Action{}
		for resource, operation := range invite.Role.Action {
			actions = append(actions, v1.Action{
				Resource:   v1.Resource(resource),
				Operations: lo.Map(operation, func(op string, _ int) v1.Operation { return v1.Operation(op) }),
			})
		}

		var message *v1.ShortMessage
		if invite.Message != nil {
			message = invite.Message
		} else {
			message = nil
		}

		pInvites = append(pInvites, v1.UserInvite{
			Id: invite.ID,
			Community: v1.Community{
				Id:   invite.Community.ID,
				Name: invite.Community.Name,
			},
			Role: v1.Role{
				Id:      invite.Community.ID,
				Name:    invite.Community.Name,
				Actions: actions,
			},
			At:      int(invite.At.Unix()),
			Message: message,
		})
	}

	return ctx.JSON(http.StatusOK, &v1.ListUserInviteResponse{
		Invites: pInvites,
	})
}

// ReplyInvite implements v1.ServerInterface.
func (h *Handler) ReplyInvite(ctx echo.Context, inviteId uuid.UUID) error {
	var body v1.ReplyUserInviteRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	if err := h.userUsecase.ReplyInvite(ctx.Request().Context(), loggedInUser.ID, inviteId, body.Agree); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// CreateCommunityRole implements v1.ServerInterface.
func (h *Handler) CreateCommunityRole(ctx echo.Context, communityId uuid.UUID) error {
	var body v1.CreateCommunityRoleRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	actions := map[string][]string{}
	for _, action := range body.Actions {
		if _, ok := actions[string(action.Resource)]; ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("duplicate resource. v=%v", action.Resource))
		}

		operations := lo.Map(action.Operations, func(operation v1.Operation, _ int) string { return string(operation) })
		actions[string(action.Resource)] = operations
	}

	if err := h.communityUsecase.CreateRole(ctx.Request().Context(), communityId, loggedInUser.ID, body.Name, actions); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusCreated)
}

// DeleteCommunityRole implements v1.ServerInterface.
func (h *Handler) DeleteCommunityRole(ctx echo.Context, communityId uuid.UUID, roleId uuid.UUID) error {
	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	if err := h.communityUsecase.DeleteRole(ctx.Request().Context(), communityId, loggedInUser.ID, roleId); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// ListCommunityRole implements v1.ServerInterface.
func (h *Handler) ListCommunityRole(ctx echo.Context, communityId uuid.UUID) error {
	roles, err := h.communityUsecase.ListRole(ctx.Request().Context(), communityId)
	if err != nil {
		return h.handle(err)
	}

	pRoles := lo.Map(roles, func(role umodel.Role, _ int) v1.Role {
		action := []v1.Action{}
		for resource, operations := range role.Action {
			action = append(action, v1.Action{
				Resource:   v1.Resource(resource),
				Operations: lo.Map(operations, func(operaiton string, _ int) v1.Operation { return v1.Operation(operaiton) }),
			})
		}

		return v1.Role{
			Id:      role.ID,
			Name:    role.Name,
			Actions: action,
		}
	})

	return ctx.JSON(http.StatusOK, v1.ListCommunityRoleResponse{
		Roles: pRoles,
	})
}

// UpdateCommunityRole implements v1.ServerInterface.
func (h *Handler) UpdateCommunityRole(ctx echo.Context, communityId uuid.UUID, roleId uuid.UUID) error {
	var body v1.UpdateCommunityRoleRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	actions := map[string][]string{}
	for _, action := range body.Actions {
		if _, ok := actions[string(action.Resource)]; ok {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Errorf("duplicate resource. v=%v", action.Resource))
		}

		operations := lo.Map(action.Operations, func(operation v1.Operation, _ int) string { return string(operation) })
		actions[string(action.Resource)] = operations
	}

	if err := h.communityUsecase.UpdateRole(ctx.Request().Context(), communityId, loggedInUser.ID, roleId, body.Name, actions); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// ListAction implements v1.ServerInterface.
func (h *Handler) ListAction(ctx echo.Context) error {
	resources, operations, err := h.roleUsecase.ListAction(ctx.Request().Context())
	if err != nil {
		return h.handle(err)
	}

	return ctx.JSON(http.StatusOK, &v1.ListActionResponse{
		Resources:  lo.Map(resources, func(resource string, _ int) v1.Resource { return v1.Resource(resource) }),
		Operations: lo.Map(operations, func(operation string, _ int) v1.Operation { return v1.Operation(operation) }),
	})
}

// EditCommunityDescription implements v1.ServerInterface.
func (h *Handler) EditCommunityDescription(ctx echo.Context, id uuid.UUID, params v1.EditCommunityDescriptionParams) error {
	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	community, err := h.communityUsecase.CanUpdate(ctx.Request().Context(), id, loggedInUser.ID)
	if err != nil {
		return h.handle(err)
	}

	description, err := h.noteUsecase.GetCommunityDescription(ctx.Request().Context(), community.ID)
	if err != nil {
		return h.handle(err)
	}

	return h.editNote(ctx, description.ID)
}

// UpdateCommunity implements v1.ServerInterface.
func (h *Handler) UpdateCommunity(ctx echo.Context, id uuid.UUID) error {
	var body v1.UpdateCommunityRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	if err := h.communityUsecase.Update(ctx.Request().Context(), id, loggedInUser.ID, body.Name); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// CreateCommunity implements v1.ServerInterface.
func (h *Handler) CreateCommunity(ctx echo.Context) error {
	var body v1.CreateCommunityRequest
	if err := ctx.Bind(&body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	communityID, err := h.communityUsecase.Create(ctx.Request().Context(), loggedInUser.ID, body.Name, body.Invitation)
	if err != nil {
		return h.handle(err)
	}

	return ctx.JSON(http.StatusCreated, &v1.CreateCommunityResponse{
		Id: *communityID,
	})
}

// EditUserProfile implements v1.ServerInterface.
func (h *Handler) EditUserProfile(ctx echo.Context, _ v1.EditUserProfileParams) (err error) {
	loggedInUser, err := lsession.GetLoginSession(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error()).SetInternal(err)
	}

	user, err := h.userUsecase.Get(ctx.Request().Context(), loggedInUser.ID)
	if err != nil {
		return h.handle(err)
	}

	profile, err := h.noteUsecase.GetUserProfile(ctx.Request().Context(), user.ID)
	if err != nil {
		return h.handle(err)
	}

	return h.editNote(ctx, profile.ID)
}

func (h *Handler) editNote(ctx echo.Context, id uuid.UUID) error {
	if _, err := h.noteUsecase.Get(ctx.Request().Context(), id); err != nil {
		return h.handle(err)
	}

	ws, err := h.upgrader.Upgrade(ctx.Response(), ctx.Request(), nil)
	if err != nil {
		return err
	} else if ws == nil {
		return nil
	}

	defer ws.Close()

	lines, err := h.noteUsecase.ListLines(ctx.Request().Context(), id)
	if err != nil {
		return h.handle(err)
	}

	currentLinesMessage := v1.CurrentLinesMessage{}
	for _, line := range lines {
		var property *v1.LineProperty
		if line.Property == nil {
			property = nil
		} else {
			property = &v1.LineProperty{
				Type: v1.LinePropertyType(line.Property.Type),
			}
		}

		contents := []v1.Content{}

		for _, content := range line.Contents {
			parsedContent, err := NewContent(content.Type, content.Bin)

			if err != nil {
				return err
			}

			contents = append(contents, *parsedContent)
		}

		currentLinesMessage = append(currentLinesMessage, v1.Line{
			Order:    line.Order,
			Property: property,
			Contents: contents,
		})
	}

	messagesToSend, err := NewMessagesToSend(currentLinesMessage)
	if err != nil {
		return err
	}

	messagesToSendBin, err := json.Marshal(&messagesToSend)
	if err != nil {
		return err
	}

	if err := ws.WriteMessage(websocket.TextMessage, messagesToSendBin); err != nil {
		return err
	}

	if exist, err := llock.Exist(ctx.Request().Context(), id); err != nil {
		return err
	} else if exist {
		err = fmt.Errorf("locked")
		return echo.NewHTTPError(http.StatusLocked, err.Error()).SetInternal(err)
	}

	defer func() {
		err = llock.Delete(ctx.Request().Context(), id)
	}()

	deadline := func() time.Duration {
		timeoutSeconds, err := strconv.Atoi(os.Getenv("TIMEOUT_SECONDS_WEBSOCKET"))

		if err != nil {
			return time.Duration(timeoutSeconds) * time.Second
		}

		return 1800 * time.Second
	}()

	for {
		ws.SetReadDeadline(time.Now().Add(deadline))

		if err := llock.Set(ctx.Request().Context(), id, deadline); err != nil {
			return err
		}

		_, recievedMessage, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				llog.Info(ctx.Request().Context(), "recieve close messege. err=%v", err)
				break
			}

			llog.Error(ctx.Request().Context(), "recieve error. err=%v:%v", reflect.TypeOf(err), err)
			break
		}

		if err := HandleRecievedMessage(recievedMessage, map[v1.RecieveType]MessageReciever{
			v1.RecieveTypeInsert: func(v v1.MessagesToRecieve_Entity) error {
				message, err := v.AsInsertedLineMessage()

				if err != nil {
					return err
				}

				if err := h.noteUsecase.InsertLine(ctx.Request().Context(), id, message.To); err != nil {
					return h.handle(err)
				}

				return nil
			},
			v1.RecieveTypeMove: func(v v1.MessagesToRecieve_Entity) error {
				message, err := v.AsMovedLineMessage()

				if err != nil {
					return err
				}

				if err := h.noteUsecase.MoveLine(ctx.Request().Context(), id, message.From, message.To); err != nil {
					return h.handle(err)
				}

				return nil
			},
			v1.RecieveTypeUpdate: func(v v1.MessagesToRecieve_Entity) error {
				message, err := v.AsEditedLineMessage()

				if err != nil {
					return err
				}

				contents := []umodel.Content{}
				for _, content := range message.Contents {
					_, bin, err := ToMetaAndBin(content)
					if err != nil {
						return err
					}

					contents = append(contents, umodel.Content{
						Type: string(content.Type),
						Bin:  bin,
					})
				}

				var property *umodel.LineProperty
				if message.Property == nil {
					property = nil
				} else {
					property = &umodel.LineProperty{
						Type: string(message.Property.Type),
					}
				}

				if err := h.noteUsecase.UpdateLine(ctx.Request().Context(), umodel.Line{
					NoteID:   id,
					Order:    message.Order,
					Property: property,
					Contents: contents,
				}); err != nil {
					return h.handle(err)
				}

				return nil
			},
			v1.RecieveTypeDelete: func(v v1.MessagesToRecieve_Entity) error {
				message, err := v.AsDeletedLineMessage()

				if err != nil {
					return err
				}

				if err := h.noteUsecase.DeleteLine(ctx.Request().Context(), id, message.To); err != nil {
					return h.handle(err)
				}

				return nil
			},
		}); err != nil {
			llog.Error(ctx.Request().Context(), "recieve error. err=%v:%v", reflect.TypeOf(err), err)
			break
		}
	}

	return nil
}

func (h *Handler) listActivity(ctx echo.Context, uActivities []umodel.Activity) ([]v1.Activity, error) {
	pActivities := []v1.Activity{}
	for _, uActivity := range uActivities {
		when := int(uActivity.At.Unix())

		resourceType, exists := lo.Find(supportedResources, func(r v1.Resource) bool { return r == v1.Resource(uActivity.Resource) })
		if !exists {
			llog.Warn(ctx.Request().Context(), "unsupported resource. value=%v", uActivity.Resource)
			continue
		}

		experienceOperation, exists := lo.Find(supportedOperations, func(r v1.Operation) bool { return r == v1.Operation(uActivity.Operation) })
		if !exists {
			llog.Warn(ctx.Request().Context(), "unsupported operation. value=%v", uActivity.Operation)
			continue
		}

		experienceResource := v1.Experience_Resource{}
		var where *v1.Activity_Where
		switch resourceType {
		case v1.ResourceCommunity:
			resource, err := h.communityUsecase.Get(ctx.Request().Context(), uActivity.Target)
			if err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "resource not found. id=%v", uActivity.Target)
					continue
				}

				return nil, err
			}

			if err := experienceResource.FromCommunity(v1.Community{
				Id:   resource.ID,
				Name: resource.Name,
			}); err != nil {
				return nil, err
			}

			where = &v1.Activity_Where{}
			if err := where.FromCommunity(v1.Community{
				Id:   resource.ID,
				Name: resource.Name,
			}); err != nil {
				return nil, err
			}
		case v1.ResourceMember:
			resource, err := h.memberUsecase.Get(ctx.Request().Context(), uActivity.Target)
			if err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "resource not found. id=%v", uActivity.Target)
					continue
				}

				return nil, err
			}

			var role *v1.Role
			if resource.Role != nil {
				actions := []v1.Action{}
				for resource, operation := range resource.Role.Action {
					actions = append(actions, v1.Action{
						Resource:   v1.Resource(resource),
						Operations: lo.Map(operation, func(op string, _ int) v1.Operation { return v1.Operation(op) }),
					})
				}
				role = &v1.Role{
					Id:      resource.Role.ID,
					Name:    resource.Role.Name,
					Actions: actions,
				}
			} else {
				role = nil
			}

			if err := experienceResource.FromMember(v1.Member{
				Id:   resource.ID,
				Role: role,
				User: v1.User{
					Id:    resource.User.ID,
					Name:  resource.User.Name,
					Image: resource.User.ImageUrl,
				},
			}); err != nil {
				return nil, err
			}

			if community, err := h.communityUsecase.GetByMember(ctx.Request().Context(), uActivity.Me); err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "community not found. id=%v", uActivity.Target)
					break
				}

				return nil, err
			} else {
				where = &v1.Activity_Where{}
				if err := where.FromCommunity(v1.Community{
					Id:   community.ID,
					Name: community.Name,
				}); err != nil {
					return nil, err
				}
			}
		case v1.ResourceRole:
			resource, err := h.roleUsecase.Get(ctx.Request().Context(), uActivity.Target)
			if err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "resource not found. id=%v", uActivity.Target)
					continue
				}

				return nil, err
			}

			actions := []v1.Action{}
			for resource, operation := range resource.Action {
				actions = append(actions, v1.Action{
					Resource:   v1.Resource(resource),
					Operations: lo.Map(operation, func(op string, _ int) v1.Operation { return v1.Operation(op) }),
				})
			}

			if err := experienceResource.FromRole(v1.Role{
				Id:      resource.ID,
				Name:    resource.Name,
				Actions: actions,
			}); err != nil {
				return nil, err
			}

			if community, err := h.communityUsecase.GetByMember(ctx.Request().Context(), uActivity.Me); err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "community not found. id=%v", uActivity.Target)
					break
				}

				return nil, err
			} else {
				where = &v1.Activity_Where{}
				if err := where.FromCommunity(v1.Community{
					Id:   community.ID,
					Name: community.Name,
				}); err != nil {
					return nil, err
				}
			}
		case v1.ResourceTopic:
			resource, err := h.topicUsecase.Get(ctx.Request().Context(), uActivity.Target)
			if err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "resource not found. id=%v", uActivity.Target)
					continue
				}

				return nil, err
			}

			pContents := []v1.Content{}
			for _, content := range resource.Contents {
				pContent, err := NewContent(content.Type, content.Bin)
				if err != nil {
					return nil, err
				}

				pContents = append(pContents, *pContent)
			}

			var pMember *v1.Member
			if resource.Created != nil {
				pMember = h.buildMember(*resource.Created)
			}

			var pLastPost *v1.Post
			if resource.LastPost != nil {
				pContents := []v1.Content{}
				for _, content := range resource.LastPost.Contents {
					pContent, err := NewContent(content.Type, content.Bin)
					if err != nil {
						return nil, err
					}

					pContents = append(pContents, *pContent)
				}

				var pMember *v1.Member
				if resource.Created != nil {
					pMember = h.buildMember(*resource.Created)
				}

				pLastPost = &v1.Post{
					Id:       resource.LastPost.ID,
					At:       resource.LastPost.At,
					Contents: pContents,
					Created:  pMember,
					Reaction: v1.Reaction{
						Likes:    resource.LastPost.Reaction.Likes,
						Dislikes: resource.LastPost.Reaction.Dislikes,
					},
				}
			}

			if err := experienceResource.FromTopic(v1.Topic{
				Id:       resource.ID,
				Name:     resource.Name,
				Contents: pContents,
				Created:  pMember,
				LastPost: pLastPost,
			}); err != nil {
				return nil, err
			}

			if community, err := h.communityUsecase.GetByMember(ctx.Request().Context(), uActivity.Me); err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "community not found. id=%v", uActivity.Target)
					break
				}

				return nil, err
			} else {
				where = &v1.Activity_Where{}
				if err := where.FromCommunity(v1.Community{
					Id:   community.ID,
					Name: community.Name,
				}); err != nil {
					return nil, err
				}
			}
		case v1.ResourcePost:
			resource, err := h.postUsecase.Get(ctx.Request().Context(), uActivity.Target)
			if err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "resource not found. id=%v", uActivity.Target)
					continue
				}

				return nil, err
			}
			pContents := []v1.Content{}
			for _, content := range resource.Contents {
				pContent, err := NewContent(content.Type, content.Bin)
				if err != nil {
					return nil, err
				}

				pContents = append(pContents, *pContent)
			}

			var pMember *v1.Member
			if resource.Created != nil {
				pMember = h.buildMember(*resource.Created)
			}

			if err := experienceResource.FromPost(v1.Post{
				Id:       resource.ID,
				At:       resource.At,
				Contents: pContents,
				Created:  pMember,
				Reaction: v1.Reaction{
					Likes:    resource.Reaction.Likes,
					Dislikes: resource.Reaction.Dislikes,
				},
			}); err != nil {
				return nil, err
			}

			if community, err := h.communityUsecase.GetByMember(ctx.Request().Context(), uActivity.Me); err != nil {
				if _, ok := err.(uerror.NotFound); ok {
					llog.Debug(ctx.Request().Context(), "community not found. id=%v", uActivity.Target)
					break
				}

				return nil, err
			} else {
				where = &v1.Activity_Where{}
				if err := where.FromCommunity(v1.Community{
					Id:   community.ID,
					Name: community.Name,
				}); err != nil {
					return nil, err
				}
			}
		}

		pActivities = append(pActivities, v1.Activity{
			When:  &when,
			Where: where,
			Did: v1.Experience{
				Resource:  experienceResource,
				Operation: experienceOperation,
			},
		})
	}

	return pActivities, nil
}

func (h *Handler) buildMember(member umodel.Member) *v1.Member {
	var role *v1.Role
	if member.Role != nil {
		actions := []v1.Action{}
		for resource, operation := range member.Role.Action {
			actions = append(actions, v1.Action{
				Resource:   v1.Resource(resource),
				Operations: lo.Map(operation, func(op string, _ int) v1.Operation { return v1.Operation(op) }),
			})
		}
		role = &v1.Role{
			Id:      member.Role.ID,
			Name:    member.Role.Name,
			Actions: actions,
		}
	} else {
		role = nil
	}

	return &v1.Member{
		Id:   member.ID,
		Role: role,
		User: v1.User{
			Id:    member.User.ID,
			Name:  member.User.Name,
			Image: member.User.ImageUrl,
		},
	}

}

func (h *Handler) handle(err error) error {
	if err == nil {
		return nil
	}

	switch err.(type) {
	case uerror.InvalidParameter:
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	case uerror.PermissionDenied:
		return echo.NewHTTPError(http.StatusForbidden, err.Error()).SetInternal(err)
	case uerror.NotFound:
		return echo.NewHTTPError(http.StatusNotFound, err.Error()).SetInternal(err)
	case uerror.AlreadyExists:
		return echo.NewHTTPError(http.StatusConflict, err.Error()).SetInternal(err)
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error()).SetInternal(err)
	}
}
