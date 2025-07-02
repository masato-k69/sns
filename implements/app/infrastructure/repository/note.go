package repository

import (
	dfactory "app/domain/factory"
	dmodel "app/domain/model"
	drepository "app/domain/repository"
	idocument "app/infrastructure/adapter/datastore/document"
	irdb "app/infrastructure/adapter/datastore/rdb"
	imodel "app/infrastructure/model"
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/samber/do"
	"github.com/samber/lo"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type noteRepository struct {
	noteStoreConnectionRDB      irdb.NoteStoreConnection
	noteStoreConnectionDocument idocument.NoteStoreConnection
}

// GetbyResource implements repository.NoteRepository.
func (n *noteRepository) GetbyResource(c context.Context, mention dmodel.Mention) (*dmodel.Note, error) {
	notes := []imodel.Note{}
	switch mention.Resource {
	case dmodel.ResourceUser:
		if err := n.noteStoreConnectionRDB.Read().
			Model(&imodel.Note{}).
			Select("notes.id as id").
			Joins("inner join note_user_relations on notes.id = note_user_relations.note_id").
			Where("note_user_relations.user_id = ?", mention.ID.String()).
			Order("notes.created_at asc").
			Scan(&notes).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to list note. user_id=%v", mention.ID.String())
		}
	case dmodel.ResourceCommunity:
		if err := n.noteStoreConnectionRDB.Read().
			Model(&imodel.Note{}).
			Select("notes.id as id").
			Joins("inner join note_community_relations on notes.id = note_community_relations.note_id").
			Where("note_community_relations.community_id = ?", mention.ID.String()).
			Order("notes.created_at asc").
			Scan(&notes).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to list note. user_id=%v", mention.ID.String())
		}
	}

	if len(notes) < 1 {
		return nil, nil
	}

	dNote, err := dfactory.NewNote(notes[0].ID)

	if err != nil {
		return nil, errors.Wrapf(err, "fialed to parse note. id=%v", notes[0].ID)
	}

	return dNote, nil
}

// UpdateLine implements repository.NoteRepository.
func (n *noteRepository) UpdateLine(c context.Context, line dmodel.Line) error {
	return n.noteStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("note_id = ?", line.NoteID.String()).
			Order("order_number asc").
			Find(&[]imodel.Line{}).Error; err != nil {
			return errors.Wrapf(err, "fialed to lock lines. note_id=%v", line.NoteID.String())
		}

		if err := tx.
			Where("line_id = ?", line.ID.String()).
			Delete(&imodel.ContentLineRelation{}).Error; err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return errors.Wrapf(err, "fialed to delete line content relations. line_id=%v", line.ID.String())
			}
		}

		option, err := idocument.Unmarshal(&imodel.LineProperty{
			LineID: line.ID.String(),
		})
		if err != nil {
			return err
		}

		if _, err := n.noteStoreConnectionDocument.DB().
			Collection(imodel.LineProperty{}.Collection()).
			DeleteOne(c, option); err != nil {
			if err != mongo.ErrNoDocuments {
				return errors.Wrapf(err, "fialed to delete line property. line_id=%v", line.ID.String())
			}
		}

		if line.Property == nil {
			return nil
		}

		if _, err := n.noteStoreConnectionDocument.DB().
			Collection(imodel.LineProperty{}.Collection()).
			InsertOne(c, &imodel.LineProperty{
				LineID: line.ID.String(),
				Type:   line.Property.Type.String(),
			}); err != nil {
			return errors.Wrapf(err, "failed to create line property. line_id=%v", line.ID.String())
		}

		return nil
	})
}

// GetLineByOrder implements repository.NoteRepository.
func (n *noteRepository) GetLineByOrder(c context.Context, noteID uuid.UUID, order dmodel.OrderNumber) (*dmodel.Line, error) {
	line := imodel.Line{}

	if err := n.noteStoreConnectionRDB.Read().
		Where("note_id = ? and order_number = ?", noteID.String(), order.Int()).
		First(&line).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "fialed to get line. note_id=%v order=%v", noteID.String(), order.Int())
	}

	propertyType, err := n.getLineProperty(c, line.ID)

	if err != nil {
		return nil, err
	}

	dLine, err := dfactory.NewLine(line.ID, line.NoteID, line.Order, propertyType)

	if err != nil {
		return nil, errors.Wrapf(err, "fialed to parse line. id=%v", line.ID)
	}

	return dLine, nil
}

// DeleteLine implements repository.NoteRepository.
func (n *noteRepository) DeleteLine(c context.Context, noteID uuid.UUID, order dmodel.OrderNumber) (*dmodel.Line, error) {
	tx := n.noteStoreConnectionRDB.Write().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	deleteLine := imodel.Line{}
	var deletePropertyType *string

	if err := func() error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("note_id = ?", noteID.String()).
			Order("order_number asc").
			Find(&[]imodel.Line{}).Error; err != nil {
			return errors.Wrapf(err, "fialed to lock lines. note_id=%v", noteID.String())
		}

		if err := tx.
			Where("note_id = ? and order_number = ?", noteID.String(), order.Int()).
			Order("order_number desc").
			First(&deleteLine).Error; err != nil {
			return errors.Wrapf(err, "fialed to get lines. note_id=%v order=%v", noteID.String(), order.Int())
		}

		lineProperty := imodel.LineProperty{
			LineID: deleteLine.ID,
		}

		option, err := idocument.Unmarshal(&lineProperty)
		if err != nil {
			return err
		}

		if err := n.noteStoreConnectionDocument.DB().
			Collection(imodel.LineProperty{}.Collection()).
			FindOne(c, option).
			Decode(&lineProperty); err != nil {
			if err == mongo.ErrNoDocuments {
				deletePropertyType = nil
			} else {
				return err
			}
		} else {
			deletePropertyType = &lineProperty.Type
		}

		if err := tx.
			Where("note_id = ? and order_number = ?", noteID.String(), order.Int()).
			Delete(&imodel.Line{}).Error; err != nil {
			return errors.Wrapf(err, "fialed to delete line. note_id=%v order=%v", noteID.String(), order.Int())
		}

		if err := tx.
			Model(&imodel.Line{}).
			Where("note_id = ? and order_number > ?", noteID.String(), order.Int()).
			Order("order_number asc").
			Update("order_number", gorm.Expr("order_number - 1")).Error; err != nil {
			return errors.Wrapf(err, "fialed to update line. note_id=%v order=%v", noteID.String(), order.Int())
		}

		if _, err := n.noteStoreConnectionDocument.DB().
			Collection(imodel.LineProperty{}.Collection()).
			DeleteOne(c, option); err != nil {
			return errors.Wrapf(err, "fialed to delete line property. line_id=%v", deleteLine.ID)
		}

		return nil
	}(); err != nil {
		tx.Rollback()
		return nil, err
	}

	deletedLine, err := dfactory.NewLine(deleteLine.ID, deleteLine.NoteID, deleteLine.Order, deletePropertyType)
	if err != nil {
		return nil, err
	}

	return deletedLine, tx.Commit().Error
}

// MoveLine implements repository.NoteRepository.
func (n *noteRepository) MoveLine(c context.Context, noteID uuid.UUID, src dmodel.OrderNumber, dst dmodel.OrderNumber) error {
	return n.noteStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("note_id = ?", noteID.String()).
			Order("order_number asc").
			Find(&[]imodel.Line{}).Error; err != nil {
			return errors.Wrapf(err, "fialed to lock lines. note_id=%v", noteID.String())
		}

		var lastLine imodel.Line
		if err := tx.
			Where("note_id = ?", noteID.String()).
			Order("order_number desc").
			First(&lastLine).Error; err != nil {
			return errors.Wrapf(err, "fialed to get lines. note_id=%v", noteID.String())
		}

		if src.Int() > lastLine.Order {
			return nil
		}

		srcOrderToUpdate := lo.Min([]int{lastLine.Order, src.Int()})
		dstOrderToUpdate := lo.Min([]int{lastLine.Order, dst.Int()})
		tmpOrderToUpdate := 0 // order重複回避の為の一時的な移動先

		sql, values := func(noteID string, src int, dst int, tmp int) (string, []interface{}) {
			operation := "update `lines` set order_number = "
			condition := fmt.Sprintf(" where note_id = '%v'", noteID)

			if src < dst { // 現在より後の行に移動する場合、移動元より後で移動先以前の行を前に移動する
				return operation + `
				case
					when order_number = ?                       then ?
					when order_number > ? and order_number <= ? then order_number - 1
					else order_number
				end
				` + condition + " order by order_number asc",
					[]interface{}{
						src, tmp,
						src, dst,
					}
			}

			if src > dst { // 現在より前の行に移動する場合、移動元より前で移動先以降の行を後に移動する
				return operation + `
				case
					when order_number =  ?                      then ?
					when order_number >= ? and order_number < ? then order_number + 1
					else order_number
				end
				` + condition + " order by order_number desc",
					[]interface{}{
						src, tmp,
						dst, src,
					}
			}

			return operation + "case when order_number = ? then ? else order_number end" + condition, []interface{}{src, tmp}
		}(noteID.String(), srcOrderToUpdate, dstOrderToUpdate, tmpOrderToUpdate)

		if err := tx.
			Exec(sql, values...).Error; err != nil {
			return errors.Wrapf(err, "fialed to update other lines. note_id=%v", noteID.String())
		}

		return tx.
			Model(&imodel.Line{}).
			Where("note_id = ? and order_number = ?", noteID.String(), tmpOrderToUpdate).
			Update("order_number", lo.Min([]int{dst.Int(), lastLine.Order})).Error
	})
}

// InsertLine implements repository.NoteRepository.
func (n *noteRepository) InsertLine(c context.Context, dLine dmodel.Line) error {
	return n.noteStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := tx.
			Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("note_id = ?", dLine.NoteID.String()).
			Order("order_number asc").
			Find(&[]imodel.Line{}).Error; err != nil {
			return errors.Wrapf(err, "fialed to lock lines. note_id=%v", dLine.NoteID.String())
		}

		var lastLine imodel.Line
		if err := tx.
			Where("note_id = ?", dLine.NoteID.String()).
			Order("order_number desc").
			First(&lastLine).Error; err != nil {
			return errors.Wrapf(err, "fialed to get lines. note_id=%v", dLine.NoteID.String())
		}

		orderToInsert := lo.Min([]int{lastLine.Order, dLine.Order.Int()})

		if err := tx.
			Model(&imodel.Line{}).
			Where("note_id = ? and order_number >= ?", dLine.NoteID.String(), orderToInsert).
			Order("order_number desc").
			Update("order_number", gorm.Expr("order_number + 1")).Error; err != nil {
			return errors.Wrapf(err, "fialed to update lines order. note_id=%v", dLine.NoteID.String())
		}

		return tx.
			Create(&imodel.Line{
				ID:     dLine.ID.String(),
				NoteID: dLine.NoteID.String(),
				Order:  orderToInsert,
			}).Error
	})
}

// ListLines implements repository.NoteRepository.
func (n *noteRepository) ListLines(c context.Context, noteID uuid.UUID) ([]dmodel.Line, error) {
	lines := []imodel.Line{}

	if err := n.noteStoreConnectionRDB.Read().
		Where("note_id = ?", noteID.String()).
		Order("order_number asc").
		Find(&lines).Error; err != nil {
		return nil, errors.Wrapf(err, "fialed to get lines. note_id=%v", noteID)
	}

	dLines := []dmodel.Line{}

	for _, line := range lines {
		propertyType, err := n.getLineProperty(c, line.ID)

		if err != nil {
			return nil, err
		}

		dLine, err := dfactory.NewLine(line.ID, line.NoteID, line.Order, propertyType)

		if err != nil {
			return nil, errors.Wrapf(err, "fialed to parse line. id=%v", line.ID)
		}

		dLines = append(dLines, *dLine)
	}

	return dLines, nil
}

// Get implements repository.NoteRepository.
func (n *noteRepository) Get(c context.Context, id uuid.UUID) (*dmodel.Note, error) {
	note := imodel.Note{ID: id.String()}

	if err := n.noteStoreConnectionRDB.Read().
		First(&note).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}

		return nil, errors.Wrapf(err, "fialed to get note. id=%v", id)
	}

	dNote, err := dfactory.NewNote(note.ID)

	if err != nil {
		return nil, errors.Wrapf(err, "fialed to parse note. id=%v", id)
	}

	return dNote, nil
}

// Create implements repository.NoteRepository.
func (n *noteRepository) Create(c context.Context, note dmodel.Note, mention dmodel.Mention) error {
	return n.noteStoreConnectionRDB.Write().Transaction(func(tx *gorm.DB) error {
		if err := n.noteStoreConnectionRDB.Write().
			Create(&imodel.Note{
				ID: note.ID.String(),
			}).Error; err != nil {
			return errors.Wrapf(err, "failed to create note. id=%v", note.ID.String())
		}

		switch mention.Resource {
		case dmodel.ResourceUser:
			if err := tx.
				Create(&imodel.NoteUserRelation{
					NoteID: note.ID.String(),
					UserID: mention.ID.String(),
				}).Error; err != nil {
				return errors.Wrapf(err, "failed to create user relation. id=%v", mention.ID.String())
			}
		case dmodel.ResourceCommunity:
			if err := tx.
				Create(&imodel.NoteCommunityRelation{
					NoteID:      note.ID.String(),
					CommunityID: mention.ID.String(),
				}).Error; err != nil {
				return errors.Wrapf(err, "failed to create community relation. id=%v", mention.ID.String())
			}
		}

		return nil
	})
}

func (n *noteRepository) getLineProperty(c context.Context, lineID string) (propertyType *string, err error) {
	lineProperty := imodel.LineProperty{
		LineID: lineID,
	}

	option, err := idocument.Unmarshal(&lineProperty)

	if err != nil {
		return nil, err
	}

	if err := n.noteStoreConnectionDocument.DB().
		Collection(imodel.LineProperty{}.Collection()).
		FindOne(c, option).
		Decode(&lineProperty); err != nil {
		if err == mongo.ErrNoDocuments {
			propertyType = nil
		} else {
			return nil, err
		}
	} else {
		propertyType = &lineProperty.Type
	}

	return propertyType, nil
}

func NewNoteRepository(i *do.Injector) (drepository.NoteRepository, error) {
	noteStoreConnectionRDB := do.MustInvoke[irdb.NoteStoreConnection](i)
	noteStoreConnectionDocument := do.MustInvoke[idocument.NoteStoreConnection](i)
	return &noteRepository{
		noteStoreConnectionRDB:      noteStoreConnectionRDB,
		noteStoreConnectionDocument: noteStoreConnectionDocument,
	}, nil
}
