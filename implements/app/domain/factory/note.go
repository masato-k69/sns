package factory

import (
	"app/domain/model"

	"github.com/google/uuid"
)

func NewNote(id string) (*model.Note, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	return &model.Note{
		ID: parsedID,
	}, nil
}

func NewLine(id string, noteID string, order int, propertyType *string) (*model.Line, error) {
	parsedID, err := uuid.Parse(id)

	if err != nil {
		return nil, err
	}

	parsedNoteID, err := uuid.Parse(noteID)

	if err != nil {
		return nil, err
	}

	parsedOrder, err := model.NewOrderNumber(order)

	if err != nil {
		return nil, err
	}

	var property *model.LineProperty
	if propertyType == nil {
		property = nil
	} else {
		parsedPropertyType, err := model.NewLinePropertyType(*propertyType)

		if err != nil {
			return nil, err
		}

		property = &model.LineProperty{
			Type: *parsedPropertyType,
		}
	}

	return &model.Line{
		ID:       parsedID,
		NoteID:   parsedNoteID,
		Order:    *parsedOrder,
		Property: property,
	}, nil
}
