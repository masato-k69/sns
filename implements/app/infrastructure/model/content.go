package model

type Content struct {
	ID   string `gorm:"primaryKey"`
	Type string
	Bin  []byte
}

type ContentLineRelation struct {
	ContentID string `gorm:"primaryKey"`
	LineID    string `gorm:"primaryKey"`
}

type ContentPostRelation struct {
	ContentID string `gorm:"primaryKey"`
	PostID    string `gorm:"primaryKey"`
}

type ContentTopicRelation struct {
	ContentID string `gorm:"primaryKey"`
	TopicID   string `gorm:"primaryKey"`
}
