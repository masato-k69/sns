package repository

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	drepository "app/domain/repository"
	imodel "app/infrastructure/model"
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"gorm.io/gorm"

	irdb "app/infrastructure/adapter/datastore/rdb"
)

type contentRepository struct {
	contentStoreConnection irdb.ContentStoreConnection
}

// Create implements repository.ContentRepository.
func (co *contentRepository) Create(c context.Context, contents []dmodel.Content, mention dmodel.Mention) error {
	return co.contentStoreConnection.Write().Transaction(func(tx *gorm.DB) error {
		for _, newContent := range contents {
			if err := tx.
				Create(&imodel.Content{
					ID:   newContent.ID.String(),
					Type: newContent.Type.String(),
					Bin:  newContent.Value,
				}).Error; err != nil {
				return errors.Wrapf(err, "failed to create content. id=%v", newContent.ID)
			}

			switch mention.Resource {
			case dmodel.ResourceLine:
				if err := tx.
					Create(&imodel.ContentLineRelation{
						ContentID: newContent.ID.String(),
						LineID:    mention.ID.String(),
					}).Error; err != nil {
					return errors.Wrapf(err, "failed to create line relation. id=%v", mention.ID.String())
				}
			case dmodel.ResourceTopic:
				if err := tx.
					Create(&imodel.ContentTopicRelation{
						ContentID: newContent.ID.String(),
						TopicID:   mention.ID.String(),
					}).Error; err != nil {
					return errors.Wrapf(err, "failed to create topic relation. id=%v", mention.ID.String())
				}
			case dmodel.ResourcePost:
				if err := tx.
					Create(&imodel.ContentPostRelation{
						ContentID: newContent.ID.String(),
						PostID:    mention.ID.String(),
					}).Error; err != nil {
					return errors.Wrapf(err, "failed to create post relation. id=%v", mention.ID.String())
				}
			}
		}

		return nil
	})
}

// DeleteByResource implements repository.ContentRepository.
func (co *contentRepository) DeleteByResource(c context.Context, mention dmodel.Mention) error {
	contents := []imodel.Content{}

	switch mention.Resource {
	case dmodel.ResourceLine:
		if err := co.contentStoreConnection.Write().
			Model(&imodel.Content{}).
			Select("contents.id as id, contents.type as type, contents.bin as bin").
			Joins("inner join content_line_relations on contents.id = content_line_relations.content_id").
			Where("content_line_relations.line_id = ?", mention.ID.String()).
			Order("contents.created_at asc").
			Scan(&contents).Error; err != nil {
			return errors.Wrapf(err, "failed to list content. line_id=%v", mention.ID.String())
		}
	case dmodel.ResourceTopic:
		if err := co.contentStoreConnection.Write().
			Model(&imodel.Content{}).
			Select("contents.id as id, contents.type as type, contents.bin as bin").
			Joins("inner join content_topic_relations on contents.id = content_topic_relations.content_id").
			Where("content_topic_relations.topic_id = ?", mention.ID.String()).
			Order("contents.created_at asc").
			Scan(&contents).Error; err != nil {
			return errors.Wrapf(err, "failed to list content. topic_id=%v", mention.ID.String())
		}
	case dmodel.ResourcePost:
		if err := co.contentStoreConnection.Write().
			Model(&imodel.Content{}).
			Select("contents.id as id, contents.type as type, contents.bin as bin").
			Joins("inner join content_post_relations on contents.id = content_post_relations.content_id").
			Where("content_post_relations.post_id = ?", mention.ID.String()).
			Order("contents.created_at asc").
			Scan(&contents).Error; err != nil {
			return errors.Wrapf(err, "failed to list content. post_id=%v", mention.ID.String())
		}
	}

	return co.contentStoreConnection.Write().
		Delete(&contents).Error
}

// ListByLine implements repository.ContentRepository.
func (co *contentRepository) ListByLine(c context.Context, lineID uuid.UUID) ([]dmodel.Content, error) {
	contents := []imodel.Content{}
	if err := co.contentStoreConnection.Read().
		Model(&imodel.Content{}).
		Select("contents.id as id, contents.type as type, contents.bin as bin").
		Joins("inner join content_line_relations on contents.id = content_line_relations.content_id").
		Where("content_line_relations.line_id = ?", lineID.String()).
		Order("contents.created_at asc").
		Scan(&contents).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list content. line_id=%v", lineID.String())
	}

	dContents := []dmodel.Content{}
	for _, content := range contents {
		dContent, err := dfactory.NewContent(content.ID, content.Type, content.Bin)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse content. id=%v", content.ID)
		}

		dContents = append(dContents, *dContent)
	}

	return dContents, nil
}

// ListByPost implements repository.ContentRepository.
func (co *contentRepository) ListByPost(c context.Context, postID uuid.UUID) ([]dmodel.Content, error) {
	contents := []imodel.Content{}
	if err := co.contentStoreConnection.Read().
		Model(&imodel.Content{}).
		Select("contents.id as id, contents.type as type, contents.bin as bin").
		Joins("inner join content_post_relations on contents.id = content_post_relations.content_id").
		Where("content_post_relations.post_id = ?", postID.String()).
		Order("contents.created_at asc").
		Scan(&contents).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list content. post_id=%v", postID.String())
	}

	dContents := []dmodel.Content{}
	for _, content := range contents {
		dContent, err := dfactory.NewContent(content.ID, content.Type, content.Bin)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse content. id=%v", content.ID)
		}

		dContents = append(dContents, *dContent)
	}

	return dContents, nil
}

// ListByTopic implements repository.ContentRepository.
func (co *contentRepository) ListByTopic(c context.Context, topicID uuid.UUID) ([]dmodel.Content, error) {
	contents := []imodel.Content{}
	if err := co.contentStoreConnection.Read().
		Model(&imodel.Content{}).
		Select("contents.id as id, contents.type as type, contents.bin as bin").
		Joins("inner join content_topic_relations on contents.id = content_topic_relations.content_id").
		Where("content_topic_relations.topic_id = ?", topicID.String()).
		Order("contents.created_at asc").
		Scan(&contents).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list content. topic_id=%v", topicID.String())
	}

	dContents := []dmodel.Content{}
	for _, content := range contents {
		dContent, err := dfactory.NewContent(content.ID, content.Type, content.Bin)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse content. id=%v", content.ID)
		}

		dContents = append(dContents, *dContent)
	}

	return dContents, nil
}

// DeleteAndCreate implements repository.ContentRepository.
func (co *contentRepository) DeleteAndCreate(c context.Context, newContents []dmodel.Content, mention dmodel.Mention) error {
	return co.contentStoreConnection.Write().Transaction(func(tx *gorm.DB) error {
		currentContents := []imodel.Content{}
		switch mention.Resource {
		case dmodel.ResourceLine:
			if err := tx.
				Model(&imodel.Content{}).
				Select("contents.id as id, contents.type as type, contents.bin as bin").
				Joins("inner join content_line_relations on contents.id = content_line_relations.content_id").
				Where("content_line_relations.line_id = ?", mention.ID.String()).
				Order("contents.created_at asc").
				Scan(&currentContents).Error; err != nil {
				return errors.Wrapf(err, "failed to list content. line_id=%v", mention.ID.String())
			}
		case dmodel.ResourceTopic:
			if err := tx.
				Model(&imodel.Content{}).
				Select("contents.id as id, contents.type as type, contents.bin as bin").
				Joins("inner join content_topic_relations on contents.id = content_topic_relations.content_id").
				Where("content_topic_relations.topic_id = ?", mention.ID.String()).
				Order("contents.created_at asc").
				Scan(&currentContents).Error; err != nil {
				return errors.Wrapf(err, "failed to list content. topic_id=%v", mention.ID.String())
			}
		case dmodel.ResourcePost:
			if err := tx.
				Model(&imodel.Content{}).
				Select("contents.id as id, contents.type as type, contents.bin as bin").
				Joins("inner join content_post_relations on contents.id = content_post_relations.content_id").
				Where("content_post_relations.post_id = ?", mention.ID.String()).
				Order("contents.created_at asc").
				Scan(&currentContents).Error; err != nil {
				return errors.Wrapf(err, "failed to list content. post_id=%v", mention.ID.String())
			}
		}

		if err := tx.
			Delete(&currentContents).Error; err != nil {
			return errors.Wrapf(err, "failed to delete content. id=%v", mention.ID.String())
		}

		for _, newContent := range newContents {
			if err := tx.
				Create(&imodel.Content{
					ID:   newContent.ID.String(),
					Type: newContent.Type.String(),
					Bin:  newContent.Value,
				}).Error; err != nil {
				return errors.Wrapf(err, "failed to create content. id=%v", newContent.ID)
			}

			switch mention.Resource {
			case dmodel.ResourceLine:
				if err := tx.
					Create(&imodel.ContentLineRelation{
						ContentID: newContent.ID.String(),
						LineID:    mention.ID.String(),
					}).Error; err != nil {
					return errors.Wrapf(err, "failed to create line relation. id=%v", mention.ID.String())
				}
			case dmodel.ResourceTopic:
				if err := tx.
					Create(&imodel.ContentTopicRelation{
						ContentID: newContent.ID.String(),
						TopicID:   mention.ID.String(),
					}).Error; err != nil {
					return errors.Wrapf(err, "failed to create topic relation. id=%v", mention.ID.String())
				}
			case dmodel.ResourcePost:
				if err := tx.
					Create(&imodel.ContentPostRelation{
						ContentID: newContent.ID.String(),
						PostID:    mention.ID.String(),
					}).Error; err != nil {
					return errors.Wrapf(err, "failed to create post relation. id=%v", mention.ID.String())
				}
			}
		}

		return nil
	})
}

// Delete implements repository.ContentRepository.
func (co *contentRepository) Delete(c context.Context, ids []uuid.UUID) error {
	parsedIDs := []string{}

	for _, id := range ids {
		parsedIDs = append(parsedIDs, id.String())
	}

	return co.contentStoreConnection.Read().
		Where("id in ?", parsedIDs).
		Delete(&[]imodel.Content{}).Error
}

// List implements repository.ContentRepository.
func (co *contentRepository) List(c context.Context, ids []uuid.UUID) ([]dmodel.Content, error) {
	parsedIDs := []string{}

	for _, id := range ids {
		parsedIDs = append(parsedIDs, id.String())
	}

	contents := []imodel.Content{}

	if err := co.contentStoreConnection.Read().
		Where("id in ?", parsedIDs).
		Find(&contents).Error; err != nil {
		return nil, errors.Wrapf(err, "fialed to get contents. note_id=%v", parsedIDs)
	}

	dContents := []dmodel.Content{}

	for _, content := range contents {
		dContent, err := dfactory.NewContent(content.ID, content.Type, content.Bin)

		if err != nil {
			return nil, errors.Wrapf(err, "fialed to parse conetnt. id=%v", content.ID)
		}

		dContents = append(dContents, *dContent)
	}

	return dContents, nil
}

func NewContentRepository(i *do.Injector) (drepository.ContentRepository, error) {
	contentStoreConnection := do.MustInvoke[irdb.ContentStoreConnection](i)
	return &contentRepository{
		contentStoreConnection: contentStoreConnection,
	}, nil
}
