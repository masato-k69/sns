package repository

import (
	"app/domain/model"
	"context"

	"github.com/google/uuid"
)

type NoteRepository interface {
	Create(c context.Context, note model.Note, mention model.Mention) error
	Get(c context.Context, id uuid.UUID) (*model.Note, error)
	GetbyResource(c context.Context, mention model.Mention) (*model.Note, error)
	InsertLine(c context.Context, line model.Line) error
	GetLineByOrder(c context.Context, noteID uuid.UUID, order model.OrderNumber) (*model.Line, error)
	ListLines(c context.Context, noteID uuid.UUID) ([]model.Line, error)
	MoveLine(c context.Context, noteID uuid.UUID, src model.OrderNumber, dst model.OrderNumber) error
	UpdateLine(c context.Context, line model.Line) error
	DeleteLine(c context.Context, noteID uuid.UUID, order model.OrderNumber) (*model.Line, error)
}
