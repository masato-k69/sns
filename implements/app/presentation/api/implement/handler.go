package implement

import (
	api "app/gen/api"
	lsession "app/lib/echo/session"
	uerror "app/usecase/error"
	uservice "app/usecase/service"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	activityUsecase uservice.ActivityUsecase
	authUsecase     uservice.AuthUsecase
	userUsecase     uservice.UserUsecase
}

// Auth implements api.ServerInterface.
func (h *Handler) Auth(ctx echo.Context) error {
	url, err := h.authUsecase.GetAuthURL(ctx.Request().Context())

	if h.handle(err); err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, api.AuthResponse{
		Location: *url,
	})
}

// VerifyAuth implements api.ServerInterface.
func (h *Handler) VerifyAuth(ctx echo.Context, params api.VerifyAuthParams) error {
	authenticatedUser, expire, err := h.authUsecase.Verify(ctx.Request().Context(), params.Code)

	if h.handle(err); err != nil {
		return err
	}

	userID, err := h.userUsecase.Save(ctx.Request().Context(), authenticatedUser.Subject, authenticatedUser.Email, authenticatedUser.Issuer, authenticatedUser.Name, authenticatedUser.ImageURL)

	if err != nil {
		return h.handle(err)
	}

	if err := h.activityUsecase.SaveUserLoginActivity(ctx.Request().Context(),
		time.Now(),
		userID.String(),
		ctx.RealIP(),
		strings.Join(ctx.Request().Header["Sec-Ch-Ua-Platform"], ","),
		strings.Join(ctx.Request().Header["Sec-Ch-Ua"], ","),
	); err != nil {
		return h.handle(err)
	}

	if err := lsession.SetLoginSession(ctx, userID.String(), authenticatedUser.Email, authenticatedUser.Name, *expire); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusOK)
}

// DeleteAuth implements api.ServerInterface.
func (h *Handler) DeleteAuth(ctx echo.Context) error {
	if err := lsession.DeleteLoginSession(ctx); err != nil {
		return h.handle(err)
	}

	return ctx.NoContent(http.StatusOK)
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
