# Adding a New Service

This guide details the process of adding a new service to the federated GraphQL architecture.

## 1. Define Protocol Buffers

### Create a New Proto Package

Start by creating a new directory in the `/proto` folder:

```bash
mkdir -p proto/your_service/v1
```

### Write the Proto Definition

Create a `.proto` file with your service definition:

```bash
touch proto/your_service/v1/your_service.proto
```

### Define the Service

Use this template for your new service:

```protobuf
syntax = "proto3";

package your_service.v1;

import "metadata/v1/metadata.proto";

option go_package = "github.com/fraser-isbester/federated-gql/gen/go/your_service/v1;yourservicev1";

// Define your service
service YourService {
  option (metadata.v1.federated) = true; // Mark as federated
  
  // Define your RPC methods
  rpc GetYourEntity(GetYourEntityRequest) returns (GetYourEntityResponse);
}

// Define your entity
message YourEntity {
  option (metadata.v1.entity) = true; // Mark as an entity
  
  string id = 1 [(metadata.v1.key) = true]; // Mark as a key field
  string name = 2;
  // Add more fields as needed
}

// Define request/response messages
message GetYourEntityRequest {
  string id = 1;
}

message GetYourEntityResponse {
  YourEntity entity = 1;
}
```

### Federation Metadata

Use these metadata options to control federation:

- `(metadata.v1.federated) = true`: Marks a service for federation
- `(metadata.v1.entity) = true`: Marks a message as an entity
- `(metadata.v1.key) = true`: Marks a field as part of the entity's key

## 2. Generate Code

### Generate Go and GraphQL Code

Run the following command to generate Go and GraphQL code:

```bash
make generate-proto
```

This will:
- Generate Go code in `/gen/go/your_service/v1`
- Generate GraphQL schema in `/gen/graphql/your_service.v1.YourService.graphql`

### Copy the GraphQL Schema to the Gateway

```bash
cp gen/graphql/your_service.v1.YourService.graphql services/graphql-gateway/graph/schema/
```

### Generate GraphQL Gateway Code

```bash
make generate-gql
```

## 3. Implement Your Service

### Create Service Directory and Setup Go Module

```bash
mkdir -p services/your_service
cd services/your_service
go mod init github.com/fraser-isbester/federated-gql/services/your_service
```

### Create Service Implementation Files

```bash
touch services/your_service/main.go
touch services/your_service/your_service_test.go
```

### Implement the Service

Use this template for your main.go:

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	yourservicev1 "github.com/fraser-isbester/federated-gql/gen/go/your_service/v1"
	"github.com/fraser-isbester/federated-gql/gen/go/your_service/v1/yourservicev1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// yourServiceServer implements the YourService interface
type yourServiceServer struct{}

// GetYourEntity implements the YourService GetYourEntity RPC method
func (s *yourServiceServer) GetYourEntity(ctx context.Context, req *connect.Request[yourservicev1.GetYourEntityRequest]) (*connect.Response[yourservicev1.GetYourEntityResponse], error) {
	// Validate input
	if req.Msg.Id == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("id is required"))
	}

	// Create a sample entity (replace with your actual implementation)
	entity := &yourservicev1.YourEntity{
		Id:   req.Msg.Id,
		Name: fmt.Sprintf("Entity %s", req.Msg.Id),
	}

	return connect.NewResponse(&yourservicev1.GetYourEntityResponse{
		Entity: entity,
	}), nil
}

func main() {
	server := &yourServiceServer{}
	mux := http.NewServeMux()
	path, handler := yourservicev1connect.NewYourServiceHandler(server)
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8083" // Choose an unused port
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting your service on %s", addr)

	err := http.ListenAndServe(
		addr,
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
```

## 4. Update the Makefile

Add entries for your new service:

```make
bin/your_service: $(shell find services/your_service -type f -name "*.go") $(shell find gen/go/your_service -type f -name "*.go")
	go build -o bin/your_service ./services/your_service

.PHONY: test-your-service
test-your-service:
	cd services/your_service && go test -v ./...

.PHONY: run-your-service
run-your-service: deps-gow
	cd ./services/your_service; gow run .
```

## 5. Implement GraphQL Resolvers

After running `make generate-gql`, implement the resolver for your service:

```go
// services/graphql-gateway/graph/your_service.v1.YourService.resolvers.go
package graph

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	yourservicev1 "github.com/fraser-isbester/federated-gql/gen/go/your_service/v1"
	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph/model"
)

// YourEntity is the resolver for the yourEntity field
func (r *queryResolver) YourEntity(ctx context.Context, id string) (*model.YourEntity, error) {
	// Construct the Connect Request
	input := &yourservicev1.GetYourEntityRequest{Id: id}

	// Make the RPC call using the Connect client
	resp, err := r.yourServiceClient.GetYourEntity(ctx, connect.NewRequest(input))
	if err != nil {
		fmt.Printf("Error fetching entity: %v\n", err)
		return nil, err
	}

	// Map the protobuf response to GraphQL model
	if resp.Msg.Entity != nil {
		return &model.YourEntity{
			ID:   resp.Msg.Entity.Id,
			Name: strPtr(resp.Msg.Entity.Name),
		}, nil
	}

	// Return nil if entity not found
	return nil, nil
}
```

### Update Resolver.go

Make sure to add your client in `services/graphql-gateway/graph/resolver.go`:

```go
//go:generate go run github.com/99designs/gqlgen generate

package graph

import (
	"github.com/fraser-isbester/federated-gql/gen/go/your_service/v1/yourservicev1connect"
	// other imports
)

// Resolver stores service clients
type Resolver struct {
	userClient        userv1connect.UserServiceClient
	productClient     productv1connect.ProductServiceClient
	yourServiceClient yourservicev1connect.YourServiceClient
}

// NewResolver creates a new resolver with initialized clients
func NewResolver() *Resolver {
	return &Resolver{
		userClient:        userv1connect.NewUserServiceClient(httpClient, "http://localhost:8082"),
		productClient:     productv1connect.NewProductServiceClient(httpClient, "http://localhost:8081"),
		yourServiceClient: yourservicev1connect.NewYourServiceClient(httpClient, "http://localhost:8083"),
	}
}
```

## 6. Test Your Service

### Write Tests

Create tests for your service in `your_service_test.go`:

```go
package main

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	yourservicev1 "github.com/fraser-isbester/federated-gql/gen/go/your_service/v1"
	"github.com/stretchr/testify/assert"
)

func TestGetYourEntity(t *testing.T) {
	server := &yourServiceServer{}
	
	// Test valid request
	req := connect.NewRequest(&yourservicev1.GetYourEntityRequest{
		Id: "test-id",
	})
	
	resp, err := server.GetYourEntity(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "test-id", resp.Msg.Entity.Id)
	
	// Test invalid request
	req = connect.NewRequest(&yourservicev1.GetYourEntityRequest{
		Id: "",
	})
	
	resp, err = server.GetYourEntity(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
```

### Verify Everything Works

1. Generate and build everything:
   ```bash
   make generate
   make build
   ```

2. Run all tests:
   ```bash
   make test
   ```

3. Run your services:
   ```bash
   make run-your-service
   make run-graphql-gateway
   ```

4. Test with a GraphQL query:
   ```graphql
   {
     yourEntity(id: "test-id") {
       id
       name
     }
   }
   ```

## Next Steps

- [Code Generation Lifecycle](./code-generation.md)
- [Federation Concepts](./federation.md)
- [Testing Best Practices](./testing.md)