package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type TopicService interface {
	Create(c context.Context, topic model.Topic, communityID uuid.UUID) error
	Get(c context.Context, id uuid.UUID) (*model.Topic, error)
	ListByCommunity(c context.Context, communityID uuid.UUID, page model.Range) ([]model.Topic, error)
}

type topicService struct {
	topicRepository repository.TopicRepository
}

// Create implements TopicService.
func (t *topicService) Create(c context.Context, topic model.Topic, communityID uuid.UUID) error {
	return t.topicRepository.Create(c, topic, communityID)
}

// Get implements TopicService.
func (t *topicService) Get(c context.Context, id uuid.UUID) (*model.Topic, error) {
	return t.topicRepository.Get(c, id)
}

// ListByCommunity implements TopicService.
func (t *topicService) ListByCommunity(c context.Context, communityID uuid.UUID, page model.Range) ([]model.Topic, error) {
	return t.topicRepository.ListByCommunity(c, communityID, page)
}

func NewTopicService(i *do.Injector) (TopicService, error) {
	topicRepository := do.MustInvoke[repository.TopicRepository](i)
	return &topicService{topicRepository: topicRepository}, nil
}
