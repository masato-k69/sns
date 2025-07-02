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

type RoleUsecase interface {
	ListAction(c context.Context) ([]string, []string, error)
	Get(c context.Context, id uuid.UUID) (*umodel.Role, error)
}

type roleUsecase struct {
	roleService dservice.RoleService
}

// Get implements RoleUsecase.
func (r *roleUsecase) Get(c context.Context, id uuid.UUID) (*umodel.Role, error) {
	role, err := r.roleService.Get(c, id)
	if err != nil {
		return nil, err
	} else if role == nil {
		return nil, uerror.NewNotFound(fmt.Sprintf("role not found. id=%v", id.String()), nil)
	}

	return &umodel.Role{
		ID:     role.ID,
		Name:   role.Name.String(),
		Action: role.Action.Strings(),
	}, nil
}

// ListAction implements RoleUsecase.
func (r *roleUsecase) ListAction(c context.Context) ([]string, []string, error) {
	resources := lo.Map(dmodel.Resources, func(resource dmodel.Resource, _ int) string { return resource.String() })
	operations := lo.Map(dmodel.Operations, func(operation dmodel.Operation, _ int) string { return operation.String() })

	return resources, operations, nil
}

func NewRoleUsecase(i *do.Injector) (RoleUsecase, error) {
	roleService := do.MustInvoke[dservice.RoleService](i)

	return &roleUsecase{
		roleService: roleService,
	}, nil
}
