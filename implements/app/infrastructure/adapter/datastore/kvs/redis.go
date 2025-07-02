package kvs

import (
	"app/lib/environment"
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/pkg/errors"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"
)

var (
	ctx          = context.Background()
	sessionStore *redisstore.RedisStore
)

func Init() error {
	redisDB, err := strconv.Atoi(os.Getenv("REDIS_DB"))

	if err != nil {
		return errors.Wrap(err, "failed to parse kvs db")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	})

	if err := client.Set(ctx, os.Getenv("REDIS_HOST"), "", time.Second).Err(); err != nil {
		return errors.Wrap(err, "failed to connect kvs")
	}

	store, err := redisstore.NewRedisStore(ctx, client)

	if err != nil {
		return errors.Wrap(err, "failed to create redis store")
	}

	store.KeyPrefix("session_")
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   60 * 60,
		HttpOnly: true,
		Secure:   !environment.IsDebug(),
	})

	sessionStore = store

	return nil
}

func SessionStore() *redisstore.RedisStore {
	return sessionStore
}
