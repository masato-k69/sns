package presentation

import (
	api "app/gen/api"
	apiv1 "app/gen/api/v1"
	lgoogle "app/lib/auth/google"
	lcontext "app/lib/context"
	lsession "app/lib/echo/session"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echomiddleware "github.com/oapi-codegen/echo-middleware"
)

func NewMiddlewaresCommon() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		secure(),
		recover(),
		cors(),
		lsession.SessionStore(),
		requestID(),
		requestLog(),
		timeout(),
	}
}

func NewMiddlewares() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		validator(),
	}
}

func NewMiddlewaresForV1() []echo.MiddlewareFunc {
	return []echo.MiddlewareFunc{
		validatorForV1(),
		auth(),
	}
}

func secure() echo.MiddlewareFunc {
	return middleware.Secure()
}

func recover() echo.MiddlewareFunc {
	return middleware.RecoverWithConfig(middleware.RecoverConfig{
		LogErrorFunc: func(c echo.Context, err error, stack []byte) error {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

			req := c.Request()
			res := c.Response()

			userID := func() string {
				session, err := lsession.GetLoginSession(c)

				if err == nil {
					return session.ID.String()
				}

				return "-"
			}

			logger.LogAttrs(c.Request().Context(), slog.LevelError, "RECOVERY",
				slog.String(lcontext.ContextKeyRequestID.String(), res.Header().Get(echo.HeaderXRequestID)),
				slog.String("user_id", userID()),
				slog.Int("status", http.StatusInternalServerError),
				slog.String("method", req.Method),
				slog.String("uri", req.RequestURI),
				slog.String("remote_ip", c.RealIP()),
				slog.String("referer", req.Referer()),
				slog.String("request_header", fmt.Sprintf("%+v", req.Header)),
				slog.String("err", fmt.Sprintf("%+v\n%+v", err, string(stack))),
			)

			return c.NoContent(http.StatusInternalServerError)
		},
	})
}

func cors() echo.MiddlewareFunc {
	return middleware.CORSWithConfig(middleware.CORSConfig{
		// Skipper: func(c echo.Context) bool { return environment.IsDebug() },
		AllowOrigins: func() []string {
			var cors []string
			json.Unmarshal([]byte(os.Getenv("ALLOW_ORIGINS")), &cors)

			fmt.Println(cors)

			return cors
		}(),
		AllowMethods: []string{
			http.MethodOptions,
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPatch,
			http.MethodPut,
			http.MethodDelete,
		},
		AllowHeaders: []string{
			echo.HeaderAuthorization,
			echo.HeaderCookie,
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAcceptEncoding,
			echo.HeaderConnection,
			echo.HeaderCacheControl,
			"Host",
			"User-Agent",
			"X-CSRF-Header",
		},
		AllowCredentials: true,
	})
}

func requestID() echo.MiddlewareFunc {
	return middleware.RequestIDWithConfig(middleware.RequestIDConfig{
		Generator: lcontext.CreateRequestID,
		RequestIDHandler: func(ctx echo.Context, s string) {
			ctx.Set(lcontext.ContextKeyRequestID.String(), s)
			ctx.SetRequest(ctx.Request().Clone(context.WithValue(ctx.Request().Context(), lcontext.ContextKeyRequestID, s)))
		},
	})
}

func requestLog() echo.MiddlewareFunc {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			req := c.Request()
			res := c.Response()

			userID := func() string {
				session, err := lsession.GetLoginSession(c)

				if err == nil {
					return session.ID.String()
				}

				return "-"
			}

			if v.Error == nil {
				logger.LogAttrs(c.Request().Context(), slog.LevelInfo, "REQUEST",
					slog.String(lcontext.ContextKeyRequestID.String(), res.Header().Get(echo.HeaderXRequestID)),
					slog.String("user_id", userID()),
					slog.Int("status", res.Status),
					slog.String("method", req.Method),
					slog.String("uri", v.URI),
					slog.String("latency", fmt.Sprintf("%dms", time.Now().UnixMilli()-v.StartTime.UnixMilli())),
					slog.String("remote_ip", c.RealIP()),
					slog.String("referer", req.Referer()),
					slog.Int64("size", res.Size),
					slog.String("request_header", fmt.Sprintf("%+v", req.Header)),
				)
			} else {
				logger.LogAttrs(c.Request().Context(), slog.LevelError, "REQUEST",
					slog.String(lcontext.ContextKeyRequestID.String(), res.Header().Get(echo.HeaderXRequestID)),
					slog.String("user_id", userID()),
					slog.Int("status", res.Status),
					slog.String("method", req.Method),
					slog.String("uri", v.URI),
					slog.String("latency", fmt.Sprintf("%dms", time.Now().UnixMilli()-v.StartTime.UnixMilli())),
					slog.String("remote_ip", c.RealIP()),
					slog.String("referer", req.Referer()),
					slog.String("request_header", fmt.Sprintf("%+v", req.Header)),
					slog.String("err", fmt.Sprintf("%+v", v.Error)),
				)
			}

			return nil
		},
	})
}

func timeout() echo.MiddlewareFunc {
	timeoutSeconds := func() int {
		timeoutSeconds, err := strconv.Atoi(os.Getenv("TIMEOUT_SECONDS_HTTP"))

		if err != nil {
			return 5
		}

		return timeoutSeconds
	}()

	return middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Skipper: func(c echo.Context) bool {
			if upgrade, ok := c.Request().Header["Upgrade"]; ok {
				if slices.Contains(upgrade, "websocket") {
					return true
				}
			}

			return false
		},
		OnTimeoutRouteErrorHandler: func(err error, c echo.Context) {},
		Timeout:                    time.Second * time.Duration(timeoutSeconds),
	})
}

func validator() echo.MiddlewareFunc {
	swagger, _ := api.GetSwagger()
	swagger.Servers = nil // serversの妥当性は検証しない

	newPaths := openapi3.NewPathsWithCapacity(0)

	for k, v := range swagger.Paths.Map() {
		newPaths.Set(os.Getenv("ROUTER_GROUP")+k, v)
	}

	swagger.Paths = newPaths

	options := echomiddleware.Options{
		ErrorHandler: func(c echo.Context, err *echo.HTTPError) error {
			// err.Message = "invalid request"
			return err
		},
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error { return nil },
		},
		Skipper: func(c echo.Context) bool {
			for _, contentType := range c.Request().Header[echo.HeaderContentType] {
				if strings.Contains(contentType, "multipart/form-data") { // Middlewareで検証しない
					return true
				}
			}

			return false
		},
	}

	return echomiddleware.OapiRequestValidatorWithOptions(swagger, &options)
}

func validatorForV1() echo.MiddlewareFunc {
	swagger, _ := apiv1.GetSwagger()
	swagger.Servers = nil // serversの妥当性は検証しない

	newPaths := openapi3.NewPathsWithCapacity(0)

	for k, v := range swagger.Paths.Map() {
		newPaths.Set(os.Getenv("ROUTER_GROUP_V1")+k, v)
	}

	swagger.Paths = newPaths

	options := echomiddleware.Options{
		ErrorHandler: func(c echo.Context, err *echo.HTTPError) error {
			err.Message = "invalid request"
			return err
		},
		Options: openapi3filter.Options{
			AuthenticationFunc: func(ctx context.Context, ai *openapi3filter.AuthenticationInput) error { return nil },
		},
		Skipper: func(c echo.Context) bool {
			for _, contentType := range c.Request().Header[echo.HeaderContentType] {
				if strings.Contains(contentType, "multipart/form-data") { // Middlewareで検証しない
					return true
				}
			}

			return false
		},
	}

	return echomiddleware.OapiRequestValidatorWithOptions(swagger, &options)
}

func auth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, err := lsession.GetLoginSession(c); err != nil {
				return c.Redirect(http.StatusFound, lgoogle.GetAuthURL())
			}

			return next(c)
		}
	}
}
