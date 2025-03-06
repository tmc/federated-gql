# GraphQL Federation Concepts

This document explains the federation concepts used in this project, which enables a unified GraphQL API across multiple microservices.

## What is GraphQL Federation?

GraphQL Federation is an architecture that allows you to split your GraphQL schema across multiple services while presenting a unified API to clients. Each service defines its own schema, and the federation layer combines them into a complete graph.

## Key Federation Concepts

### 1. Entities

Entities are objects that can be referenced across service boundaries. They have a unique identity and can be resolved by multiple services.

**In Protocol Buffers**:
```protobuf
message User {
  option (metadata.v1.entity) = true;
  
  string user_id = 1 [(metadata.v1.key) = true];
  string name = 2;
}
```

**In GraphQL**:
```graphql
type User @key(fields: "userID") {
  userID: String!
  name: String
}
```

### 2. Keys

Keys uniquely identify entities across services. They allow the federation layer to resolve references.

**In Protocol Buffers**:
```protobuf
string user_id = 1 [(metadata.v1.key) = true];
```

**In GraphQL**:
```graphql
type User @key(fields: "userID") {
  userID: String!
}
```

### 3. References

References allow one service to refer to an entity defined in another service.

**Example**: Products service referencing a User:

```graphql
type Product @key(fields: "productID") {
  productID: String!
  name: String
  price: Float
  owner: User
}

# Reference resolver in Products service
extend type User @key(fields: "userID") {
  userID: String! @external
  products: [Product]
}
```

### 4. External Fields

Fields marked as external are defined in another service but referenced in the current service.

**In Protocol Buffers**:
```protobuf
string user_id = 1 [(metadata.v1.key) = true, (metadata.v1.external) = true];
```

**In GraphQL**:
```graphql
extend type User @key(fields: "userID") {
  userID: String! @external
  products: [Product]
}
```

### 5. Extended Types

Extended types allow a service to contribute fields to a type defined in another service.

**Example**: Adding `products` field to User:

```graphql
# In Products service
extend type User @key(fields: "userID") {
  userID: String! @external
  products: [Product]
}
```

## Federation Implementation

### Protocol Buffer Metadata

This project uses Protocol Buffer extensions to define federation concepts:

```protobuf
// Identifies this message as an entity
optional bool entity = 50001;

// Identifies this field as a key field
optional bool key = 50001;

// Marks a field as external
optional bool external = 50002;

// Fields required from other services
optional string requires = 50003;

// Fields computed from other services
optional string computed_from = 50004;
```

### Code Generation

The `protoc-gen-graphql` tool reads these metadata options and generates GraphQL schemas with federation directives.

### Gateway Implementation

The GraphQL gateway uses these directives to:

1. Build a federated schema
2. Resolve references across services
3. Execute queries across multiple services
4. Combine results into unified responses

## Entity Resolution Process

1. **Client queries an entity**:
   ```graphql
   {
     user(userID: "alice") {
       name
       products {
         name
         price
       }
     }
   }
   ```

2. **Gateway gets the User from Users service**:
   ```
   GET /users/alice -> {userID: "alice", name: "Alice Johnson"}
   ```

3. **Gateway resolves products field using Products service**:
   - Products service gets a reference with just `userID: "alice"`
   - It uses this key to find Alice's products
   - Returns products list

4. **Gateway combines the results**:
   ```json
   {
     "data": {
       "user": {
         "name": "Alice Johnson",
         "products": [
           {"name": "Product 1", "price": 19.99},
           {"name": "Product 2", "price": 29.99}
         ]
       }
     }
   }
   ```

## Entity Resolvers

Entity resolvers implement the `__resolveReference` function to resolve entities from references:

```go
func (r *userResolver) User_ResolveReference(ctx context.Context, obj *model.User) (*model.User, error) {
    // obj contains just the key fields (userID)
    // Use the key to fetch the full entity
    resp, err := r.userClient.GetUser(ctx, connect.NewRequest(&userv1.GetUserRequest{
        UserId: obj.UserID,
    }))
    
    if err != nil {
        return nil, err
    }
    
    // Return the full entity
    return &model.User{
        UserID: resp.Msg.User.UserId,
        Name:   strPtr(resp.Msg.User.Name),
    }, nil
}
```

## Benefits of Federation

1. **Service independence**: Teams can develop and deploy services independently
2. **Schema modularity**: Each service defines only its relevant part of the schema
3. **Incremental adoption**: Add federation to existing services incrementally
4. **Performance optimization**: Gateway can parallelize requests to different services

## Federation Best Practices

1. **Consistent entity keys**: Use the same key fields across all services
2. **Minimize cross-service field resolution**: Optimize to reduce the number of service calls
3. **Avoid deep nesting**: Excessive nesting across services hurts performance
4. **Cache reference resolution**: Add caching for frequently resolved references
5. **Version carefully**: Coordinate changes to entity definitions across services

## Testing Federation

1. **Test individual services** against their specific schema
2. **Test the gateway** with the combined schema
3. **Test cross-service scenarios** that involve multiple services
4. **Test entity resolution** specifically

## Next Steps

- [Adding New Services](./adding-services.md)
- [Code Generation](./code-generation.md)
- [Testing Best Practices](./testing.md)