package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"stocktracker"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

var (
	secretCache, _ = secretcache.New()
)

type appSecrets struct {
	AlphaVantageApiKey string `json:"ALPHAVANTAGE_API_KEY"`
}

func HandleRequest() (string, error) {
	secretId := "prod/GraphQLStocks"
	secretCache, err := secretcache.New()
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unable to create new secrets cache: %v", err))
	}

	result, err := secretCache.GetSecretString(secretId)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unable to get secret: %v", err))
	}

	var secrets appSecrets
	err = json.Unmarshal([]byte(result), &secrets)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unable to unmarshal secret json: %v", err))
	}

	sess := session.Must(session.NewSession())
	ddb := dynamodb.New(sess, &aws.Config{})
	repo := stocktracker.NewStockRepository(os.Getenv("TRACKED_STOCKS_TABLE"), os.Getenv("STOCKS_TABLE"), ddb)

	ots, err := repo.GetOldestTrackedStock()
	if err != nil {
		return "", fmt.Errorf("unable to get the oldest tracked stock: %w", err)
	}

	err = repo.TouchTrackedStock(ots)
	if err != nil {
		return "", fmt.Errorf("unable to touch the oldest tracked stock: %w", err)
	}

	stockApi := stocktracker.NewStockApi(secrets.AlphaVantageApiKey)
	sr, err := stockApi.Get(ots.Symbol)
	if err != nil {
		return "", fmt.Errorf("unable to get stock history for %s: %w", ots.Symbol, err)
	}

	err = repo.UpdateItems(sr)
	if err != nil {
		return "", fmt.Errorf("unable to update items in dynamodb: %w", err)
	}

	fmt.Printf("Stock Results for %s: %+v\n", ots.Symbol, sr)

	return fmt.Sprintf("Stock results: %v", sr), nil
}

func main() {
	lambda.Start(HandleRequest)
}
