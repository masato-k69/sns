package model

type Note struct {
	ID string `gorm:"primaryKey"`
}

type NoteUserRelation struct {
	NoteID string `gorm:"primaryKey"`
	UserID string `gorm:"primaryKey"`
}

type NoteCommunityRelation struct {
	NoteID      string `gorm:"primaryKey"`
	CommunityID string `gorm:"primaryKey"`
}

type Line struct {
	ID     string `gorm:"primaryKey"`
	NoteID string
	Order  int `gorm:"column:order_number"`
}

type LineProperty struct {
	LineID string `bson:"line_id"`
	Type   string `bson:"type"`
}

func (m LineProperty) Collection() string {
	return "line_properties"
}
