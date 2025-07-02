package model

type Thread struct {
	ID string `gorm:"primaryKey"`
}

type ThreadTopicRelation struct {
	ThreadID string `gorm:"primaryKey"`
	TopicID  string `gorm:"primaryKey"`
}
