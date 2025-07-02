package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type PostService interface {
	Create(c context.Context, post model.Post, topicID uuid.UUID, threadID uuid.UUID) error
	Get(c context.Context, id uuid.UUID) (*model.Post, error)
	Last(c context.Context, topicID uuid.UUID) (*model.Post, error)
	ListByThread(c context.Context, threadID uuid.UUID, page model.Range) ([]model.Post, error)
}

type postService struct {
	postRepository repository.PostRepository
}

// Last implements PostService.
func (p *postService) Last(c context.Context, topicID uuid.UUID) (*model.Post, error) {
	return p.postRepository.Last(c, topicID)
}

// Create implements PostService.
func (p *postService) Create(c context.Context, post model.Post, topicID uuid.UUID, threadID uuid.UUID) error {
	return p.postRepository.Create(c, post, topicID, threadID)
}

// Get implements PostService.
func (p *postService) Get(c context.Context, id uuid.UUID) (*model.Post, error) {
	return p.postRepository.Get(c, id)
}

// ListByThread implements PostService.
func (p *postService) ListByThread(c context.Context, threadID uuid.UUID, page model.Range) ([]model.Post, error) {
	return p.postRepository.ListByThread(c, threadID, page)
}

func NewPostService(i *do.Injector) (PostService, error) {
	postRepository := do.MustInvoke[repository.PostRepository](i)
	return &postService{postRepository: postRepository}, nil
}
