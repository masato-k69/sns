package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type MemberRepository interface {
	Create(c context.Context, member model.Member, mention model.Mention) error
	Get(c context.Context, id uuid.UUID) (*model.Member, error)
	GetJoinedCommunityID(c context.Context, memberID uuid.UUID) (*uuid.UUID, error)
	GetByCommunityAndUser(c context.Context, communityID uuid.UUID, userID uuid.UUID) (*model.Member, error)
	ListByCommunity(c context.Context, communityID uuid.UUID, page model.Range) ([]model.Member, error)
	ListByUser(c context.Context, userID uuid.UUID) ([]model.Member, error)
}
