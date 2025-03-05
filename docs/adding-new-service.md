# Adding a New Service to the Federated GraphQL System

This guide provides step-by-step instructions for adding a new service to the federated GraphQL system.

## 1. Define the Service Protocol with Protocol Buffers

First, create a new protocol buffer file for your service in the `/proto` directory:

1. Create a directory structure for your service:
   ```bash
   mkdir -p proto/[service-name]/v1
   ```

2. Create a protocol buffer file in this directory (e.g., `proto/[service-name]/v1/[service-name].proto`):
   ```protobuf
   syntax = "proto3";
   package [service-name].v1;
   
   option go_package = "github.com/fraser-isbester/federated-gql/gen/go/[service-name]/v1;[service-name]v1";
   
   service [ServiceName]Service {
     // Define your service methods here
     rpc Get[ServiceName](Get[ServiceName]Request) returns (Get[ServiceName]Response) {}
   }
   
   // Request message for Get[ServiceName]
   message Get[ServiceName]Request {
     // Define your request fields
     string [service_name]_id = 1;
   }
   
   // Response message for Get[ServiceName]
   message Get[ServiceName]Response {
     // Define your response fields
     string [service_name]_id = 1;
     string name = 2;
     // Add additional fields as needed
   }
   ```

## 2. Generate Service Code

Use the `buf` tool to generate Go code from your protocol buffer definition:

1. Run the generation command:
   ```bash
   make generate
   ```

This will:
- Generate the Go code for your service in `/gen/go/[service-name]/v1/`
- Create the Connect-RPC handlers in `/gen/go/[service-name]/v1/[service-name]v1connect/`

## 3. Implement the Service

Create a new directory for your service implementation:

1. Create a service directory:
   ```bash
   mkdir -p services/[service-name]
   ```

2. Create a `go.mod` file for your service:
   ```bash
   cd services/[service-name]
   go mod init github.com/fraser-isbester/federated-gql/services/[service-name]
   ```

3. Create a `main.go` file with your service implementation:
   ```go
   package main
   
   import (
       "context"
       "fmt"
       "log"
       "net/http"
       "os"
   
       "connectrpc.com/connect"
       [service-name]v1 "github.com/fraser-isbester/federated-gql/gen/go/[service-name]/v1"
       "[service-name]v1connect" "github.com/fraser-isbester/federated-gql/gen/go/[service-name]/v1/[service-name]v1connect"
       "golang.org/x/net/http2"
       "golang.org/x/net/http2/h2c"
   )
   
   // [service-name]Server implements the [ServiceName]Service interface
   type [service-name]Server struct{}
   
   // Get[ServiceName] implements the [ServiceName]Service Get[ServiceName] RPC method
   func (s *[service-name]Server) Get[ServiceName](ctx context.Context, req *connect.Request[[service-name]v1.Get[ServiceName]Request]) (*connect.Response[[service-name]v1.Get[ServiceName]Response], error) {
       if req.Msg.[ServiceName]Id == "" {
           return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("[service_name]_id is required"))
       }
   
       // Implement your business logic here
       return connect.NewResponse(&[service-name]v1.Get[ServiceName]Response{
           [ServiceName]Id: req.Msg.[ServiceName]Id,
           Name:            fmt.Sprintf("[ServiceName] %s", req.Msg.[ServiceName]Id),
           // Set additional fields as needed
       }), nil
   }
   
   func main() {
       server := &[service-name]Server{}
       mux := http.NewServeMux()
       path, handler := [service-name]v1connect.New[ServiceName]ServiceHandler(server)
       mux.Handle(path, handler)
   
       port := os.Getenv("PORT")
       if port == "" {
           port = "8080"
       }
   
       addr := fmt.Sprintf(":%s", port)
       log.Printf("Starting [service-name] service on %s", addr)
   
       err := http.ListenAndServe(
           addr,
           h2c.NewHandler(mux, &http2.Server{}),
       )
       if err != nil {
           log.Fatal(err)
       }
   }
   ```

4. Set up Go module dependencies:
   ```bash
   go mod tidy
   ```

5. Add a local replacement for the generated code module:
   ```
   replace github.com/fraser-isbester/federated-gql/gen/go => ../../gen/go
   ```

## 4. Add the Service to the GraphQL Gateway

Update the GraphQL schema and resolver to integrate your new service:

1. Update the schema in `/services/graphql-gateway/graph/schema.graphqls`:
   ```graphql
   # Add your new type definition
   type [ServiceName] {
     id: ID!
     name: String!
     # Add additional fields as needed
   }
   
   # Add queries for your service
   extend type Query {
     [service-name](id: ID!): [ServiceName]
     # Add additional queries as needed
   }
   ```

2. Implement the resolver in `/services/graphql-gateway/graph/schema.resolvers.go`:
   ```go
   // Add client for your service in the Resolver struct
   type Resolver struct {
       [service-name]Client [service-name]v1connect.[ServiceName]ServiceClient
   }
   
   // Implement the resolver for your service
   func (r *queryResolver) [ServiceName](ctx context.Context, id string) (*model.[ServiceName], error) {
       resp, err := r.[service-name]Client.Get[ServiceName](ctx, connect.NewRequest(&[service-name]v1.Get[ServiceName]Request{
           [ServiceName]Id: id,
       }))
       if err != nil {
           return nil, err
       }
       
       return &model.[ServiceName]{
           ID:   resp.Msg.[ServiceName]Id,
           Name: resp.Msg.Name,
           // Map additional fields as needed
       }, nil
   }
   ```

3. Update the GraphQL gateway's `server.go` to initialize your service client:
   ```go
   // Add imports for your service
   [service-name]v1connect "github.com/fraser-isbester/federated-gql/gen/go/[service-name]/v1/[service-name]v1connect"
   
   // In the main function, initialize your service client
   [service-name]Client := [service-name]v1connect.New[ServiceName]ServiceClient(
       http.DefaultClient,
       "http://localhost:[port]", // Replace with actual service address
   )
   
   // Pass the client to the resolver
   resolver := &graph.Resolver{
       // Existing fields
       [service-name]Client: [service-name]Client,
   }
   ```

## 5. Update the Makefile

Add a new command to run your service:

```make
.PHONY: run-[service-name]
run-[service-name]:
	go run ./services/[service-name]
```

## 6. Testing the Service

1. Start your new service:
   ```bash
   make run-[service-name]
   ```

2. In another terminal, start the GraphQL gateway:
   ```bash
   make run-graphql-gateway
   ```

3. Test your service through the GraphQL playground:
   - Open a browser and navigate to `http://localhost:8080/`
   - Execute a query for your new service:
     ```graphql
     query {
       [service-name](id: "123") {
         id
         name
         # Additional fields
       }
     }
     ```

## Troubleshooting

- If code generation fails, ensure your protocol buffer file follows the correct format
- If service connection fails, verify the service is running and available at the expected address
- For GraphQL errors, check the resolver implementation and model definitions

## Best Practices

- Use singular nouns for service names (e.g., "order" not "orders")
- Follow existing service patterns for consistency
- Document your service APIs thoroughly
- Implement proper error handling
- Consider versioning your services from the start (as shown with the "v1" pattern)
- Use environment variables for configuration