package service

import (
	lgoogle "app/lib/auth/google"
	uerror "app/usecase/error"
	umodel "app/usecase/model"
	"context"
	"time"

	"github.com/samber/do"
)

type AuthUsecase interface {
	GetAuthURL(c context.Context) (*string, error)
	Verify(c context.Context, authCode string) (*umodel.AuthenticatedUser, *time.Time, error)
}

type authUsecase struct {
}

// GetAuthURL implements AuthUsecase.
func (a *authUsecase) GetAuthURL(c context.Context) (*string, error) {
	url := lgoogle.GetAuthURL()
	return &url, nil
}

// Verify implements AuthUsecase.
func (a *authUsecase) Verify(c context.Context, authCode string) (*umodel.AuthenticatedUser, *time.Time, error) {
	authenticateUser, expire, err := lgoogle.GetAuthenticatedUser(c, authCode)

	if err != nil {
		return nil, nil, uerror.NewInvalidParameter("failed to get authenticated user", err)
	}

	return &umodel.AuthenticatedUser{
		Subject:  authenticateUser.Subject,
		Email:    authenticateUser.Email,
		Issuer:   authenticateUser.Issuer,
		Name:     authenticateUser.Name,
		ImageURL: authenticateUser.ImageURL,
	}, expire, nil
}

func NewAuthUsecase(i *do.Injector) (AuthUsecase, error) {
	return &authUsecase{}, nil
}
