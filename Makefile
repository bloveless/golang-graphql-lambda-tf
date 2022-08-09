.PHONY: build init plan apply destroy

build:
	GOOS=linux GOARCH=amd64 go build -o build/bin/graphql cmd/lambda/main.go
	GOOS=linux GOARCH=amd64 go build -o build/bin/stockpoller cmd/stockpoller/main.go

run:
	USER_ORDERS_TABLE=UserOrders-a781d5cd AWS_REGION=us-west-2 go run cmd/server/main.go

init:
	terraform -chdir=infra init

plan: build
	terraform -chdir=infra plan

apply: build
	terraform -chdir=infra apply

destroy:
	terraform -chdir=infra destroy

test-local-mutation:
	curl -XPOST -d @test_mutation.json http://localhost:8080/graphql

test-mutation:
	curl -XPOST -d @test_mutation.json https://djyjb0560c.execute-api.us-west-2.amazonaws.com/dev/graphql

test-local-query:
	curl -XPOST -d '{"query":"query test {\n user(id:\"b5a1990f-64d4-40ed-9754-1579d75822b6\") {\n id\n name\n }\n}\n"}' http://localhost:8080/graphql

test-query:
	curl -XPOST -d '{"query":"query test {\n user(id:\"b5a1990f-64d4-40ed-9754-1579d75822b6\") {\n id\n name\n }\n}\n"}' https://djyjb0560c.execute-api.us-west-2.amazonaws.com/dev/graphql
