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

