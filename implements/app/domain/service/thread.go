package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type ThreadService interface {
	Create(c context.Context, thread model.Thread, topicID uuid.UUID) error
	Get(c context.Context, id uuid.UUID) (*model.Thread, error)
	ListByTopic(c context.Context, topicID uuid.UUID, page model.Range) ([]model.Thread, error)
}

type threadService struct {
	threadRepository repository.ThreadRepository
}

// Create implements ThreadService.
func (t *threadService) Create(c context.Context, thread model.Thread, topicID uuid.UUID) error {
	return t.threadRepository.Create(c, thread, topicID)
}

// Get implements ThreadService.
func (t *threadService) Get(c context.Context, id uuid.UUID) (*model.Thread, error) {
	return t.threadRepository.Get(c, id)
}

// ListByTopic implements ThreadService.
func (t *threadService) ListByTopic(c context.Context, topicID uuid.UUID, page model.Range) ([]model.Thread, error) {
	return t.threadRepository.ListByTopic(c, topicID, page)
}

func NewThreadService(i *do.Injector) (ThreadService, error) {
	threadRepository := do.MustInvoke[repository.ThreadRepository](i)
	return &threadService{threadRepository: threadRepository}, nil
}
