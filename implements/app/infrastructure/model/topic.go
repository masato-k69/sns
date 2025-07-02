package model

type Topic struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

type TopicCommunityRelation struct {
	TopicID     string `gorm:"primaryKey"`
	CommunityID string `gorm:"primaryKey"`
}

type TopicFromMemberRelation struct {
	TopicID  string `gorm:"primaryKey"`
	MemberID string `gorm:"primaryKey"`
}
