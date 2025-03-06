# Testing Best Practices

This document outlines the best practices for testing the federated GraphQL architecture.

## Testing Overview

The project includes tests at multiple levels:
1. **Unit tests**: Test individual functions and methods
2. **Service tests**: Test each microservice in isolation
3. **GraphQL resolver tests**: Test GraphQL resolvers with mock services
4. **Integration tests**: Test the entire system with real services

## Service Testing

### gRPC Service Tests

Each gRPC service should have comprehensive tests for all RPC methods:

```go
func TestGetUser(t *testing.T) {
	server := &userServer{}
	
	// Test valid request
	req := connect.NewRequest(&userv1.GetUserRequest{
		UserId: "alice",
	})
	
	resp, err := server.GetUser(context.Background(), req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "alice", resp.Msg.User.UserId)
	assert.Equal(t, "Alice Johnson", resp.Msg.User.Name)
	
	// Test invalid request
	req = connect.NewRequest(&userv1.GetUserRequest{
		UserId: "",
	})
	
	resp, err = server.GetUser(context.Background(), req)
	assert.Error(t, err)
	assert.Nil(t, resp)
}
```

### Test Coverage Goals

- Test each RPC method with valid inputs
- Test validation logic with invalid inputs
- Test error conditions and edge cases

## GraphQL Resolver Testing

### Testing Resolvers

Use the testing utilities from gqlgen to test resolvers:

```go
func TestQueryResolver_User(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	
	// Setup mock expectation
	mockUserClient.EXPECT().
		GetUser(gomock.Any(), gomock.Any()).
		Return(&connect.Response[userv1.GetUserResponse]{
			Msg: &userv1.GetUserResponse{
				User: &userv1.User{
					UserId: "alice",
					Name:   "Alice Johnson",
				},
			},
		}, nil)
	
	// Create resolver with mock client
	r := &Resolver{
		userClient: mockUserClient,
	}
	
	// Execute query
	user, err := r.Query().User(context.Background(), "alice")
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "alice", user.UserID)
	assert.Equal(t, "Alice Johnson", *user.Name)
}
```

### Testing Entity Resolvers

Test entity resolution specifically:

```go
func TestUserResolver_ResolveReference(t *testing.T) {
	// Setup
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	mockUserClient := mocks.NewMockUserServiceClient(ctrl)
	
	// Setup mock expectation
	mockUserClient.EXPECT().
		GetUser(gomock.Any(), connect.NewRequest(&userv1.GetUserRequest{
			UserId: "alice",
		})).
		Return(&connect.Response[userv1.GetUserResponse]{
			Msg: &userv1.GetUserResponse{
				User: &userv1.User{
					UserId: "alice",
					Name:   "Alice Johnson",
				},
			},
		}, nil)
	
	// Create resolver with mock client
	r := &Resolver{
		userClient: mockUserClient,
	}
	
	// Create a reference with just the key
	ref := &model.User{
		UserID: "alice",
	}
	
	// Resolve the reference
	user, err := r.User().User_ResolveReference(context.Background(), ref)
	
	// Assertions
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, "alice", user.UserID)
	assert.Equal(t, "Alice Johnson", *user.Name)
}
```

## Integration Testing

### Testing the Full GraphQL API

Use a test client to make real GraphQL queries:

```go
func TestUserQuery(t *testing.T) {
	// Start test server
	srv := startTestServer()
	defer srv.Close()
	
	// Create GraphQL client
	client := graphql.NewClient(srv.URL)
	
	// Define query
	req := graphql.NewRequest(`
		query GetUser($id: String!) {
			user(userID: $id) {
				userID
				name
			}
		}
	`)
	req.Var("id", "alice")
	
	// Execute query
	var resp struct {
		User struct {
			UserID string `json:"userID"`
			Name   string `json:"name"`
		} `json:"user"`
	}
	err := client.Run(context.Background(), req, &resp)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "alice", resp.User.UserID)
	assert.Equal(t, "Alice Johnson", resp.User.Name)
}
```

### Testing Federation Scenarios

Test cross-service scenarios specifically:

```go
func TestUserWithProducts(t *testing.T) {
	// Start test server
	srv := startTestServer()
	defer srv.Close()
	
	// Create GraphQL client
	client := graphql.NewClient(srv.URL)
	
	// Define query that spans multiple services
	req := graphql.NewRequest(`
		query GetUserWithProducts($id: String!) {
			user(userID: $id) {
				userID
				name
				products {
					productID
					name
					price
				}
			}
		}
	`)
	req.Var("id", "alice")
	
	// Execute query
	var resp struct {
		User struct {
			UserID   string `json:"userID"`
			Name     string `json:"name"`
			Products []struct {
				ProductID string  `json:"productID"`
				Name      string  `json:"name"`
				Price     float64 `json:"price"`
			} `json:"products"`
		} `json:"user"`
	}
	err := client.Run(context.Background(), req, &resp)
	
	// Assertions
	assert.NoError(t, err)
	assert.Equal(t, "alice", resp.User.UserID)
	assert.NotEmpty(t, resp.User.Products)
}
```

## Mocking

### Creating Mocks

Use gomock to generate mock clients:

```bash
# Install mockgen
go install go.uber.org/mock/mockgen@latest

# Generate mocks for a service client
mockgen -destination=mocks/mock_user_client.go \
  -package=mocks \
  github.com/fraser-isbester/federated-gql/gen/go/user/v1/userv1connect UserServiceClient
```

### Using Mocks in Tests

```go
func TestWithMocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	
	// Create mock client
	mockClient := mocks.NewMockUserServiceClient(ctrl)
	
	// Set expectations
	mockClient.EXPECT().
		GetUser(gomock.Any(), gomock.Any()).
		Return(&connect.Response[userv1.GetUserResponse]{
			Msg: &userv1.GetUserResponse{
				User: &userv1.User{
					UserId: "test-user",
					Name:   "Test User",
				},
			},
		}, nil)
	
	// Use mock in your test
	// ...
}
```

## Test Organization

Organize tests by service and functionality:

```
/services
  /users
    main.go
    users_test.go       # Service-specific tests
  /product
    main.go
    product_test.go     # Service-specific tests
  /graphql-gateway
    server.go
    server_test.go      # Server tests
    /graph
      resolver_test.go  # Resolver tests
      product.v1.ProductService.resolvers_test.go
      user.v1.UserService.resolvers_test.go
/integration_tests      # Cross-service tests
  federation_test.go    # Tests spanning multiple services
```

## Continuous Integration

### CI Pipeline Steps

1. **Lint code**: Run gofmt and golint
2. **Generate code**: Verify codegen is up-to-date
3. **Unit tests**: Run all unit tests
4. **Integration tests**: Run integration tests
5. **Code coverage**: Generate coverage reports

### Example CI Configuration

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [ main ]
  pull_request:
    branches: [ main ]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    - uses: actions/setup-go@v3
      with:
        go-version: '1.24'
    
    - name: Install dependencies
      run: |
        go install github.com/mitranim/gow@latest
        go install github.com/bufbuild/buf/cmd/buf@latest
    
    - name: Verify generated code
      run: |
        make generate
        git diff --exit-code
    
    - name: Run tests
      run: make test
```

## Performance Testing

### Benchmarking Resolvers

```go
func BenchmarkUserResolver(b *testing.B) {
	r := NewResolver()
	
	for i := 0; i < b.N; i++ {
		user, _ := r.Query().User(context.Background(), "alice")
		if user == nil {
			b.Fatal("Expected user not to be nil")
		}
	}
}
```

### Load Testing

Use tools like k6 to test the API under load:

```js
// k6-script.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export default function() {
  const query = `
    query {
      user(userID: "alice") {
        userID
        name
        products {
          productID
          name
        }
      }
    }
  `;
  
  const res = http.post('http://localhost:8080/query', 
    JSON.stringify({ query: query }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'has user data': (r) => r.json('data.user') !== null,
  });
  
  sleep(1);
}
```

## Test Coverage

Aim for high test coverage:

- 90%+ for core service logic
- 80%+ for GraphQL resolvers
- 70%+ for overall codebase

Generate coverage reports:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Next Steps

- [Deployment Guide](./deployment.md)
- [Adding New Services](./adding-services.md)
- [Code Generation](./code-generation.md)