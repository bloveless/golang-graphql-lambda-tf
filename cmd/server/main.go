package main

import (
	"log"
	"net/http"
	"os"
	"stocktracker"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/graph-gophers/graphql-go/relay"
)

func main() {
	userOrdersTable := os.Getenv("USER_ORDERS_TABLE")
	if userOrdersTable == "" {
		panic("USER_ORDERS_TABLE environment variable is required")
	}

	sess := session.Must(session.NewSession())
	ddb := dynamodb.New(sess, &aws.Config{})
	gql := stocktracker.NewGraphql(ddb, userOrdersTable)

	http.Handle("/graphql", &relay.Handler{Schema: gql.MainSchema})
	log.Print("Running server on http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
