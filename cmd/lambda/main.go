package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	graphqllambda "golang-graphql-lambda-tf"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type server struct {
	gql             graphqllambda.GraphQL
	userOrdersTable string
}

var (
	// ErrNameNotProvided is thrown when a name is not provided
	QueryNameNotProvided = errors.New("no query was provided in the HTTP body")
)

func (s server) Handler(context context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fmt.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

	// If no query is provided in the HTTP request body, throw an error
	if len(request.Body) < 1 {
		return events.APIGatewayProxyResponse{}, QueryNameNotProvided
	}

	var params struct {
		Query         string                 `json:"query"`
		OperationName string                 `json:"operationName"`
		Variables     map[string]interface{} `json:"variables"`
	}

	if request.IsBase64Encoded {
		fmt.Print(request.Body)
		bodyBytes, err := base64.URLEncoding.DecodeString(request.Body)
		if err != nil {
			fmt.Print("unable to base64 decode request body ", err)
		}

		if err := json.Unmarshal(bodyBytes, &params); err != nil {
			fmt.Print("Could not decode body", err)
		}
	}

	if !request.IsBase64Encoded {
		if err := json.Unmarshal([]byte(request.Body), &params); err != nil {
			fmt.Print("Could not decode body", err)
		}
	}

	response := s.gql.MainSchema.Exec(context, params.Query, params.OperationName, params.Variables)
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{}, errors.New("could not decode body")
	}

	return events.APIGatewayProxyResponse{
		Body:       string(responseJSON),
		StatusCode: 200,
	}, nil
}

func main() {
	userOrdersTable := os.Getenv("USER_ORDERS_TABLE")
	if userOrdersTable == "" {
		panic("USER_ORDERS_TABLE environment variable is required")
	}

	sess := session.Must(session.NewSession())
	ddb := dynamodb.New(sess, &aws.Config{})
	s := server{
		gql:             graphqllambda.New(ddb, userOrdersTable),
		userOrdersTable: userOrdersTable,
	}

	lambda.Start(s.Handler)
}
