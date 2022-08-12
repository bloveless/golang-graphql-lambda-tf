package stocktracker

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type StockRepository struct {
	stockTable string
	ddbClient  *dynamodb.DynamoDB
}

func NewStockRepository(stockTable string, ddbClient *dynamodb.DynamoDB) StockRepository {
	return StockRepository{
		stockTable: stockTable,
		ddbClient:  ddbClient,
	}
}

type stockKey struct {
	PK string `json:"PK"`
	SK string `json:"SK"`
}

type stockValue struct {
	High   float64 `json:":high"`
	Low    float64 `json:":low"`
	Open   float64 `json:":open"`
	Close  float64 `json:":close"`
	Volume float64 `json:":volume"`
	Now    string  `json:":now"`
}

func (r StockRepository) UpdateItems(sr StockResponse) error {

	for dateTime, data := range sr.TimeSeries {

		sk := stockKey{
			PK: fmt.Sprintf("stockvalue#%s#%s", sr.Symbol, dateTime.UTC().Format(time.RFC3339)),
			SK: dateTime.UTC().Format(time.RFC3339),
		}

		key, err := dynamodbattribute.MarshalMap(sk)
		if err != nil {
			return fmt.Errorf("unable to create dynamodb item key from %+v: %w", sk, err)
		}

		sv := stockValue{
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
			UpdateExpression:          aws.String("SET highprice = :high, lowprice = :low, openprice = :open, closeprice = :close, volume = :volume, createdat = if_not_exists(createdat, :now), modifiedat = :now"),
		}

		_, err = r.ddbClient.UpdateItem(uii)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Updated item: %s\n", sk.SK)
	}

	return nil
}
