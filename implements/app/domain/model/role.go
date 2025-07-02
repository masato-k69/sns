package model

import (
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type Role struct {
	ID     uuid.UUID
	Name   Name
	Action Action
}

func (m *Role) CanCreate(resource Resource) bool {
	return m.can(resource, OperationCreate)
}

// func (m *Role) CanRead(resource Resource) bool {
// 	return m.can(resource, OperationRead)
// }

func (m *Role) CanUpdate(resource Resource) bool {
	return m.can(resource, OperationUpdate)
}

func (m *Role) CanDelete(resource Resource) bool {
	return m.can(resource, OperationDelete)
}

func (m *Role) can(resource Resource, operation Operation) bool {
	for r, o := range m.Action {
		if r == resource {
			if slices.Contains(o, operation) {
				return true
			}
		}
	}

	return false
}

type Operation string

func (m Operation) String() string {
	return string(m)
}

func NewOperation(v string) (*Operation, error) {
	t := Operation(v)
	for _, operation := range Operations {
		if t == operation {
			return &t, nil
		}
	}

	return nil, fmt.Errorf("invalid argument. v=%v", v)
}

const (
	OperationCreate Operation = "create"
	// OperationRead   Operation = "read"
	OperationUpdate Operation = "update"
	OperationDelete Operation = "delete"
)

var (
	Operations = []Operation{
		OperationCreate,
		OperationUpdate,
		OperationDelete,
	}
)

type Action map[Resource][]Operation

func (m *Action) Strings() map[string][]string {
	t := map[string][]string{}

	for k, v := range *m {
		t[k.String()] = lo.Map(v, func(e Operation, _ int) string { return e.String() })
	}

	return t
}

func NewAction(v map[string][]string) (*Action, error) {
	t := Action{}

	for key, value := range v {
		parsedResouce, err := NewResource(key)

		if err != nil {
			return nil, err
		}

		parsedOperations := []Operation{}

		for _, operation := range value {
			parsedOperation, err := NewOperation(operation)

			if err != nil {
				return nil, err
			}

			parsedOperations = append(parsedOperations, *parsedOperation)
		}

		t[*parsedResouce] = parsedOperations
	}

	return &t, nil
}

var (
	CommunityMemberAction = Action{
		ResourceCommunity: []Operation{OperationUpdate},
		ResourceMember:    []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceRole:      []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceTopic:     []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceThread:    []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourcePost:      []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceTag:       []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceElection:  []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceChoose:    []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceProject:   []Operation{OperationCreate, OperationUpdate, OperationDelete},
	}

	ProjectMemberAction = Action{
		ResourceProject:   []Operation{OperationUpdate},
		ResourceMember:    []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceRole:      []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceMilestone: []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceTask:      []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourceThread:    []Operation{OperationCreate, OperationUpdate, OperationDelete},
		ResourcePost:      []Operation{OperationCreate, OperationUpdate, OperationDelete},
	}
)
