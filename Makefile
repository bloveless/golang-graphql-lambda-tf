.PHONY: build init plan apply destroy

build:
	GOOS=linux GOARCH=amd64 go build -o build/bin/app .

init:
	terraform -chdir=infra init

plan: build
	terraform -chdir=infra plan

apply: build
	terraform -chdir=infra apply

destroy:
	terraform -chdir=infra destroy

test-mutation:
	curl -XPOST -d @test_mutation.json https://djyjb0560c.execute-api.us-west-2.amazonaws.com/dev/graphql

test-query:
	curl -XPOST -d '{"query":"query test {\n user(id:\"b5a1990f-64d4-40ed-9754-1579d75822b6\") {\n id\n name\n }\n}\n"}' https://djyjb0560c.execute-api.us-west-2.amazonaws.com/dev/graphql
