package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type ContentService interface {
	Create(c context.Context, contents []model.Content, mention model.Mention) error
	DeleteAndCreate(c context.Context, contents []model.Content, mention model.Mention) error
	List(c context.Context, ids []uuid.UUID) ([]model.Content, error)
	ListByTopic(c context.Context, topicID uuid.UUID) ([]model.Content, error)
	ListByLine(c context.Context, lineID uuid.UUID) ([]model.Content, error)
	ListByPost(c context.Context, postID uuid.UUID) ([]model.Content, error)
	DeleteByResource(c context.Context, mention model.Mention) error
}

type contentService struct {
	contentRepository repository.ContentRepository
}

// Create implements ContentService.
func (co *contentService) Create(c context.Context, contents []model.Content, mention model.Mention) error {
	return co.contentRepository.Create(c, contents, mention)
}

// DeleteByResource implements ContentService.
func (co *contentService) DeleteByResource(c context.Context, mention model.Mention) error {
	return co.contentRepository.DeleteByResource(c, mention)
}

// ListByLine implements ContentService.
func (co *contentService) ListByLine(c context.Context, lineID uuid.UUID) ([]model.Content, error) {
	return co.contentRepository.ListByLine(c, lineID)
}

// ListByPost implements ContentService.
func (co *contentService) ListByPost(c context.Context, postID uuid.UUID) ([]model.Content, error) {
	return co.contentRepository.ListByPost(c, postID)
}

// ListByTopic implements ContentService.
func (co *contentService) ListByTopic(c context.Context, topicID uuid.UUID) ([]model.Content, error) {
	return co.contentRepository.ListByTopic(c, topicID)
}

// DeleteAndCreate implements ContentService.
func (co *contentService) DeleteAndCreate(c context.Context, contents []model.Content, mention model.Mention) error {
	return co.contentRepository.DeleteAndCreate(c, contents, mention)
}

// List implements ContentService.
func (co *contentService) List(c context.Context, ids []uuid.UUID) ([]model.Content, error) {
	return co.contentRepository.List(c, ids)
}

func NewContentService(i *do.Injector) (ContentService, error) {
	contentRepository := do.MustInvoke[repository.ContentRepository](i)
	return &contentService{contentRepository: contentRepository}, nil
}
