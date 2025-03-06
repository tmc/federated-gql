# Getting Started

This guide will help you set up and run the Federated GraphQL project.

## Prerequisites

- Go 1.24+ installed
- Protocol Buffers compiler (protoc)
- [Buf CLI](https://buf.build/docs/installation) for proto generation

## Initial Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/fraser-isbester/federated-gql.git
   cd federated-gql
   ```

2. Install gow for live reloading:
   ```bash
   go install github.com/mitranim/gow@latest
   ```

3. Install dependencies:
   ```bash
   cd proto && buf mod update && cd ..
   ```

4. Generate code:
   ```bash
   make generate
   ```

5. Build all services:
   ```bash
   make build
   ```

## Running the Services

You need to run each service separately:

1. Run the users service (in one terminal):
   ```bash
   make run-users
   ```

2. Run the product service (in another terminal):
   ```bash
   make run-products
   ```

3. Run the GraphQL gateway (in a third terminal):
   ```bash
   make run-graphql-gateway
   ```

## Accessing the GraphQL Playground

Once all services are running, access the GraphQL playground at:

```
http://localhost:8080
```

## Example Queries

Try these example GraphQL queries:

### Get User
```graphql
{
  user(userID: "alice") {
    userID
    name
  }
}
```

### Get Product
```graphql
{
  product(productID: "prod-1") {
    productID
    name
    price
  }
}
```

## Next Steps

- Learn about [adding a new service](./adding-services.md)
- Understand the [code generation lifecycle](./code-generation.md)
- Explore the [architecture](./architecture.md)