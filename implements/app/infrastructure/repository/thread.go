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

type threadRepository struct {
	threadStoreConnection irdb.ThreadStoreConnection
}

// Create implements repository.ThreadRepository.
func (t *threadRepository) Create(c context.Context, thread dmodel.Thread, topicID uuid.UUID) error {
	return t.threadStoreConnection.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Create(&imodel.Thread{
				ID: thread.ID.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create thread. id=%v", thread.ID.String())
		}

		if err := tx.
			Create(&imodel.ThreadTopicRelation{
				ThreadID: thread.ID.String(),
				TopicID:  topicID.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create topic_relation. thread_id=%v", thread.ID.String())
		}

		return nil
	})
}

// Get implements repository.ThreadRepository.
func (t *threadRepository) Get(c context.Context, id uuid.UUID) (*dmodel.Thread, error) {
	iThread := imodel.Thread{ID: id.String()}
	if err := t.threadStoreConnection.Read().
		First(&iThread).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get thread. id=%v", id.String())
	}

	return dfactory.NewThread(iThread.ID)
}

// ListByTopic implements repository.ThreadRepository.
func (t *threadRepository) ListByTopic(c context.Context, topicID uuid.UUID, page dmodel.Range) ([]dmodel.Thread, error) {
	iThreads := []imodel.Thread{}
	if err := t.threadStoreConnection.Read().
		Model(&imodel.Thread{}).
		Select("threads.id as id").
		Joins("inner join thread_topic_relations on threads.id = thread_topic_relations.thread_id").
		Where("thread_topic_relations.topic_id = ?", topicID.String()).
		Order("threads.created_at asc").
		Limit(page.Limit).Offset(page.Offset).
		Scan(&iThreads).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list thread. topic_id=%v", topicID.String())
	}

	dThreads := []dmodel.Thread{}
	for _, iThread := range iThreads {
		dThread, err := dfactory.NewThread(iThread.ID)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse thread. id=%v", iThread.ID)
		}

		dThreads = append(dThreads, *dThread)
	}

	return dThreads, nil
}

func NewThreadRepository(i *do.Injector) (drepository.ThreadRepository, error) {
	threadStoreConnection := do.MustInvoke[irdb.ThreadStoreConnection](i)
	return &threadRepository{
		threadStoreConnection: threadStoreConnection,
	}, nil
}
