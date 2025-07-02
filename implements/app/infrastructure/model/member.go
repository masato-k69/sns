package model

type Member struct {
	ID     string `gorm:"primaryKey"`
	UserID string
	RoleID string
}

type MemberCommunityRelation struct {
	MemberID    string `gorm:"primaryKey"`
	CommunityID string `gorm:"primaryKey"`
}
