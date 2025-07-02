package factory

import (
	"app/domain/model"
	"time"

	"github.com/google/uuid"
)

func NewUserLoginActivity(at time.Time, userID string, ipAddress string, operatinSystem string, userAgent string) (*model.UserLoginActivity, error) {
	parsedUserID, err := uuid.Parse(userID)

	if err != nil {
		return nil, err
	}

	parsedIPAddress, err := model.NewIPAddress(ipAddress)

	if err != nil {
		return nil, err
	}

	parsedOperationSystem, err := model.NewOperationgSystem(operatinSystem)

	if err != nil {
		return nil, err
	}

	parsedUserAgent, err := model.NewUserAgent(userAgent)

	if err != nil {
		return nil, err
	}

	return &model.UserLoginActivity{
		At:              at,
		UserID:          parsedUserID,
		IPAddress:       *parsedIPAddress,
		OperationSystem: *parsedOperationSystem,
		UserAgent:       *parsedUserAgent,
	}, nil
}

func NewMemberActivity(at time.Time, member string, target string, resource string, operation string) (*model.MemberActivity, error) {
	parsedMember, err := uuid.Parse(member)

	if err != nil {
		return nil, err
	}

	parsedTarget, err := uuid.Parse(target)

	if err != nil {
		return nil, err
	}

	parsedResource, err := model.NewResource(resource)

	if err != nil {
		return nil, err
	}

	parsedOperation, err := model.NewOperation(operation)

	if err != nil {
		return nil, err
	}

	return &model.MemberActivity{
		At:        at,
		Member:    parsedMember,
		Target:    parsedTarget,
		Resource:  *parsedResource,
		Operation: *parsedOperation,
	}, nil
}

func NewMemberLikeActivity(at time.Time, member string, target string, resource string, like bool, comment *string) (*model.MemberLikeActivity, error) {
	parsedMember, err := uuid.Parse(member)

	if err != nil {
		return nil, err
	}

	parsedTarget, err := uuid.Parse(target)

	if err != nil {
		return nil, err
	}

	parsedResource, err := model.NewResource(resource)

	if err != nil {
		return nil, err
	}

	var parsedComment *model.ShortMessage
	if comment != nil {
		m, err := model.NewShortMessage(*comment)
		if err != nil {
			return nil, err
		}

		parsedComment = m
	}

	return &model.MemberLikeActivity{
		At:       at,
		Member:   parsedMember,
		Target:   parsedTarget,
		Resource: *parsedResource,
		Like:     like,
		Comment:  parsedComment,
	}, nil
}
