package implement

import (
	dservice "app/domain/service"
	"app/infrastructure/adapter/datastore/document"
	"app/infrastructure/adapter/datastore/rdb"
	"app/infrastructure/adapter/datastore/searchengine"
	"app/infrastructure/adapter/datastore/timeseries"
	"app/infrastructure/adapter/mq"
	"app/infrastructure/repository"
	uservice "app/usecase/service"

	"github.com/samber/do"
)

func NewHandler() Handler {
	i := do.New()

	do.Provide(i, document.NewNoteStoreConnection)
	do.Provide(i, document.NewRoleStoreConnection)
	do.Provide(i, rdb.NewNoteStoreConnection)
	do.Provide(i, rdb.NewContentStoreConnection)
	do.Provide(i, rdb.NewUserStoreConnection)
	do.Provide(i, rdb.NewRoleStoreConnection)
	do.Provide(i, rdb.NewMemberStoreConnection)
	do.Provide(i, rdb.NewCommunityStoreConnection)
	do.Provide(i, rdb.NewInviteStoreConnection)
	do.Provide(i, rdb.NewTopicStoreConnection)
	do.Provide(i, rdb.NewThreadStoreConnection)
	do.Provide(i, rdb.NewPostStoreConnection)
	do.Provide(i, mq.NewActivityStoreConnection)
	do.Provide(i, mq.NewResourceSearchIndexStoreConnection)
	do.Provide(i, searchengine.NewResourceSearchIndexStoreConnection)
	do.Provide(i, timeseries.NewActivityStore)

	do.Provide(i, repository.NewNoteRepository)
	do.Provide(i, repository.NewContentRepository)
	do.Provide(i, repository.NewUserRepository)
	do.Provide(i, repository.NewRoleRepository)
	do.Provide(i, repository.NewMemberRepository)
	do.Provide(i, repository.NewCommunityRepository)
	do.Provide(i, repository.NewActivityRepository)
	do.Provide(i, repository.NewResourceSearchIndexRepository)
	do.Provide(i, repository.NewInviteRepository)
	do.Provide(i, repository.NewTopicRepository)
	do.Provide(i, repository.NewThreadRepository)
	do.Provide(i, repository.NewPostRepository)

	do.Provide(i, dservice.NewNoteService)
	do.Provide(i, dservice.NewContentService)
	do.Provide(i, dservice.NewUserService)
	do.Provide(i, dservice.NewRoleService)
	do.Provide(i, dservice.NewMemberService)
	do.Provide(i, dservice.NewCommunityService)
	do.Provide(i, dservice.NewActivityService)
	do.Provide(i, dservice.NewResourceSearchIndexService)
	do.Provide(i, dservice.NewInviteService)
	do.Provide(i, dservice.NewTopicService)
	do.Provide(i, dservice.NewThreadService)
	do.Provide(i, dservice.NewPostService)

	do.Provide(i, uservice.NewNoteUsecase)
	do.Provide(i, uservice.NewUserUsecase)
	do.Provide(i, uservice.NewCommunityUsecase)
	do.Provide(i, uservice.NewRoleUsecase)
	do.Provide(i, uservice.NewActivityUsecase)
	do.Provide(i, uservice.NewAuthUsecase)

	activityUsecase := do.MustInvoke[uservice.ActivityUsecase](i)
	authUsecase := do.MustInvoke[uservice.AuthUsecase](i)
	userUsecase := do.MustInvoke[uservice.UserUsecase](i)

	return Handler{
		activityUsecase: activityUsecase,
		authUsecase:     authUsecase,
		userUsecase:     userUsecase,
	}
}
