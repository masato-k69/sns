package repository

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	drepository "app/domain/repository"
	"app/gen/pubsub"
	"app/infrastructure/adapter/datastore/timeseries"
	"app/infrastructure/adapter/mq"
	imodel "app/infrastructure/model"
	llog "app/lib/log"
	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"github.com/samber/lo"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type activityRepository struct {
	activityStoreMQ mq.ActivityStoreConnection
	activityStoreTS timeseries.TimeseriesStore
}

// ListRecentMemberActivity implements repository.ActivityRepository.
func (a *activityRepository) ListRecentMemberActivity(c context.Context, memberID uuid.UUID, page dmodel.Range) ([]dmodel.MemberActivity, error) {
	option := timeseries.NewQueryOption(imodel.MemberActivity{}.Measurement(), []timeseries.QueryCondition{
		{
			Key:   "member",
			Ope:   timeseries.EQ,
			Value: memberID.String(),
		},
	}, nil, &timeseries.RowRange{
		Limit:  page.Limit,
		Offset: page.Offset,
	}, true)

	activities := []imodel.MemberActivity{}
	if err := a.activityStoreTS.Find(c, option, &activities); err != nil {
		return nil, errors.Wrapf(err, "failed to list member activity. id=%v", memberID.String())
	}

	dActivities := []dmodel.MemberActivity{}
	for _, activity := range activities {
		dActivity, err := dfactory.NewMemberActivity(activity.At, activity.Member, activity.Target, activity.Resource, activity.Operation)
		if err != nil {
			return nil, err
		}

		dActivities = append(dActivities, *dActivity)
	}

	return dActivities, nil
}

// ListMembersLikeActivity implements repository.ActivityRepository.
func (a *activityRepository) ListMembersLikeActivity(c context.Context, memberIDs []uuid.UUID, page dmodel.Range) ([]dmodel.MemberLikeActivity, error) {
	option := timeseries.NewQueryOption(imodel.MemberLikeActivity{}.Measurement(), []timeseries.QueryCondition{
		{
			Key:   "member",
			Ope:   timeseries.Contains,
			Value: memberIDs,
		},
	}, nil, &timeseries.RowRange{
		Limit:  page.Limit,
		Offset: page.Offset,
	}, false)

	activities := []imodel.MemberLikeActivity{}
	if err := a.activityStoreTS.Find(c, option, &activities); err != nil {
		return nil, errors.Wrapf(err, "failed to list member like activity. ids=%v", memberIDs)
	}

	dActivities := []dmodel.MemberLikeActivity{}
	for _, activity := range activities {
		var comment *string
		if activity.Comment != "" {
			comment = &activity.Comment
		} else {
			comment = nil
		}

		dActivity, err := dfactory.NewMemberLikeActivity(activity.At, activity.Member, activity.Target, activity.Resource, activity.Like.Bool(), comment)
		if err != nil {
			return nil, err
		}

		dActivities = append(dActivities, *dActivity)
	}

	return dActivities, nil
}

// ListMembersActivity implements repository.ActivityRepository.
func (a *activityRepository) ListMembersActivity(c context.Context, memberIDs []uuid.UUID, page dmodel.Range) ([]dmodel.MemberActivity, error) {
	option := timeseries.NewQueryOption(imodel.MemberActivity{}.Measurement(), []timeseries.QueryCondition{
		{
			Key:   "member",
			Ope:   timeseries.Contains,
			Value: memberIDs,
		},
	}, nil, &timeseries.RowRange{
		Limit:  page.Limit,
		Offset: page.Offset,
	}, false)

	activities := []imodel.MemberActivity{}
	if err := a.activityStoreTS.Find(c, option, &activities); err != nil {
		return nil, errors.Wrapf(err, "failed to list member activity. ids=%v", memberIDs)
	}

	dActivities := []dmodel.MemberActivity{}
	for _, activity := range activities {
		dActivity, err := dfactory.NewMemberActivity(activity.At, activity.Member, activity.Target, activity.Resource, activity.Operation)
		if err != nil {
			return nil, err
		}

		dActivities = append(dActivities, *dActivity)
	}

	return dActivities, nil
}

// ListUserLoginActivity implements repository.ActivityRepository.
func (a *activityRepository) ListUserLoginActivity(c context.Context, userID uuid.UUID, page dmodel.Range) ([]dmodel.UserLoginActivity, error) {
	option := timeseries.NewQueryOption(imodel.UserLoginActivity{}.Measurement(), []timeseries.QueryCondition{
		{
			Key:   "user_id",
			Ope:   timeseries.EQ,
			Value: userID.String(),
		},
	}, nil, &timeseries.RowRange{
		Limit:  page.Limit,
		Offset: page.Offset,
	}, false)

	activities := []imodel.UserLoginActivity{}
	if err := a.activityStoreTS.Find(c, option, &activities); err != nil {
		return nil, errors.Wrapf(err, "failed to list user login activity. id=%v", userID.String())
	}

	dActivities := []dmodel.UserLoginActivity{}
	for _, activity := range activities {
		dActivity, err := dfactory.NewUserLoginActivity(activity.At, activity.UserID, activity.IPAddress, activity.OperationgSystem, activity.UserAgent)
		if err != nil {
			return nil, err
		}

		dActivities = append(dActivities, *dActivity)
	}

	return dActivities, nil
}

// ListMemberLikeActivity implements repository.ActivityRepository.
func (a *activityRepository) ListMemberLikeActivity(c context.Context, target dmodel.Mention, like bool, page dmodel.Range) ([]dmodel.MemberLikeActivity, error) {
	option := timeseries.NewQueryOption(imodel.MemberLikeActivity{}.Measurement(), []timeseries.QueryCondition{
		{
			Key:   "resource",
			Ope:   timeseries.EQ,
			Value: target.Resource.String(),
		},
		{
			Key:   "target",
			Ope:   timeseries.EQ,
			Value: target.ID.String(),
		},
		{
			Key:   "like",
			Ope:   timeseries.EQ,
			Value: timeseries.NewBoolean(like),
		},
	}, nil, &timeseries.RowRange{
		Limit:  page.Limit,
		Offset: page.Offset,
	}, false)

	activities := []imodel.MemberLikeActivity{}
	if err := a.activityStoreTS.Find(c, option, &activities); err != nil {
		return nil, errors.Wrapf(err, "failed to list member like activity. resource=%v target=%v like=%v", target.Resource.String(), target.ID.String(), like)
	}

	dActivities := []dmodel.MemberLikeActivity{}
	for _, activity := range activities {
		var comment *string
		if activity.Comment != "" {
			comment = &activity.Comment
		} else {
			comment = nil
		}

		dActivity, err := dfactory.NewMemberLikeActivity(activity.At, activity.Member, activity.Target, activity.Resource, activity.Like.Bool(), comment)
		if err != nil {
			return nil, err
		}

		dActivities = append(dActivities, *dActivity)
	}

	return dActivities, nil
}

// CountMemberLikeActivity implements repository.ActivityRepository.
func (a *activityRepository) CountMemberLikeActivity(c context.Context, mention dmodel.Mention, like bool) (*int, error) {
	option := timeseries.NewQueryOption(imodel.MemberLikeActivity{}.Measurement(), []timeseries.QueryCondition{
		{
			Key:   "resource",
			Ope:   timeseries.EQ,
			Value: mention.Resource.String(),
		},
		{
			Key:   "target",
			Ope:   timeseries.EQ,
			Value: mention.ID.String(),
		},
		{
			Key:   "like",
			Ope:   timeseries.EQ,
			Value: timeseries.NewBoolean(like),
		},
	}, nil, nil, false)

	result, err := a.activityStoreTS.Count(c, option)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to count member like activity. resource=%v target=%v like=%v", mention.Resource.String(), mention.ID.String(), like)
	}

	return &result, nil
}

// SaveMemberLikeActivity implements repository.ActivityRepository.
func (a *activityRepository) SaveMemberLikeActivity(c context.Context, activity dmodel.MemberLikeActivity) error {
	var comment *pubsub.Text
	if activity.Comment != nil {
		comment = &pubsub.Text{Value: activity.Comment.String()}
	}

	return a.activityStoreMQ.Publish(c, mq.ExchangeActivity, mq.RoutingKeyActivityMemberLike,
		pubsub.MemberLikeActivity{
			At:       &pubsub.At{Value: timestamppb.New(activity.At)},
			Member:   &pubsub.UUID{Value: activity.Member.String()},
			Target:   &pubsub.UUID{Value: activity.Target.String()},
			Resource: &pubsub.Resource{Value: activity.Resource.String()},
			Like:     activity.Like,
			Comment:  comment,
		})
}

// SaveMemberActivity implements repository.ActivityRepository.
func (a *activityRepository) SaveMemberActivity(c context.Context, activity dmodel.MemberActivity) error {
	return a.activityStoreMQ.Publish(c, mq.ExchangeActivity, mq.RoutingKeyActivityMember,
		pubsub.MemberActivity{
			At:     &pubsub.At{Value: timestamppb.New(activity.At)},
			Member: &pubsub.UUID{Value: activity.Member.String()},
			Target: &pubsub.UUID{Value: activity.Target.String()},
			Action: &pubsub.Action{
				Resource:  &pubsub.Resource{Value: activity.Resource.String()},
				Operation: &pubsub.Operation{Value: activity.Operation.String()},
			},
		})
}

// SaveUserLoginActivity implements repository.ActivityRepository.
func (a *activityRepository) SaveUserLoginActivity(c context.Context, activity dmodel.UserLoginActivity) error {
	return a.activityStoreMQ.Publish(c, mq.ExchangeActivity, mq.RoutingKeyActivityUser,
		pubsub.UserLoginActivity{
			At:              &pubsub.At{Value: timestamppb.New(activity.At)},
			UserId:          &pubsub.UUID{Value: activity.UserID.String()},
			IpAddress:       activity.IPAddress.String(),
			OperationSystem: activity.OperationSystem.String(),
			UserAgent:       activity.UserAgent.String(),
		})
}

func NewActivityRepository(i *do.Injector) (drepository.ActivityRepository, error) {
	activityStoreMQ := do.MustInvoke[mq.ActivityStoreConnection](i)
	activityStoreTS := do.MustInvoke[timeseries.TimeseriesStore](i)
	return &activityRepository{
		activityStoreMQ: activityStoreMQ,
		activityStoreTS: activityStoreTS,
	}, nil
}

type activityRepositoryForAsync struct {
	activityStore timeseries.TimeseriesStore
}

// ListRecentMemberActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) ListRecentMemberActivity(c context.Context, memberID uuid.UUID, page dmodel.Range) ([]dmodel.MemberActivity, error) {
	panic("unimplemented")
}

// ListMembersLikeActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) ListMembersLikeActivity(c context.Context, memberIDs []uuid.UUID, page dmodel.Range) ([]dmodel.MemberLikeActivity, error) {
	panic("unimplemented")
}

// ListMembersActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) ListMembersActivity(c context.Context, memberIDs []uuid.UUID, page dmodel.Range) ([]dmodel.MemberActivity, error) {
	panic("unimplemented")
}

// ListUserLoginActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) ListUserLoginActivity(c context.Context, userID uuid.UUID, page dmodel.Range) ([]dmodel.UserLoginActivity, error) {
	panic("unimplemented")
}

// ListMemberLikeActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) ListMemberLikeActivity(c context.Context, mention dmodel.Mention, like bool, page dmodel.Range) ([]dmodel.MemberLikeActivity, error) {
	panic("unimplemented")
}

// CountMemberLikeActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) CountMemberLikeActivity(c context.Context, mention dmodel.Mention, like bool) (*int, error) {
	panic("unimplemented")
}

// SaveMemberLikeActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) SaveMemberLikeActivity(c context.Context, activity dmodel.MemberLikeActivity) error {

	activities := []imodel.MemberLikeActivity{}
	option := timeseries.NewQueryOption(imodel.MemberLikeActivity{}.Measurement(), []timeseries.QueryCondition{
		{
			Key:   "member",
			Ope:   timeseries.EQ,
			Value: activity.Member.String(),
		},
		{
			Key:   "target",
			Ope:   timeseries.EQ,
			Value: activity.Target.String(),
		},
	}, nil, nil, false)

	if err := a.activityStore.Find(c, option, &activities); err != nil {
		return errors.Wrapf(err, "failed to find member like activity. member=%v target=%v resource=%v like=%v", activity.Member.String(), activity.Target.String(), activity.Resource.String(), activity.Like)
	}

	currentActivity, exists := lo.First(activities)
	if exists {
		llog.Debug(c, "like already exists. member=%v target=%v", currentActivity.Member, currentActivity.Target)
		return nil
	}

	var comment *string
	if activity.Comment != nil {
		v := activity.Comment.String()
		comment = &v
	}

	return a.activityStore.Save(c, imodel.NewMemberLikeActivity(activity.At,
		activity.Member.String(),
		activity.Target.String(),
		activity.Resource.String(),
		activity.Like,
		comment,
	))
}

// SaveMemberActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) SaveMemberActivity(c context.Context, activity dmodel.MemberActivity) error {
	return a.activityStore.Save(c, imodel.NewMemberActivity(activity.At,
		activity.Member.String(),
		activity.Target.String(),
		activity.Resource.String(),
		activity.Operation.String(),
	))
}

// SaveUserLoginActivity implements repository.ActivityRepository.
func (a *activityRepositoryForAsync) SaveUserLoginActivity(c context.Context, activity dmodel.UserLoginActivity) error {
	return a.activityStore.Save(c, imodel.NewUserLoginActivity(activity.At,
		activity.UserID.String(),
		activity.IPAddress.String(),
		activity.OperationSystem.String(),
		activity.UserAgent.String(),
	))
}

func NewActivityRepositoryForAsync(i *do.Injector) (drepository.ActivityRepository, error) {
	activityStore := do.MustInvoke[timeseries.TimeseriesStore](i)
	return &activityRepositoryForAsync{activityStore: activityStore}, nil
}
