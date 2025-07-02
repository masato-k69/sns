package interfaces

import "context"

type Handler interface {
	Handle(c context.Context, message []byte) error
}
