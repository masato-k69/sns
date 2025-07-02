package implement

import (
	"app/gen/pubsub"
	"app/usecase/service"
	"context"
	"encoding/json"
)

type userLoginActivityHandler struct {
	usecase service.ActivityUsecase
}

// Handle implements subscriber.Handler.
func (u userLoginActivityHandler) Handle(c context.Context, message []byte) error {
	var m pubsub.UserLoginActivity
	if err := json.Unmarshal(message, &m); err != nil {
		return err
	}

	return u.usecase.SaveUserLoginActivity(c, m.At.Value.AsTime(), m.UserId.Value, m.IpAddress, m.OperationSystem, m.UserAgent)
}

type memberActivityHandler struct {
	usecase service.ActivityUsecase
}

// Handle implements subscriber.Handler.
func (me memberActivityHandler) Handle(c context.Context, message []byte) error {
	var m pubsub.MemberActivity
	if err := json.Unmarshal(message, &m); err != nil {
		return err
	}

	return me.usecase.SaveMemberActivity(c, m.At.Value.AsTime(), m.Member.Value, m.Target.Value, m.Action.Resource.Value, m.Action.Operation.Value)
}

type memberLikeActivityHandler struct {
	usecase service.ActivityUsecase
}

// Handle implements subscriber.Handler.
func (me memberLikeActivityHandler) Handle(c context.Context, message []byte) error {
	var m pubsub.MemberLikeActivity
	if err := json.Unmarshal(message, &m); err != nil {
		return err
	}

	var comment *string
	if m.Comment != nil {
		comment = &m.Comment.Value
	}

	return me.usecase.SaveMemberLikeActivity(c, m.At.Value.AsTime(), m.Member.Value, m.Target.Value, m.Resource.Value, m.Like, comment)
}
