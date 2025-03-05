.PHONY: generate
generate: generate-proto generate-gql

.PHONY: generate-proto
generate-proto:
	cd tools/protoc-gen-graphql && go install
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
test: test-users test-product test-graphql

.PHONY: test-users
test-users:
	cd services/users && go test -v ./...

.PHONY: test-product
test-product:
	cd services/product && go test -v ./...

.PHONY: test-graphql
test-graphql:
	cd services/graphql-gateway && go test -v ./...

.PHONY: build
build: bin/users bin/product bin/graphql-gateway

# File-based targets to track dependencies
bin/users: $(shell find services/users -type f -name "*.go") $(shell find gen/go/user -type f -name "*.go")
	go build -o bin/users ./services/users

bin/product: $(shell find services/product -type f -name "*.go") $(shell find gen/go/product -type f -name "*.go")
	go build -o bin/product ./services/product

bin/graphql-gateway: $(shell find services/graphql-gateway -type f -name "*.go") $(shell find gen/go -type f -name "*.go")
	go build -o bin/graphql-gateway ./services/graphql-gateway

.PHONY: run-users
run-users:
	go run ./services/users

.PHONY: run-products
run-products:
	go run ./services/product

.PHONY: run-order
run-order:
	go run ./services/order

.PHONY: run-graphql-gateway
run-graphql-gateway:
	go run ./services/graphql-gateway
