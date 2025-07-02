package model

type Post struct {
	ID string `gorm:"primaryKey"`
	At int
}

type PostTopicRelation struct {
	PostID  string `gorm:"primaryKey"`
	TopicID string `gorm:"primaryKey"`
}

type PostThreadRelation struct {
	PostID   string `gorm:"primaryKey"`
	ThreadID string `gorm:"primaryKey"`
}

type PostFromMemberRelation struct {
	PostID   string `gorm:"primaryKey"`
	MemberID string `gorm:"primaryKey"`
}

type PostToMemberRelation struct {
	PostID   string `gorm:"primaryKey"`
	MemberID string `gorm:"primaryKey"`
}

type PostToRoleRelation struct {
	PostID string `gorm:"primaryKey"`
	RoleID string `gorm:"primaryKey"`
}
