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

type postRepository struct {
	postStoreConnection irdb.PostStoreConnection
}

// Last implements repository.PostRepository.
func (p *postRepository) Last(c context.Context, topicID uuid.UUID) (*dmodel.Post, error) {
	iPosts := []imodel.Post{}
	if err := p.postStoreConnection.Read().
		Model(&imodel.Post{}).
		Select("posts.id as id, posts.at as at").
		Joins("inner join post_topic_relations on posts.id = post_topic_relations.post_id").
		Where("post_topic_relations.topic_id = ?", topicID.String()).
		Order("posts.created_at desc").
		Limit(1).
		Scan(&iPosts).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list post. topic_id=%v", topicID.String())
	}

	if len(iPosts) < 1 {
		return nil, nil
	}

	dPost, err := p.toPost(iPosts[0])
	if err != nil {
		return nil, err
	}

	return dPost, nil
}

// Create implements repository.PostRepository.
func (p *postRepository) Create(c context.Context, post dmodel.Post, topicID uuid.UUID, threadID uuid.UUID) error {
	return p.postStoreConnection.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Create(&imodel.Post{
				ID: post.ID.String(),
				At: post.At.Int(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create post. id=%v", post.ID.String())
		}

		if err := tx.
			Create(&imodel.PostTopicRelation{
				PostID:  post.ID.String(),
				TopicID: topicID.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create topic relation. topic_id=%v", topicID.String())
		}

		if err := tx.
			Create(&imodel.PostThreadRelation{
				PostID:   post.ID.String(),
				ThreadID: threadID.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create thread relation. thread_id=%v", threadID.String())
		}

		if err := tx.
			Create(&imodel.PostFromMemberRelation{
				PostID:   post.ID.String(),
				MemberID: post.From.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create from relation. member_id=%v", post.From.String())
		}

		for _, mention := range post.To {
			switch mention.Resource {
			case dmodel.ResourceMember:
				if err := tx.
					Create(&imodel.PostToMemberRelation{
						PostID:   post.ID.String(),
						MemberID: mention.ID.String(),
					}).Error; err != nil {
					return errors.Wrapf(err, "failed to create to member relation. member_id=%v ", mention.ID.String())
				}
			case dmodel.ResourceRole:
				if err := tx.
					Create(&imodel.PostToRoleRelation{
						PostID: post.ID.String(),
						RoleID: mention.ID.String(),
					}).Error; err != nil {
					return errors.Wrapf(err, "failed to create to role relation. role_id=%v", mention.ID.String())
				}
			}
		}

		return nil
	})
}

// Get implements repository.PostRepository.
func (p *postRepository) Get(c context.Context, id uuid.UUID) (*dmodel.Post, error) {
	iPost := imodel.Post{ID: id.String()}
	if err := p.postStoreConnection.Read().
		First(&iPost).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "failed to get post. id=%v", id.String())
	}

	return p.toPost(iPost)
}

// ListByThread implements repository.PostRepository.
func (p *postRepository) ListByThread(c context.Context, threadID uuid.UUID, page dmodel.Range) ([]dmodel.Post, error) {
	iPosts := []imodel.Post{}
	if err := p.postStoreConnection.Read().
		Model(&imodel.Post{}).
		Select("posts.id as id, posts.at as at").
		Joins("inner join post_thread_relations on posts.id = post_thread_relations.post_id").
		Where("post_thread_relations.thread_id = ?", threadID.String()).
		Order("posts.created_at asc").
		Limit(page.Limit).Offset(page.Offset).
		Scan(&iPosts).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list post. thread_id=%v", threadID.String())
	}

	dPosts := []dmodel.Post{}
	for _, iPost := range iPosts {
		dPost, err := p.toPost(iPost)
		if err != nil {
			return nil, err
		}

		dPosts = append(dPosts, *dPost)
	}

	return dPosts, nil
}

func (p *postRepository) toPost(post imodel.Post) (*dmodel.Post, error) {
	var posted *string
	memberRelation := imodel.PostFromMemberRelation{PostID: post.ID}
	if err := p.postStoreConnection.Read().
		First(&memberRelation).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			posted = nil
		} else {
			return nil, errors.Wrapf(err, "failed to get from member relation. post_id=%v", post.ID)
		}
	} else {
		posted = &memberRelation.MemberID
	}

	iPostToMembers := []imodel.PostToMemberRelation{}
	if err := p.postStoreConnection.Read().
		Where("post_id = ?", post.ID).
		Find(&iPostToMembers).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list to member relation. post_id=%v", post.ID)
	}

	iPostToRoles := []imodel.PostToRoleRelation{}
	if err := p.postStoreConnection.Read().
		Where("post_id = ?", post.ID).
		Find(&iPostToRoles).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to list to role relation. post_id=%v", post.ID)
	}

	dPostTo := []dmodel.Mention{}
	for _, iPostTomember := range iPostToMembers {
		mention, err := dmodel.NewMention(iPostTomember.MemberID, dmodel.ResourceMember.String())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse mention. member_id=%v", iPostTomember.MemberID)
		}

		dPostTo = append(dPostTo, *mention)
	}
	for _, iPostToRole := range iPostToRoles {
		mention, err := dmodel.NewMention(iPostToRole.RoleID, dmodel.ResourceRole.String())
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse mention. role_id=%v", iPostToRole.RoleID)
		}

		dPostTo = append(dPostTo, *mention)
	}

	return dfactory.NewPost(post.ID, posted, dPostTo, post.At)
}

func NewPostRepository(i *do.Injector) (drepository.PostRepository, error) {
	postStoreConnection := do.MustInvoke[irdb.PostStoreConnection](i)
	return &postRepository{
		postStoreConnection: postStoreConnection,
	}, nil
}
