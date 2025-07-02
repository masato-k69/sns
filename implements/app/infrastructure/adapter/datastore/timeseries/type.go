package timeseries

import (
	"strconv"
	"time"
)

type Measurement string

type Point interface {
	Measurement() Measurement
	Tags() map[string]string
	Fields() map[string]interface{}
	Timestamp() time.Time
}

type ComparisonOperator string

const (
	EQ       ComparisonOperator = "=="
	NE       ComparisonOperator = "!="
	GT       ComparisonOperator = ">"
	LT       ComparisonOperator = "<"
	GE       ComparisonOperator = ">="
	LE       ComparisonOperator = "<="
	Contains ComparisonOperator = "contains"
)

type QueryCondition struct {
	Key   string
	Ope   ComparisonOperator
	Value any
}

type TimeRange struct {
	Start UnixTime
	Stop  UnixTime
}

type RowRange struct {
	Limit  int
	Offset int
}

type queryOption struct {
	Measurement Measurement
	Conditions  []QueryCondition
	timeRange   *TimeRange
	RowRange    *RowRange
	Desc        bool
}

func (q queryOption) TimeRange() TimeRange {
	if q.timeRange == nil {
		return TimeRange{
			Start: NewUnixTime(time.Unix(0, 0)),
			Stop:  NewUnixTime(time.Now()),
		}
	}

	return TimeRange{
		Start: q.timeRange.Start,
		Stop:  q.timeRange.Stop,
	}
}

func NewQueryOption(measurement Measurement, conditions []QueryCondition, timerange *TimeRange, rowRange *RowRange, desc bool) queryOption {
	return queryOption{
		Measurement: measurement,
		Conditions:  conditions,
		timeRange:   timerange,
		RowRange:    rowRange,
		Desc:        desc,
	}
}

type deleteOption struct {
	Measurement Measurement
	Conditions  []QueryCondition
	timeRange   *TimeRange
}

func (d deleteOption) TimeRange() TimeRange {
	if d.timeRange == nil {
		return TimeRange{
			Start: NewUnixTime(time.Unix(0, 0)),
			Stop:  NewUnixTime(time.Now()),
		}
	}

	return TimeRange{
		Start: d.timeRange.Start,
		Stop:  d.timeRange.Stop,
	}
}

func NewDeleteOption(measurement Measurement, conditions []QueryCondition, timerange *TimeRange) deleteOption {
	return deleteOption{
		Measurement: measurement,
		Conditions:  conditions,
		timeRange:   timerange,
	}
}

// 真偽値（true:1, false:0）
type Boolean int

func NewBoolean(b bool) Boolean {
	var result Boolean

	if b {
		result = 1
	}

	return result
}

func (b *Boolean) Bool() bool {
	var result bool

	if *b == 1 {
		result = true
	}

	return result
}

func (b *Boolean) Tag() string {
	return strconv.Itoa(int(*b))
}

type UnixTime int64

func (u UnixTime) Time() time.Time {
	return time.Unix(int64(u), 0)
}

func NewUnixTime(t time.Time) UnixTime {
	return UnixTime(t.Unix())
}
