package model

import (
	"fmt"

	"github.com/google/uuid"
)

type Resource string

func (m Resource) String() string {
	return string(m)
}

func NewResource(v string) (*Resource, error) {
	t := Resource(v)
	for _, resource := range Resources {
		if t == resource {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("invalid argument. v=%v", v)
}

const (
	ResourceUser      Resource = "user"
	ResourceCommunity Resource = "community"
	ResourceMember    Resource = "member"
	ResourceRole      Resource = "role"
	ResourceTopic     Resource = "topic"
	ResourceThread    Resource = "thread"
	ResourcePost      Resource = "post"
	ResourceProject   Resource = "project"
	ResourceMilestone Resource = "milestone"
	ResourceTask      Resource = "task"
	ResourceTag       Resource = "tag"
	ResourceElection  Resource = "election"
	ResourceChoose    Resource = "choose"

	// internal
	ResourceLine    Resource = "line"
	ResourceLike    Resource = "like"
	ResourceDislike Resource = "dislike"
)

var (
	Resources = []Resource{
		ResourceUser,
		ResourceCommunity,
		ResourceMember,
		ResourceRole,
		ResourceTopic,
		ResourceThread,
		ResourcePost,
		ResourceProject,
		ResourceMilestone,
		ResourceTask,
		ResourceTag,
		ResourceElection,
		ResourceChoose,
	}
)

type ResourceSearchIndex struct {
	ResourceID uuid.UUID
	Type       Resource
	Keyword    Text
}
