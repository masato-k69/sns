package service

import (
	"app/domain/model"
	"app/domain/repository"
	"context"

	"github.com/google/uuid"
	"github.com/samber/do"
)

type NoteService interface {
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

type noteService struct {
	noteRepository repository.NoteRepository
}

// GetbyResource implements NoteService.
func (n *noteService) GetbyResource(c context.Context, mention model.Mention) (*model.Note, error) {
	return n.noteRepository.GetbyResource(c, mention)
}

// UpdateLine implements NoteService.
func (n *noteService) UpdateLine(c context.Context, line model.Line) error {
	return n.noteRepository.UpdateLine(c, line)
}

// GetLineByOrder implements NoteService.
func (n *noteService) GetLineByOrder(c context.Context, noteID uuid.UUID, order model.OrderNumber) (*model.Line, error) {
	return n.noteRepository.GetLineByOrder(c, noteID, order)
}

// DeleteLine implements NoteService.
func (n *noteService) DeleteLine(c context.Context, noteID uuid.UUID, order model.OrderNumber) (*model.Line, error) {
	return n.noteRepository.DeleteLine(c, noteID, order)
}

// MoveLine implements NoteService.
func (n *noteService) MoveLine(c context.Context, noteID uuid.UUID, src model.OrderNumber, dst model.OrderNumber) error {
	return n.noteRepository.MoveLine(c, noteID, src, dst)
}

// InsertLine implements NoteService.
func (n *noteService) InsertLine(c context.Context, line model.Line) error {
	return n.noteRepository.InsertLine(c, line)
}

// ListLines implements NoteService.
func (n *noteService) ListLines(c context.Context, noteID uuid.UUID) ([]model.Line, error) {
	return n.noteRepository.ListLines(c, noteID)
}

// Get implements NoteService.
func (n *noteService) Get(c context.Context, id uuid.UUID) (*model.Note, error) {
	return n.noteRepository.Get(c, id)
}

// Create implements NoteService.
func (n *noteService) Create(c context.Context, note model.Note, mention model.Mention) error {
	return n.noteRepository.Create(c, note, mention)
}

func NewNoteService(i *do.Injector) (NoteService, error) {
	noteRepository := do.MustInvoke[repository.NoteRepository](i)
	return &noteService{noteRepository: noteRepository}, nil
}
