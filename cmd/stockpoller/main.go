package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
)

var (
	secretCache, _ = secretcache.New()
)

type appSecrets struct {
	AlphaVantageApiKey string `json:"ALPHAVANTAGE_API_KEY"`
}

func HandleRequest() (string, error) {
	client := http.Client{Timeout: 5 * time.Second}

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

	resp, err := client.Get(fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_INTRADAY&symbol=IBM&interval=60min&apikey=%s", secrets.AlphaVantageApiKey))
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unable to get stock information for IBM: %v", err))
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Unable to read response body: %v", err))
	}

	return string(body), nil
}

func main() {
	lambda.Start(HandleRequest)
}
