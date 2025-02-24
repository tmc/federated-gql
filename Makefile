.PHONY: generate
generate:
	buf generate

.PHONY: test
test:
	go test ./...

.PHONY: run-users
run-users:
	go run ./services/users

.PHONY: run-products
run-products:
	go run ./services/products

.PHONY: run-graphql-gateway
run-graphql-gateway:
	go run ./services/graphql-gateway
