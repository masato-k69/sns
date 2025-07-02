package model

import "github.com/google/uuid"

type Note struct {
	ID uuid.UUID
}

type Line struct {
	NoteID   uuid.UUID
	Order    int
	Property *LineProperty
	Contents []Content
}

type LineProperty struct {
	Type string
}
