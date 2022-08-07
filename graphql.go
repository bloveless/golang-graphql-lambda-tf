package graphqllambda

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
	"github.com/graph-gophers/graphql-go"
)

var mainSchema *graphql.Schema

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
type Resolver struct {
	ddb             *dynamodb.DynamoDB
	userOrdersTable string
}

// User : Resolver function for the "User" query
func (r *Resolver) User(args struct{ ID graphql.ID }) *userResolver {
	getUserInput := dynamodb.GetItemInput{
		TableName: &r.userOrdersTable,
		Key: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(fmt.Sprintf("user#%s", args.ID)),
			},
		},
	}

	getUserOutput, err := r.ddb.GetItem(&getUserInput)
	if err != nil {
		panic(errors.New(fmt.Sprintf("unable to get item from dynamo %v", err)))
	}

	if getUserOutput != nil && getUserOutput.Item != nil {
		id := strings.Split(*getUserOutput.Item["PK"].S, "#")
		return &userResolver{&user{
			ID:   graphql.ID(id[1]),
			Name: *getUserOutput.Item["Name"].S,
		}}
	}

	return nil
}

func (r *Resolver) CreateUser(args *struct{ Name string }) *userResolver {
	newUserId := graphql.ID(uuid.New().String())
	createUserInput := dynamodb.PutItemInput{
		TableName:              &r.userOrdersTable,
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

	fmt.Print("create user input", createUserInput)

	createUserOutput, err := r.ddb.PutItem(&createUserInput)
	if err != nil {
		panic(errors.New(fmt.Sprintf("unable to put item in dynamo %v", err)))
	}

	fmt.Print("Create user output: ", createUserOutput)

	return &userResolver{&user{ID: newUserId, Name: args.Name}}
}

var userData = make(map[graphql.ID]*user)

type GraphQL struct {
	MainSchema      *graphql.Schema
	userOrdersTable string
}

func New(ddb *dynamodb.DynamoDB, userOrdersTable string) GraphQL {
	return GraphQL{
		MainSchema: graphql.MustParseSchema(schema, &Resolver{ddb, userOrdersTable}),
	}
}
