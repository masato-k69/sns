package context

import "golang.org/x/exp/rand"

type ContextKey string

func (c ContextKey) String() string {
	return string(c)
}

const (
	ContextKeyRequestID ContextKey = "request_id"
)

func CreateRequestID() string {
	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, 12)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}

	return string(b)
}
