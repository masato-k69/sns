package session

import (
	"app/lib/environment"
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"
)

type LoggedInUser struct {
	ID    uuid.UUID
	Email string
	Name  string
}

const (
	SessionKey      = "session"
	SessionKeyID    = "id"
	SessionKeyEmail = "email"
	SessionKeyName  = "name"
)

var (
	sessionStore *redisstore.RedisStore
)

func Init() error {
	sessionStoreDB, err := strconv.Atoi(os.Getenv("SESSION_STORE_DB"))

	if err != nil {
		return errors.Wrap(err, "failed to parse session db")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("SESSION_STORE_HOST"), os.Getenv("SESSION_STORE_PORT")),
		Password: os.Getenv("SESSION_STORE_PASSWORD"),
		DB:       sessionStoreDB,
	})

	c := context.Background()

	if err := client.Set(c, os.Getenv("SESSION_STORE_HOST"), "", time.Second).Err(); err != nil {
		return errors.Wrap(err, "failed to connect session store")
	}

	store, err := redisstore.NewRedisStore(c, client)

	if err != nil {
		return errors.Wrap(err, "failed to create session store")
	}

	store.KeyPrefix(os.Getenv("SESSION_KEY_PREFIX"))

	sessionStore = store

	fmt.Printf("connection established. %v:%v \n", os.Getenv("SESSION_STORE_HOST"), os.Getenv("SESSION_STORE_PORT"))

	return nil
}

func SetLoginSession(c echo.Context, id string, email string, name string, expire time.Time) error {
	session, err := sessionStore.Get(c.Request(), SessionKey)

	if err != nil {
		return err
	}

	session.Values[SessionKeyID] = id
	session.Values[SessionKeyEmail] = email
	session.Values[SessionKeyName] = name

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(expire.Unix() - time.Now().Unix()),
		HttpOnly: true,
		Secure:   !environment.IsDebug(),
	}

	return sessionStore.Save(c.Request(), c.Response(), session)
}

func GetLoginSession(c echo.Context) (*LoggedInUser, error) {
	session, err := sessionStore.Get(c.Request(), SessionKey)

	if err != nil {
		return nil, err
	}

	id, ok := session.Values[SessionKeyID].(string)

	if !ok {
		return nil, fmt.Errorf("%v is not set on the session", SessionKeyID)
	}

	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, fmt.Errorf("%v is not set on the session", SessionKeyID)
	}

	email, ok := session.Values[SessionKeyEmail].(string)

	if !ok {
		return nil, fmt.Errorf("%v is not set on the session", SessionKeyEmail)
	}

	name, ok := session.Values[SessionKeyName].(string)

	if !ok {
		return nil, fmt.Errorf("%v is not set on the session", SessionKeyName)
	}

	return &LoggedInUser{
		ID:    parsedID,
		Email: email,
		Name:  name,
	}, nil
}

func DeleteLoginSession(c echo.Context) error {
	session, err := sessionStore.Get(c.Request(), SessionKey)

	if err != nil {
		return err
	}

	session.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !environment.IsDebug(),
	}

	return sessionStore.Save(c.Request(), c.Response(), session)
}

func SessionStore() echo.MiddlewareFunc {
	return session.Middleware(sessionStore)
}

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if _, err := GetLoginSession(c); err != nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			return next(c)
		}
	}
}
