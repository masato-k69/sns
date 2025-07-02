package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type MemberService interface {
	Create(c context.Context, member model.Member, mention model.Mention) error
	Get(c context.Context, id uuid.UUID) (*model.Member, error)
	GetJoinedCommunityID(c context.Context, memberID uuid.UUID) (*uuid.UUID, error)
	GetByCommunityAndUser(c context.Context, communityID uuid.UUID, userID uuid.UUID) (*model.Member, error)
	ListByCommunity(c context.Context, communityID uuid.UUID, page model.Range) ([]model.Member, error)
	ListByUser(c context.Context, userID uuid.UUID) ([]model.Member, error)
}

type memberService struct {
	memberRepository repository.MemberRepository
}

// GetJoinedCommunityID implements MemberService.
func (m *memberService) GetJoinedCommunityID(c context.Context, memberID uuid.UUID) (*uuid.UUID, error) {
	return m.memberRepository.GetJoinedCommunityID(c, memberID)
}

// ListByUser implements MemberService.
func (m *memberService) ListByUser(c context.Context, userID uuid.UUID) ([]model.Member, error) {
	return m.memberRepository.ListByUser(c, userID)
}

// GetByCommunityAndUser implements MemberService.
func (m *memberService) GetByCommunityAndUser(c context.Context, communityID uuid.UUID, userID uuid.UUID) (*model.Member, error) {
	return m.memberRepository.GetByCommunityAndUser(c, communityID, userID)
}

// Get implements MemberService.
func (m *memberService) Get(c context.Context, id uuid.UUID) (*model.Member, error) {
	return m.memberRepository.Get(c, id)
}

// ListByCommunity implements MemberService.
func (m *memberService) ListByCommunity(c context.Context, communityID uuid.UUID, page model.Range) ([]model.Member, error) {
	return m.memberRepository.ListByCommunity(c, communityID, page)
}

// Create implements MemberService.
func (m *memberService) Create(c context.Context, member model.Member, mention model.Mention) error {
	return m.memberRepository.Create(c, member, mention)
}

func NewMemberService(i *do.Injector) (MemberService, error) {
	memberRepository := do.MustInvoke[repository.MemberRepository](i)
	return &memberService{memberRepository: memberRepository}, nil
}
