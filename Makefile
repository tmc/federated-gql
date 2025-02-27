.PHONY: generate
generate: generate-proto generate-gql

.PHONY: generate-proto
generate-proto:
	cd proto && buf generate

.PHONY: clean-gql
clean-gql:
	rm -f services/graphql-gateway/graph/generated.go
	rm -f services/graphql-gateway/graph/model/models_gen.go
	rm -f services/graphql-gateway/graph/federation.go

.PHONY: generate-gql
generate-gql: clean-gql
	cd services/graphql-gateway && go mod tidy
	cd services/graphql-gateway && go run github.com/99designs/gqlgen generate

.PHONY: generate-all
generate-all: generate-proto generate-gql

.PHONY: test
test:
	go test ./...

.PHONY: build
build: build-users build-products build-graphql-gateway

.PHONY: build-users
build-users:
	go build -o bin/users ./services/users

.PHONY: build-products
build-products:
	go build -o bin/product ./services/product

.PHONY: build-graphql-gateway
build-graphql-gateway:
	go build -o bin/graphql-gateway ./services/graphql-gateway

.PHONY: run-users
run-users:
	go run ./services/users

.PHONY: run-products
run-products:
	go run ./services/product

.PHONY: run-graphql-gateway
run-graphql-gateway:
	go run ./services/graphql-gateway
