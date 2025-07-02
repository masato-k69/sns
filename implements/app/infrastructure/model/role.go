package model

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Role struct {
	ID   string `gorm:"primaryKey"`
	Name string
}

type RoleCommunityRelation struct {
	RoleID      string `gorm:"primaryKey"`
	CommunityID string `gorm:"primaryKey"`
	Default     bool
}

type Action struct {
	RoleID string       `bson:"role_id"`
	Items  []ActionItem `bson:"items"`
}

type ActionItem struct {
	Resource  string   `bson:"resource"`
	Operation []string `bson:"operation"`
}

func (m Action) Collection() string {
	return "actions"
}

func (m Action) CollectionKey() primitive.D {
	return bson.D{{Key: "role_id", Value: m.RoleID}}
}
