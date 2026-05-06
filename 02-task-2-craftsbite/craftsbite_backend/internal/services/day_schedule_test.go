//go:build integration

package services_test

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/sayad-ika/craftsbite/internal/repository"
	"github.com/sayad-ika/craftsbite/internal/services"
)

func testTable() string {
	if t := os.Getenv("DYNAMODB_TABLE"); t != "" {
		return t
	}
	return "craftsbite-test"
}

func newTestClient(t *testing.T) *dynamodb.Client {
	t.Helper()
	endpoint := os.Getenv("DYNAMODB_ENDPOINT")
	if endpoint == "" {
		endpoint = "http://localhost:8000"
	}

	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("ap-southeast-1"),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
		config.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint}, nil
			}),
		),
	)
	if err != nil {
		t.Fatalf("failed to create test DynamoDB config: %v", err)
	}
	return dynamodb.NewFromConfig(cfg)
}

func ensureTable(t *testing.T, client *dynamodb.Client, table string) {
	t.Helper()
	_, err := client.CreateTable(context.Background(), &dynamodb.CreateTableInput{
		TableName:   aws.String(table),
		BillingMode: types.BillingModePayPerRequest,
		AttributeDefinitions: []types.AttributeDefinition{
			{AttributeName: aws.String("PK"), AttributeType: types.ScalarAttributeTypeS},
			{AttributeName: aws.String("SK"), AttributeType: types.ScalarAttributeTypeS},
		},
		KeySchema: []types.KeySchemaElement{
			{AttributeName: aws.String("PK"), KeyType: types.KeyTypeHash},
			{AttributeName: aws.String("SK"), KeyType: types.KeyTypeRange},
		},
	})
	// Ignore "already exists" errors.
	if err != nil {
		log.Printf("ensureTable: %v (may already exist)", err)
	}
}

func cleanDates(t *testing.T, client *dynamodb.Client, table string, dates []string) {
	t.Helper()
	for _, d := range dates {
		_ = repository.DeleteDaySchedule(context.Background(), client, table, d)
	}
}

func nextWeekdayDate(weekday time.Weekday) string {
	d := time.Now().UTC().AddDate(0, 0, 2) // start from day-after-tomorrow to avoid cutoff issues
	for d.Weekday() != weekday {
		d = d.AddDate(0, 0, 1)
	}
	return d.Format("2006-01-02")
}

func nextSaturdayDate() string {
	d := time.Now().UTC().AddDate(0, 0, 1)
	for d.Weekday() != time.Saturday {
		d = d.AddDate(0, 0, 1)
	}
	return d.Format("2006-01-02")
}

func nextSundayDate() string {
	d := time.Now().UTC().AddDate(0, 0, 1)
	for d.Weekday() != time.Sunday {
		d = d.AddDate(0, 0, 1)
	}
	return d.Format("2006-01-02")
}

func TestBulkSetDaySchedule_AllWeekends(t *testing.T) {
	client := newTestClient(t)
	table := testTable()
	ensureTable(t, client, table)
	store := repository.NewStore(client, table)

	sat := nextSaturdayDate()
	sun := nextSundayDate()
	dates := []string{sat, sun}
	cleanDates(t, client, table, dates)
	t.Cleanup(func() { cleanDates(t, client, table, dates) })

	input := services.SetDayScheduleInput{
		DayStatus: "normal",
		SetBy:     "test-admin",
	}

	_, err := services.BulkSetDaySchedule(context.Background(), store, dates, input)
	if err == nil {
		t.Fatal("expected ErrAllWeekend, got nil")
	}
	if err != services.ErrAllWeekend {
		t.Fatalf("expected ErrAllWeekend, got: %v", err)
	}

	for _, d := range dates {
		schedule, err2 := store.GetDay(context.Background(), d)
		if err2 != nil {
			t.Fatalf("store.GetDay(%s): %v", d, err2)
		}
		if schedule != nil {
			t.Errorf("expected no record for %s (weekend), but one was written", d)
		}
	}
}

func TestBulkSetDaySchedule_WeekendSkipping(t *testing.T) {
	client := newTestClient(t)
	table := testTable()
	ensureTable(t, client, table)
	store := repository.NewStore(client, table)

	mon := nextWeekdayDate(time.Monday)
	sat := nextSaturdayDate()
	sun := nextSundayDate()
	tue := nextWeekdayDate(time.Tuesday)

	allDates := []string{mon, sat, sun, tue}
	cleanDates(t, client, table, allDates)
	t.Cleanup(func() { cleanDates(t, client, table, allDates) })

	input := services.SetDayScheduleInput{
		DayStatus:      "normal",
		AvailableMeals: []string{"lunch"},
		SetBy:          "test-admin",
	}

	result, err := services.BulkSetDaySchedule(context.Background(), store, allDates, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.SuccessDates) != 2 {
		t.Errorf("expected 2 success dates (mon, tue), got %d: %v", len(result.SuccessDates), result.SuccessDates)
	}

	for _, weekend := range []string{sat, sun} {
		schedule, err2 := store.GetDay(context.Background(), weekend)
		if err2 != nil {
			t.Fatalf("store.GetDay(%s): %v", weekend, err2)
		}
		if schedule != nil {
			t.Errorf("weekend date %s should not have a record, but one was written", weekend)
		}
	}

	for _, weekday := range []string{mon, tue} {
		schedule, err2 := store.GetDay(context.Background(), weekday)
		if err2 != nil {
			t.Fatalf("store.GetDay(%s): %v", weekday, err2)
		}
		if schedule == nil {
			t.Errorf("weekday %s should have a record, but none was found", weekday)
		}
	}
}

func TestBulkSetDaySchedule_Success(t *testing.T) {
	client := newTestClient(t)
	table := testTable()
	ensureTable(t, client, table)
	store := repository.NewStore(client, table)

	mon := nextWeekdayDate(time.Monday)
	fri := nextWeekdayDate(time.Friday)
	dates := []string{mon, fri}
	cleanDates(t, client, table, dates)
	t.Cleanup(func() { cleanDates(t, client, table, dates) })

	input := services.SetDayScheduleInput{
		DayStatus:      "event_day",
		AvailableMeals: []string{"lunch", "event_dinner"},
		SetBy:          "test-admin",
	}

	result, err := services.BulkSetDaySchedule(context.Background(), store, dates, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.SuccessDates) != 2 {
		t.Errorf("expected 2 success dates, got %d", len(result.SuccessDates))
	}
	for _, d := range dates {
		found := false
		for _, sd := range result.SuccessDates {
			if sd == d {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("date %s missing from SuccessDates", d)
		}
	}
}

func TestBulkSetDaySchedule_InvalidDateFormat(t *testing.T) {
	client := newTestClient(t)
	table := testTable()
	ensureTable(t, client, table)
	store := repository.NewStore(client, table)

	input := services.SetDayScheduleInput{
		DayStatus: "normal",
		SetBy:     "test-admin",
	}

	_, err := services.BulkSetDaySchedule(context.Background(), store, []string{"not-a-date"}, input)
	if err == nil {
		t.Fatal("expected error for invalid date format, got nil")
	}
}
