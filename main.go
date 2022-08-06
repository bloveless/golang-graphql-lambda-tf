package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
)

var mainSchema *graphql.Schema
var dynamodbClient *dynamodb.DynamoDB
var userOrdersTable string

// schema : GraphQL schema definition. This is an example schema
var schema = `
	schema {
		query: Query
		mutation: Mutation
	}
	type User {
		id: ID!
		name: String!
	}
	type Query {
		user(id: ID!): User
	}
	type Mutation {
		createUser(name: String!): User
	}
`

type user struct {
	ID   graphql.ID
	Name string
}

var users = []*user{
	{
		ID:   "1000",
		Name: "Pedro Marquez",
	},

	{
		ID:   "1001",
		Name: "John Doe",
	},
}

type userResolver struct {
	u *user
}

func (r *userResolver) ID() graphql.ID {
	return r.u.ID
}

func (r *userResolver) Name() string {
	return r.u.Name
}

// Resolver : Struct with all the resolver functions
type resolver struct{}

// User : Resolver function for the "User" query
func (r *resolver) User(args struct{ ID graphql.ID }) *userResolver {
	if p := userData[args.ID]; p != nil {
		return &userResolver{p}
	}
	return nil
}

func (r *resolver) CreateUser(args *struct{ Name string }) *userResolver {
	newUserId := graphql.ID(uuid.New().String())
	createUserInput := dynamodb.PutItemInput{
		TableName:              &userOrdersTable,
		ReturnConsumedCapacity: aws.String("TOTAL"),
		Item: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(fmt.Sprintf("user#%s", newUserId)),
			},
			"Name": {
				S: aws.String(args.Name),
			},
		},
	}

	createUserOutput, err := dynamodbClient.PutItem(&createUserInput)
	if err != nil {
		panic(errors.New(fmt.Sprintf("unable to put item in dynamo %v", err)))
	}

	log.Print("Create user output: ", createUserOutput)

	return &userResolver{&user{ID: newUserId, Name: args.Name}}
}

var userData = make(map[graphql.ID]*user)

var (
	// ErrNameNotProvided is thrown when a name is not provided
	QueryNameNotProvided = errors.New("no query was provided in the HTTP body")
)

func Handler(context context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	log.Printf("Processing Lambda request %s\n", request.RequestContext.RequestID)

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
		log.Print(request.Body)
		bodyBytes, err := base64.URLEncoding.DecodeString(request.Body)
		if err != nil {
			log.Print("unable to base64 decode request body ", err)
		}

		if err := json.Unmarshal(bodyBytes, &params); err != nil {
			log.Print("Could not decode body", err)
		}
	}

	if !request.IsBase64Encoded {
		if err := json.Unmarshal([]byte(request.Body), &params); err != nil {
			log.Print("Could not decode body", err)
		}
	}

	response := mainSchema.Exec(context, params.Query, params.OperationName, params.Variables)
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
	for _, u := range users {
		userData[u.ID] = u
	}
	mainSchema = graphql.MustParseSchema(schema, &resolver{})

	sess := session.Must(session.NewSession())
	dynamodbClient = dynamodb.New(sess, &aws.Config{})
	userOrdersTable = os.Getenv("USER_ORDERS_TABLE")
	if userOrdersTable == "" {
		panic("USER_ORDERS_TABLE environment variable is required")
	}

	lambda.Start(Handler)
}
