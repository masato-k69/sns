package implement

import (
	"app/gen/pubsub"
	"app/usecase/service"
	"context"
	"encoding/json"

	"github.com/google/uuid"
)

type resourceSearchIndexCreateHandler struct {
	usecase service.ResourceSearchIndexUsecase
}

// Handle implements subscriber.Handler.
func (r resourceSearchIndexCreateHandler) Handle(c context.Context, message []byte) error {
	var m pubsub.ResourceSearchIndex
	if err := json.Unmarshal(message, &m); err != nil {
		return err
	}

	resourceID, err := uuid.Parse(m.ResourceId.Value)
	if err != nil {
		return err
	}

	return r.usecase.Create(c, resourceID, m.ResourceType.Value, m.Keyword)
}

type resourceSearchIndexUpdateHandler struct {
	usecase service.ResourceSearchIndexUsecase
}

// Handle implements subscriber.Handler.
func (r resourceSearchIndexUpdateHandler) Handle(c context.Context, message []byte) error {
	var m pubsub.ResourceSearchIndex
	if err := json.Unmarshal(message, &m); err != nil {
		return err
	}

	resourceID, err := uuid.Parse(m.ResourceId.Value)
	if err != nil {
		return err
	}

	return r.usecase.Update(c, resourceID, m.ResourceType.Value, m.Keyword)
}

type resourceSearchIndexDeleteHandler struct {
	usecase service.ResourceSearchIndexUsecase
}

// Handle implements subscriber.Handler.
func (r resourceSearchIndexDeleteHandler) Handle(c context.Context, message []byte) error {
	var m pubsub.ResourceSearchIndex
	if err := json.Unmarshal(message, &m); err != nil {
		return err
	}

	resourceID, err := uuid.Parse(m.ResourceId.Value)
	if err != nil {
		return err
	}

	return r.usecase.Delete(c, resourceID)
}
