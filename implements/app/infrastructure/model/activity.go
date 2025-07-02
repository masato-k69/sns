package model

import (
	"app/infrastructure/adapter/datastore/timeseries"
	"encoding/json"
	"time"
)

type UserLoginActivity struct {
	At               time.Time `json:"at"`
	UserID           string    `json:"user_id"`
	IPAddress        string    `json:"ip_address"`
	OperationgSystem string    `json:"operationg_system"`
	UserAgent        string    `json:"user_agent"`
}

// Timestamp implements timeseries.Point.
func (u UserLoginActivity) Timestamp() time.Time {
	return u.At
}

// Measurement implements timeseries.Point.
func (u UserLoginActivity) Measurement() timeseries.Measurement {
	return "user_login_activities"
}

// Tags implements timeseries.Point.
func (u UserLoginActivity) Tags() map[string]string {
	var result map[string]string

	indirect, _ := json.Marshal(struct {
		At               time.Time `json:"at"`
		UserID           string    `json:"user_id"`
		IPAddress        string    `json:"ip_address"`
		OperationgSystem string    `json:"operationg_system"`
		UserAgent        string    `json:"user_agent"`
	}{
		At:               u.At,
		UserID:           u.UserID,
		IPAddress:        u.IPAddress,
		OperationgSystem: u.OperationgSystem,
		UserAgent:        u.UserAgent,
	})
	json.Unmarshal(indirect, &result)

	return result
}

// Fields implements timeseries.Point.
func (u UserLoginActivity) Fields() map[string]interface{} {
	var result map[string]interface{}

	indirect, _ := json.Marshal(u)
	json.Unmarshal(indirect, &result)

	return result
}

func NewUserLoginActivity(at time.Time, userID string, ipAddress string, operatingSystem string, userAgent string) timeseries.Point {
	return UserLoginActivity{
		At:               at,
		UserID:           userID,
		IPAddress:        ipAddress,
		OperationgSystem: operatingSystem,
		UserAgent:        userAgent,
	}
}

type MemberActivity struct {
	At        time.Time `json:"at"`
	Member    string    `json:"member"`
	Target    string    `json:"target"`
	Resource  string    `json:"resource"`
	Operation string    `json:"operation"`
}

// Timestamp implements timeseries.Point.
func (m MemberActivity) Timestamp() time.Time {
	return m.At
}

// Measurement implements timeseries.Point.
func (m MemberActivity) Measurement() timeseries.Measurement {
	return "member_activities"
}

// Tags implements timeseries.Point.
func (m MemberActivity) Tags() map[string]string {
	var result map[string]string

	indirect, _ := json.Marshal(struct {
		At        time.Time `json:"at"`
		Member    string    `json:"member"`
		Target    string    `json:"target"`
		Resource  string    `json:"resource"`
		Operation string    `json:"operation"`
	}{
		At:        m.At,
		Member:    m.Member,
		Target:    m.Target,
		Resource:  m.Resource,
		Operation: m.Operation,
	})
	json.Unmarshal(indirect, &result)

	return result
}

// Fields implements timeseries.Point.
func (m MemberActivity) Fields() map[string]interface{} {
	var result map[string]interface{}

	indirect, _ := json.Marshal(m)
	json.Unmarshal(indirect, &result)

	return result
}

func NewMemberActivity(at time.Time, member string, target string, resource string, operation string) timeseries.Point {
	return MemberActivity{
		At:        at,
		Member:    member,
		Target:    target,
		Resource:  resource,
		Operation: operation,
	}
}

type MemberLikeActivity struct {
	At       time.Time          `json:"at"`
	Member   string             `json:"member"`
	Target   string             `json:"target"`
	Resource string             `json:"resource"`
	Like     timeseries.Boolean `json:"like"`
	Comment  string             `json:"comment"`
}

// Timestamp implements timeseries.Point.
func (m MemberLikeActivity) Timestamp() time.Time {
	return m.At
}

// Measurement implements timeseries.Point.
func (m MemberLikeActivity) Measurement() timeseries.Measurement {
	return "member_like_activities"
}

// Tags implements timeseries.Point.
func (m MemberLikeActivity) Tags() map[string]string {
	var result map[string]string

	indirect, _ := json.Marshal(struct {
		At       time.Time `json:"at"`
		Member   string    `json:"member"`
		Target   string    `json:"target"`
		Resource string    `json:"resource"`
		Like     string    `json:"like"`
		Comment  string    `json:"comment"`
	}{
		At:       m.At,
		Member:   m.Member,
		Target:   m.Target,
		Resource: m.Resource,
		Like:     m.Like.Tag(),
		Comment:  m.Comment,
	})
	json.Unmarshal(indirect, &result)

	return result
}

// Fields implements timeseries.Point.
func (m MemberLikeActivity) Fields() map[string]interface{} {
	var result map[string]interface{}

	indirect, _ := json.Marshal(m)
	json.Unmarshal(indirect, &result)

	return result
}

func NewMemberLikeActivity(at time.Time, member string, target string, resource string, like bool, comment *string) timeseries.Point {
	parsedComment := ""
	if comment != nil {
		parsedComment = *comment
	}

	return MemberLikeActivity{
		At:       at,
		Member:   member,
		Target:   target,
		Resource: resource,
		Like:     timeseries.NewBoolean(like),
		Comment:  parsedComment,
	}
}
