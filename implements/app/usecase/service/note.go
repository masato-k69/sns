package service

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	dservice "app/domain/service"
	uerror "app/usecase/error"
	umodel "app/usecase/model"

	"context"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
)

type NoteUsecase interface {
	Get(c context.Context, id uuid.UUID) (*umodel.Note, error)
	GetUserProfile(c context.Context, userID uuid.UUID) (*umodel.Note, error)
	GetCommunityDescription(c context.Context, communityID uuid.UUID) (*umodel.Note, error)
	InsertLine(c context.Context, noteID uuid.UUID, order int) error
	ListLines(c context.Context, noteID uuid.UUID) ([]umodel.Line, error)
	MoveLine(c context.Context, noteID uuid.UUID, src int, dst int) error
	UpdateLine(c context.Context, line umodel.Line) error
	DeleteLine(c context.Context, noteID uuid.UUID, order int) error
}

type noteUsecase struct {
	noteService    dservice.NoteService
	contentService dservice.ContentService
}

// GetCommunityDescription implements NoteUsecase.
func (n *noteUsecase) GetCommunityDescription(c context.Context, communityID uuid.UUID) (*umodel.Note, error) {
	mention, err := dmodel.NewMention(communityID.String(), string(dmodel.ResourceCommunity.String()))
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse mention", err)
	}

	note, err := n.noteService.GetbyResource(c, *mention)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, uerror.NewNotFound("note not found", nil)
	}

	return &umodel.Note{
		ID: note.ID,
	}, nil
}

// GetUserProfile implements NoteUsecase.
func (n *noteUsecase) GetUserProfile(c context.Context, userID uuid.UUID) (*umodel.Note, error) {
	mention, err := dmodel.NewMention(userID.String(), dmodel.ResourceUser.String())
	if err != nil {
		return nil, uerror.NewInvalidParameter("failed to parse mention", err)
	}

	note, err := n.noteService.GetbyResource(c, *mention)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, uerror.NewNotFound("note not found", nil)
	}

	return &umodel.Note{
		ID: note.ID,
	}, nil
}

// UpdateLine implements NoteUsecase.
func (n *noteUsecase) UpdateLine(c context.Context, line umodel.Line) error {
	parsedOrder, err := dmodel.NewOrderNumber(line.Order)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse order", err)
	}

	currentLine, err := n.noteService.GetLineByOrder(c, line.NoteID, *parsedOrder)
	if err != nil {
		return err
	}
	if currentLine == nil {
		return uerror.NewNotFound("line not found", nil)
	}

	newContents := []dmodel.Content{}
	for _, content := range line.Contents {
		newContentID := uuid.New()
		newContent, err := dfactory.NewContent(newContentID.String(), content.Type, content.Bin)

		if err != nil {
			return uerror.NewInvalidParameter("failed to parse content", err)
		}

		newContents = append(newContents, *newContent)
	}

	var updatePropertyType *string
	if line.Property == nil {
		updatePropertyType = nil
	} else {
		updatePropertyType = &line.Property.Type
	}

	updateLine, err := dfactory.NewLine(currentLine.ID.String(), currentLine.NoteID.String(), currentLine.Order.Int(), updatePropertyType)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse line", err)
	}

	mention, err := dmodel.NewMention(currentLine.ID.String(), dmodel.ResourceLine.String())
	if err != nil {
		return err
	}

	if err := n.contentService.DeleteAndCreate(c, newContents, *mention); err != nil {
		return nil
	}

	return n.noteService.UpdateLine(c, *updateLine)
}

// DeleteLine implements NoteUsecase.
func (n *noteUsecase) DeleteLine(c context.Context, noteID uuid.UUID, order int) error {
	parsedOrder, err := dmodel.NewOrderNumber(order)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse order", err)
	}

	line, err := n.noteService.GetLineByOrder(c, noteID, *parsedOrder)
	if err != nil {
		return err
	}
	if line == nil {
		return uerror.NewNotFound("line not found", nil)
	}

	deletedLine, err := n.noteService.DeleteLine(c, noteID, *parsedOrder)
	if err != nil {
		return err
	}

	mention, err := dmodel.NewMention(deletedLine.ID.String(), dmodel.ResourceLine.String())
	if err != nil {
		return err
	}

	if err := n.contentService.DeleteByResource(c, *mention); err != nil {
		return errors.Wrapf(err, "failed to delete contents. line_id=%v", deletedLine.ID)
	}

	return nil
}

// MoveLine implements NoteUsecase.
func (n *noteUsecase) MoveLine(c context.Context, noteID uuid.UUID, src int, dst int) error {
	if src == dst {
		return nil
	}

	srcOrder, err := dmodel.NewOrderNumber(src)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse src order", err)
	}

	dstOrder, err := dmodel.NewOrderNumber(dst)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse src order", err)
	}

	return n.noteService.MoveLine(c, noteID, *srcOrder, *dstOrder)
}

// InsertLine implements NoteUsecase.
func (n *noteUsecase) InsertLine(c context.Context, noteID uuid.UUID, order int) error {
	dLine, err := dfactory.NewLine(uuid.NewString(), noteID.String(), order, nil)
	if err != nil {
		return uerror.NewInvalidParameter("failed to parse line", err)
	}

	return n.noteService.InsertLine(c, *dLine)
}

// ListLines implements NoteUsecase.
func (n *noteUsecase) ListLines(c context.Context, noteID uuid.UUID) ([]umodel.Line, error) {
	lines, err := n.noteService.ListLines(c, noteID)
	if err != nil {
		return nil, err
	}

	uLines := []umodel.Line{}
	for _, line := range lines {
		uContents := []umodel.Content{}

		contents, err := n.contentService.ListByLine(c, line.ID)
		if err != nil {
			return nil, err
		}

		for _, content := range contents {
			uContents = append(uContents, umodel.Content{
				Type: content.Type.String(),
				Bin:  content.Value,
			})
		}

		var property *umodel.LineProperty
		if line.Property == nil {
			property = nil
		} else {
			property = &umodel.LineProperty{
				Type: line.Property.Type.String(),
			}
		}

		uLines = append(uLines, umodel.Line{
			NoteID:   line.NoteID,
			Order:    line.Order.Int(),
			Property: property,
			Contents: uContents,
		})
	}

	return uLines, nil
}

// Get implements NoteUsecase.
func (n *noteUsecase) Get(c context.Context, id uuid.UUID) (*umodel.Note, error) {
	note, err := n.noteService.Get(c, id)
	if err != nil {
		return nil, err
	}
	if note == nil {
		return nil, uerror.NewNotFound("note not found", nil)
	}

	return &umodel.Note{
		ID: note.ID,
	}, nil
}

func NewNoteUsecase(i *do.Injector) (NoteUsecase, error) {
	noteService := do.MustInvoke[dservice.NoteService](i)
	contentService := do.MustInvoke[dservice.ContentService](i)
	return &noteUsecase{
		noteService:    noteService,
		contentService: contentService,
	}, nil
}
