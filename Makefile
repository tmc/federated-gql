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
run-users: deps-gow
	cd ./services/users; gow run .

.PHONY: run-products
run-products: deps-gow
	cd ./services/product; gow run .

.PHONY: run-graphql-gateway
run-graphql-gateway: deps-gow
	cd ./services/graphql-gateway; gow run .

.PHONY: deps-gow
deps-gow:
	@command -v gow || go install github.com/mitranim/gow@latest
	@command -v gow || (echo "gow not found in PATH. Please add $$HOME/go/bin to PATH" && exit 1)
