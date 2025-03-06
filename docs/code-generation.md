# Code Generation Lifecycle

This document explains the code generation workflow in the federated GraphQL project, which is crucial for maintaining consistency between Protocol Buffers and GraphQL schemas.

## Overview

The project uses a multi-stage code generation process:

1. **Proto Definition**: Services and entities are defined in Protocol Buffer (.proto) files
2. **Go Code Generation**: Generate Go structs and Connect clients/servers from protos
3. **GraphQL Schema Generation**: Generate GraphQL schemas from the same protos
4. **GraphQL Resolver Generation**: Generate resolver interfaces from GraphQL schemas
5. **Manual Implementation**: Implement the resolvers to connect GraphQL and gRPC

## Proto Definition

### File Structure

Protocol Buffers are organized in the `/proto` directory:

```
/proto
  /metadata
    /v1
      metadata.proto    # Federation metadata options
  /user
    /v1
      user.proto       # User service definition
  /product
    /v1
      product.proto    # Product service definition
  buf.yaml            # Buf module configuration
  buf.gen.yaml        # Buf generation configuration
```

### Federation Metadata

The `metadata.proto` file defines options for federation:

```protobuf
// Identifies this message as an entity
optional bool entity = 50001;

// Identifies this field as a key field for the containing entity
optional bool key = 50001;

// Marks a field as external, indicating it's defined in another service
optional bool external = 50002;

// Service federation options
optional bool federated = 50001;
```

## Code Generation Tools

### Buf CLI

We use [Buf](https://buf.build/) to manage proto dependencies and code generation:

- `buf.yaml`: Module configuration
- `buf.gen.yaml`: Generation configuration

### Custom Generator: protoc-gen-graphql

The custom `protoc-gen-graphql` plugin generates GraphQL schemas from proto definitions:

- **Source**: `/tools/protoc-gen-graphql`
- **Operation**: Converts proto services and messages to GraphQL schema files
- **Template**: Uses customizable templates (default or user-provided)

## Code Generation Command Flow

### 1. Generate Proto Code

```bash
make generate-proto
```

This command:
1. Installs the `protoc-gen-graphql` tool
2. Runs `buf generate` in the `/proto` directory

Outputs:
- Go code in `/gen/go/`
- GraphQL schemas in `/gen/graphql/`

### 2. Generate GraphQL Code

```bash
make generate-gql
```

This command:
1. Cleans previous GraphQL generated files
2. Copies GraphQL schema files to the gateway
3. Runs `gqlgen generate` to create resolvers

Outputs:
- Updated resolver interfaces in the GraphQL gateway

## Code Generation Workflow

### Adding or Modifying a Service

1. **Define or update proto files**:
   ```bash
   # Edit proto files in /proto directory
   nano proto/your_service/v1/your_service.proto
   ```

2. **Generate code**:
   ```bash
   make generate-proto
   ```

3. **Copy GraphQL schemas to gateway**:
   ```bash
   cp gen/graphql/your_service.v1.YourService.graphql services/graphql-gateway/graph/schema/
   ```

4. **Generate GraphQL gateway code**:
   ```bash
   make generate-gql
   ```

5. **Implement or update resolvers**:
   ```bash
   # Edit resolver implementation
   nano services/graphql-gateway/graph/your_service.v1.YourService.resolvers.go
   ```

### Commit Strategy

When committing changes, follow this order:

1. Commit proto file changes and generated code together:
   ```bash
   git add proto/your_service/v1/your_service.proto
   git add gen/go/your_service
   git add gen/graphql/your_service.v1.YourService.graphql
   git commit -m "Define YourService proto"
   ```

2. Commit GraphQL gateway schema and generated code:
   ```bash
   git add services/graphql-gateway/graph/schema/your_service.v1.YourService.graphql
   git add services/graphql-gateway/graph/generated.go
   git commit -m "Add YourService GraphQL schema"
   ```

3. Commit resolver implementations:
   ```bash
   git add services/graphql-gateway/graph/your_service.v1.YourService.resolvers.go
   git commit -m "Implement YourService resolvers"
   ```

## Custom Templates

The `protoc-gen-graphql` generator supports custom templates:

1. Create a custom template:
   ```bash
   cp tools/protoc-gen-graphql/templates/graphql-service-schema.tmpl my-custom-template.tmpl
   # Edit my-custom-template.tmpl
   ```

2. Configure template path in `buf.gen.yaml`:
   ```yaml
   - local: protoc-gen-graphql
     out: ../gen/graphql
     opt:
       - paths=source_relative
       - template_path=/path/to/my-custom-template.tmpl
   ```

3. Generate with the custom template:
   ```bash
   make generate-proto
   ```

## Troubleshooting

### Inconsistent Generated Code

If you see inconsistencies between proto and GraphQL:

1. Clean the generated files:
   ```bash
   make clean-gql
   ```

2. Regenerate everything:
   ```bash
   make generate
   ```

### Missing Federation Directives

If federation directives are missing:

1. Verify metadata options in proto files:
   ```protobuf
   option (metadata.v1.entity) = true;
   string id = 1 [(metadata.v1.key) = true];
   ```

2. Regenerate the GraphQL schemas:
   ```bash
   make generate-proto
   make generate-gql
   ```

## Best Practices

1. **Always run the full generation cycle** before commits
2. **Keep proto definitions canonical** - they are the source of truth
3. **Use consistent metadata** across all services
4. **Test generated schemas** after generation
5. **Commit generated code** to version control

## Next Steps

- [Federation Concepts](./federation.md)
- [Testing Best Practices](./testing.md)
- [Production Deployment](./deployment.md)