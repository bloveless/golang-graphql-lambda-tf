package stocktracker_test

import (
	"os"
	"stocktracker"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/joho/godotenv"
)

func loadEnv(t *testing.T) {
	err := godotenv.Load()
	if err != nil {
		t.Log("Error loading .env file")
		t.FailNow()
	}
}

func TestStockApi(t *testing.T) {
	loadEnv(t)

	stockApi := stocktracker.NewStockApi(os.Getenv("ALPHAVANTAGE_API_KEY"))

	sr, err := stockApi.Get("AAPL")
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	sess := session.Must(session.NewSession())
	ddb := dynamodb.New(sess, &aws.Config{})

	repo := stocktracker.NewStockRepository(os.Getenv("STOCKS_TABLE"), ddb)
	repo.UpdateItems(sr)
}
