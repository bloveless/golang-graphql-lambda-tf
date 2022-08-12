package stocktracker

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type StockRepository struct {
	trackedStocksTable string
	stockTable         string
	ddbClient          *dynamodb.DynamoDB
}

func NewStockRepository(trackedStocksTable string, stockTable string, ddbClient *dynamodb.DynamoDB) StockRepository {
	return StockRepository{
		trackedStocksTable: trackedStocksTable,
		stockTable:         stockTable,
		ddbClient:          ddbClient,
	}
}

type stockKey struct {
	PK string `json:"PK"`
	SK string `json:"SK"`
}

type stockValue struct {
	Symbol string  `json:":symbol"`
	High   float64 `json:":high"`
	Low    float64 `json:":low"`
	Open   float64 `json:":open"`
	Close  float64 `json:":close"`
	Volume float64 `json:":volume"`
	Now    string  `json:":now"`
}

type TrackedStock struct {
	Enabled    string `json:"enabled"`
	LastPolled string `json:"last_polled"`
	Symbol     string `json:"symbol"`
}

func (r StockRepository) GetOldestTrackedStock() (TrackedStock, error) {
	yes := struct {
		Yes string `json:":yes"`
	}{Yes: "yes"}
	values, err := dynamodbattribute.MarshalMap(yes)
	if err != nil {
		return TrackedStock{}, fmt.Errorf("unable to marshal yes: %w", err)
	}

	qi := &dynamodb.QueryInput{
		TableName:                 &r.trackedStocksTable,
		KeyConditionExpression:    aws.String("enabled = :yes"),
		ExpressionAttributeValues: values,
		Limit:                     aws.Int64(1),
	}

	out, err := r.ddbClient.Query(qi)
	if err != nil {
		return TrackedStock{}, fmt.Errorf("unable to update the oldest tracked stock: %w", err)
	}

	if *out.Count > 0 {
		ts := TrackedStock{
			Enabled:    *out.Items[0]["enabled"].S,
			LastPolled: *out.Items[0]["last_polled"].S,
			Symbol:     *out.Items[0]["symbol"].S,
		}

		return ts, nil
	}

	return TrackedStock{}, fmt.Errorf("unable to find an oldest stock")
}

func (r StockRepository) TouchTrackedStock(ts TrackedStock) error {
	tsk := struct {
		Enabled    string `json:"enabled"`
		LastPolled string `json:"last_polled"`
	}{
		Enabled:    "yes",
		LastPolled: ts.LastPolled,
	}

	key, err := dynamodbattribute.MarshalMap(tsk)
	if err != nil {
		return fmt.Errorf("unable to generate key for updating tracked stock value: %w", err)
	}

	dii := &dynamodb.DeleteItemInput{
		TableName: aws.String(r.trackedStocksTable),
		Key:       key,
	}

	_, err = r.ddbClient.DeleteItem(dii)
	if err != nil {
		return fmt.Errorf("unable to delete old tracked stock value during touch: %w", err)
	}

	v := struct {
		Enabled    string `json:"enabled"`
		LastPolled string `json:"last_polled"`
		Symbol     string `json:"symbol"`
	}{
		Enabled:    ts.Enabled,
		LastPolled: time.Now().UTC().Format(time.RFC3339),
		Symbol:     ts.Symbol,
	}

	values, err := dynamodbattribute.MarshalMap(v)
	if err != nil {
		return fmt.Errorf("unable to generate values for updating tracked stock value: %w", err)
	}

	pii := &dynamodb.PutItemInput{
		TableName: aws.String(r.trackedStocksTable),
		// Key:                       key,
		Item: values,
		// ReturnValues: aws.String("UPDATED_NEW"),
		// UpdateExpression:          aws.String("SET enabled = :enabled, last_polled = :last_polled, symbol = :symbol"),
	}

	_, err = r.ddbClient.PutItem(pii)
	if err != nil {
		return fmt.Errorf("unable to update dynamodb for touching tracked stock: %w", err)
	}

	return nil
}

func (r StockRepository) UpdateItems(sr StockResponse) error {

	for dateTime, data := range sr.TimeSeries {

		sk := stockKey{
			PK: fmt.Sprintf("stockvalue#%s", sr.Symbol),
			SK: dateTime.UTC().Format(time.RFC3339),
		}

		key, err := dynamodbattribute.MarshalMap(sk)
		if err != nil {
			return fmt.Errorf("unable to update  create dynamodb item key from %+v: %w", sk, err)
		}

		sv := stockValue{
			Symbol: sr.Symbol,
			High:   data.High,
			Low:    data.Low,
			Open:   data.Open,
			Close:  data.Close,
			Volume: data.Volume,
			Now:    time.Now().UTC().Format(time.RFC3339),
		}

		values, err := dynamodbattribute.MarshalMap(sv)
		if err != nil {
			return fmt.Errorf("unable to create dynamodb item values from %+v: %w", sv, err)
		}

		uii := &dynamodb.UpdateItemInput{
			TableName:                 aws.String(r.stockTable),
			Key:                       key,
			ExpressionAttributeValues: values,
			ReturnValues:              aws.String("UPDATED_NEW"),
			UpdateExpression:          aws.String("SET symbol = :symbol, high_usd = :high, low_usd = :low, open_usd = :open, close_usd = :close, volume = :volume, created_at = if_not_exists(created_at, :now), modified_at = :now"),
		}

		_, err = r.ddbClient.UpdateItem(uii)
		if err != nil {
			return fmt.Errorf("unable to update dynamodb: %w", err)
		}

		fmt.Printf("Updated item: %s\n", sk.SK)
	}

	return nil
}
