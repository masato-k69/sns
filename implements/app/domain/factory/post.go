package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewPost(id string, from *string, to []model.Mention, at int) (*model.Post, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	parsedAt, err := model.NewUnixTime(at)

	if err != nil {
		return nil, err
	}

	var dFrom *uuid.UUID
	if from != nil {
		parsedFrom, err := uuid.Parse(*from)
		if err != nil {
			return nil, err
		}

		dFrom = &parsedFrom
	}

	return &model.Post{
		ID:   parsedID,
		At:   *parsedAt,
		From: dFrom,
		To:   to,
	}, nil
}
