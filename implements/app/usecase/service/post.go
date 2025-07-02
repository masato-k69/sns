package service

import (
	dmodel "app/domain/model"
	dservice "app/domain/service"
	uerror "app/usecase/error"
	umodel "app/usecase/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do"
	"github.com/samber/lo"
)

type PostUsecase interface {
	Get(c context.Context, id uuid.UUID) (*umodel.Post, error)
}

type postUsecase struct {
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

// Get implements PostUsecase.
func (t *postUsecase) Get(c context.Context, id uuid.UUID) (*umodel.Post, error) {
	dPost, err := t.postService.Get(c, id)
	if err != nil {
		return nil, err
	} else if dPost == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("post not found. id=%v", id.String()), nil)
	}

	return t.toPost(c, *dPost)
}

func (t *postUsecase) toPost(c context.Context, post dmodel.Post) (*umodel.Post, error) {
	dContents, err := t.contentService.ListByPost(c, post.ID)
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
		dMember, err := t.memberService.Get(c, *post.From)
		if err != nil {
			return nil, err
		} else if dMember == nil {
			return nil, uerror.NewNotFound(fmt.Sprintf("member not found. id=%v", post.From.String()), nil)
		}

		created, err := t.toMember(c, dMember)
		if err != nil {
			return nil, err
		}

		uCreated = created
	}

	likeMention, err := dmodel.NewMention(post.ID.String(), dmodel.ResourcePost.String())
	if err != nil {
		return nil, err
	}

	likes, err := t.activityService.CountMemberLikeActivity(c, *likeMention, true)
	if err != nil {
		return nil, err
	}

	dislikes, err := t.activityService.CountMemberLikeActivity(c, *likeMention, false)
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

func (t *postUsecase) toMember(c context.Context, member *dmodel.Member) (*umodel.Member, error) {
	uUser, err := func(id uuid.UUID) (*umodel.User, error) {
		user, err := t.userService.Get(c, id)
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
		role, err := t.roleService.Get(c, id)
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

func NewPostUsecase(i *do.Injector) (PostUsecase, error) {
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

	return &postUsecase{
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
