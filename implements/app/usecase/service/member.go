package service

import (
	dservice "app/domain/service"
	uerror "app/usecase/error"
	umodel "app/usecase/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type MemberUsecase interface {
	Get(c context.Context, id uuid.UUID) (*umodel.Member, error)
}

type memberUsecase struct {
	memberService dservice.MemberService
	roleService   dservice.RoleService
	userService   dservice.UserService
}

// Get implements MemberUsecase.
func (m *memberUsecase) Get(c context.Context, id uuid.UUID) (*umodel.Member, error) {
	member, err := m.memberService.Get(c, id)
	if err != nil {
		return nil, err
	} else if member == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("member not found. id=%v", id.String()), nil)
	}

	user, err := m.userService.Get(c, member.UserID)
	if err != nil {
		return nil, err
	} else if user == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("user not found. id=%v", member.UserID), nil)
	}

	var role *umodel.Role
	if dRole, err := m.roleService.Get(c, member.RoleID); err != nil {
		return nil, err
	} else if dRole == nil {
		role = nil
	} else {
		role = &umodel.Role{
			ID:     dRole.ID,
			Name:   dRole.Name.String(),
			Action: dRole.Action.Strings(),
		}
	}

	var imageURL *string
	if user.ImageURL != nil {
		v := user.ImageURL.String()
		imageURL = &v
	} else {
		imageURL = nil
	}

	return &umodel.Member{
		ID: member.ID,
		User: umodel.User{
			ID:       user.ID,
			Subject:  user.Subject.String(),
			Email:    user.Email.String(),
			Name:     user.Name.String(),
			ImageUrl: imageURL,
		},
		Role: role,
	}, nil
}

func NewMemberUsecase(i *do.Injector) (MemberUsecase, error) {
	memberService := do.MustInvoke[dservice.MemberService](i)
	roleService := do.MustInvoke[dservice.RoleService](i)
	userService := do.MustInvoke[dservice.UserService](i)

	return &memberUsecase{
		memberService: memberService,
		roleService:   roleService,
		userService:   userService,
	}, nil
}
