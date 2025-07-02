package lock

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
)

func Init() error {
	sessionStoreDB, err := strconv.Atoi(os.Getenv("LOCK_STORE_DB"))

	if err != nil {
		return errors.Wrap(err, "failed to parse session db")
	}

	c := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("LOCK_STORE_HOST"), os.Getenv("LOCK_STORE_PORT")),
		Password: os.Getenv("LOCK_STORE_PASSWORD"),
		DB:       sessionStoreDB,
	})

	if err := c.Set(context.Background(), os.Getenv("LOCK_STORE_HOST"), "", time.Second).Err(); err != nil {
		return errors.Wrap(err, "failed to connect session store")
	}

	client = c

	return nil
}

func Set(c context.Context, key uuid.UUID, ttl time.Duration) error {
	if err := client.Del(c, key.String()).Err(); err != nil {
		if !errors.Is(err, redis.Nil) {
			return err
		}
	}

	return client.Set(context.Background(), key.String(), "", ttl).Err()
}

func Exist(c context.Context, key uuid.UUID) (bool, error) {
	if _, err := client.Get(c, key.String()).Result(); err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func Delete(c context.Context, key uuid.UUID) error {
	if err := client.Del(c, key.String()).Err(); err != nil {
		if !errors.Is(err, redis.Nil) {
			return err
		}
	}

	return nil
}
