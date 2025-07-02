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

type topicRepository struct {
	topicStoreConnection irdb.TopicStoreConnection
}

// Create implements repository.TopicRepository.
func (t *topicRepository) Create(c context.Context, topic dmodel.Topic, communityID uuid.UUID) error {
	return t.topicStoreConnection.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Create(&imodel.Topic{
				ID:   topic.ID.String(),
				Name: topic.Name.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create topic. id=%v", topic.ID.String())
		}

		if err := tx.
			Create(&imodel.TopicCommunityRelation{
				TopicID:     topic.ID.String(),
				CommunityID: communityID.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create community relation. topic_id=%v", topic.ID.String())
		}

		if err := tx.
			Create(&imodel.TopicFromMemberRelation{
				TopicID:  topic.ID.String(),
				MemberID: topic.Created.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create from member relation. topic_id=%v", topic.ID.String())
		}

		return nil
	})
}

// Get implements repository.TopicRepository.
func (t *topicRepository) Get(c context.Context, id uuid.UUID) (*dmodel.Topic, error) {
	iTopic := imodel.Topic{ID: id.String()}
	if err := t.topicStoreConnection.Read().
		First(&iTopic).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get topic. id=%v", id.String())
	}

	var created *string
	memberRelation := imodel.TopicFromMemberRelation{TopicID: iTopic.ID}
	if err := t.topicStoreConnection.Read().
		First(&memberRelation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			created = nil
		} else {
			return nil, errors.Wrapf(err, "failed to create from member relation. topic_id=%v", iTopic.ID)
		}
	} else {
		created = &memberRelation.MemberID
	}

	return dfactory.NewTopic(iTopic.ID, iTopic.Name, created)
}

// ListByCommunity implements repository.TopicRepository.
func (t *topicRepository) ListByCommunity(c context.Context, communityID uuid.UUID, page dmodel.Range) ([]dmodel.Topic, error) {
	iTopics := []imodel.Topic{}
	if err := t.topicStoreConnection.Read().
		Model(&imodel.Topic{}).
		Select("topics.id as id, topics.name as name").
		Joins("inner join topic_community_relations on topics.id = topic_community_relations.topic_id").
		Where("topic_community_relations.community_id = ?", communityID.String()).
		Order("topics.created_at asc").
		Limit(page.Limit).Offset(page.Offset).
		Scan(&iTopics).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list topic. community_id=%v", communityID.String())
	}

	dTopics := []dmodel.Topic{}
	for _, iTopic := range iTopics {
		var created *string
		memberRelation := imodel.TopicFromMemberRelation{TopicID: iTopic.ID}
		if err := t.topicStoreConnection.Read().
			First(&memberRelation).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				created = nil
			} else {
				return nil, errors.Wrapf(err, "failed to get from member relation. topic_id=%v", iTopic.ID)
			}
		} else {
			created = &memberRelation.MemberID
		}

		dTopic, err := dfactory.NewTopic(iTopic.ID, iTopic.Name, created)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse topic. id=%v", iTopic.ID)
		}

		dTopics = append(dTopics, *dTopic)
	}

	return dTopics, nil
}

func NewTopicRepository(i *do.Injector) (drepository.TopicRepository, error) {
	topicStoreConnection := do.MustInvoke[irdb.TopicStoreConnection](i)
	return &topicRepository{
		topicStoreConnection: topicStoreConnection,
	}, nil
}
