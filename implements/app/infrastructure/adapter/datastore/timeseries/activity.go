package timeseries

import (
	llog "app/lib/log"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	influxdb2 "github.com/influxdata/influxdb-client-go"
	"github.com/influxdata/influxdb-client-go/api/write"
	"github.com/influxdata/influxdb-client-go/domain"
	"github.com/samber/do"
	"github.com/samber/lo"
)

type TimeseriesStore interface {
	Save(c context.Context, point Point) error
	Count(c context.Context, option queryOption) (int, error)
	Find(c context.Context, option queryOption, v any) error
	Delete(c context.Context, option deleteOption) error
}

type activityStore struct {
	client influxdb2.Client
	org    *domain.Organization
	bucket *domain.Bucket
}

// Count implements TimeseriesStore.
func (a activityStore) Count(c context.Context, option queryOption) (int, error) {
	conditions := []string{fmt.Sprintf("r._measurement==\"%v\"", option.Measurement)}

	for _, c := range option.Conditions {
		switch c.Ope {
		case Contains:
			if values, ok := c.Value.([]string); ok {
				set := fmt.Sprintf("[%v]", strings.Join(lo.Map(values, func(value string, _ int) string { return fmt.Sprintf("\"%v\"", value) }), " , "))
				conditions = append(conditions, fmt.Sprintf("contains(value: r.%v, set: %v)", c.Key, set))
			}
		default:
			conditions = append(conditions, fmt.Sprintf("r.%v%v\"%v\"", c.Key, c.Ope, c.Value))
		}
	}

	timeRange := option.TimeRange()

	countQuery := fmt.Sprintf(`
		from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => %v)
			|> keep(columns: ["_time", "_field", "_value"])
			|> pivot(rowKey:["_time"], columnKey:["_field"], valueColumn:"_value")
			|> count(column: "at")
			|> rename(columns: {"at": "count"})
	`,
		a.bucket.Name,
		timeRange.Start,
		timeRange.Stop,
		strings.Join(conditions, " and "),
	)

	countResult := struct {
		Value int `json:"count"`
	}{}

	result, err := a.client.QueryAPI(a.org.Name).Query(c, countQuery)

	if err != nil {
		return 0, err
	}

	records := map[string]interface{}{}

	for result.Next() {
		records = result.Record().Values()
	}

	if result.Err() != nil {
		return 0, err
	}

	llog.Debug(c, "%v", records)

	if len(records) == 0 {
		return 0, nil
	}

	bytes, err := json.Marshal(records)

	if err != nil {
		return 0, err
	}

	if err := json.Unmarshal(bytes, &countResult); err != nil {
		return 0, err
	}

	return countResult.Value, nil
}

// Delete implements TimeseriesStore.
func (a activityStore) Delete(c context.Context, option deleteOption) error {
	conditions := []string{fmt.Sprintf("_measurement=\"%v\"", option.Measurement)}

	for _, c := range option.Conditions {
		conditions = append(conditions, fmt.Sprintf("r.%v%v\"%v\"", c.Key, c.Ope, c.Value))
	}

	timeRange := option.TimeRange()

	condition := strings.Join(conditions, " AND ")

	llog.Debug(c, "%v", condition)

	return a.client.DeleteAPI().Delete(c, a.org, a.bucket, timeRange.Start.Time(), timeRange.Stop.Time(), condition)
}

// Find implements TimeseriesStore.
func (a activityStore) Find(c context.Context, option queryOption, v any) error {
	conditions := []string{fmt.Sprintf("r._measurement==\"%v\"", option.Measurement)}

	for _, c := range option.Conditions {
		switch c.Ope {
		case Contains:
			if values, ok := c.Value.([]string); ok {
				set := fmt.Sprintf("[%v]", strings.Join(lo.Map(values, func(value string, _ int) string { return fmt.Sprintf("\"%v\"", value) }), " , "))
				conditions = append(conditions, fmt.Sprintf("contains(value: r.%v, set: %v)", c.Key, set))
			}
		default:
			conditions = append(conditions, fmt.Sprintf("r.%v%v\"%v\"", c.Key, c.Ope, c.Value))
		}
	}

	timeRange := option.TimeRange()

	var limit string
	if option.RowRange != nil {
		limit = fmt.Sprintf("|> limit(n: %v, offset: %v)", option.RowRange.Limit, option.RowRange.Offset)
	} else {
		limit = ""
	}

	query := fmt.Sprintf(`
		from(bucket: "%v")
			|> range(start: %v, stop: %v)
			|> filter(fn: (r) => %v)
			|> keep(columns: ["_time", "_field", "_value"])
			|> pivot(rowKey:["_time"], columnKey:["_field"], valueColumn:"_value")
			|> sort(columns:["_time"], desc: %v)
			%v
	`,
		a.bucket.Name,
		timeRange.Start,
		timeRange.Stop,
		strings.Join(conditions, " and "),
		option.Desc,
		limit,
	)

	result, err := a.client.QueryAPI(a.org.Name).Query(c, query)

	if err != nil {
		return err
	}

	records := []map[string]interface{}{}

	for result.Next() {
		records = append(records, result.Record().Values())
	}

	if result.Err() != nil {
		return err
	}

	if len(records) == 0 {
		return nil
	}

	bytes, err := json.Marshal(records)

	if err != nil {
		return err
	}

	if err := json.Unmarshal(bytes, v); err != nil {
		return err
	}

	return nil
}

// Save implements TimeseriesStore.
func (a activityStore) Save(c context.Context, point Point) error {
	writeAPI := a.client.WriteAPIBlocking(a.org.Name, a.bucket.Name)
	return writeAPI.WritePoint(c,
		write.NewPoint(string(point.Measurement()),
			point.Tags(),
			point.Fields(),
			point.Timestamp(),
		),
	)
}

func NewActivityStore(i *do.Injector) (TimeseriesStore, error) {
	client := influxdb2.NewClient(os.Getenv("INFLUXDB_ACTIVITY_URL"), os.Getenv("INFLUXDB_ACTIVITY_AUTH_TOKEN"))

	org, err := client.OrganizationsAPI().FindOrganizationByName(context.Background(), os.Getenv("INFLUXDB_ACTIVITY_ORG"))

	if err != nil {
		return nil, err
	}

	bucket, err := client.BucketsAPI().FindBucketByName(context.Background(), os.Getenv("INFLUXDB_ACTIVITY_BUCKET"))

	if err != nil {
		return nil, err
	}

	fmt.Printf("connection established. %v \n", os.Getenv("INFLUXDB_ACTIVITY_URL"))

	return activityStore{
		client: client,
		org:    org,
		bucket: bucket,
	}, nil
}
