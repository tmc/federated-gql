# Federated GraphQL
A proof of concept for federating gRPC services with GraphQL federations.

## Project Overview

This project demonstrates how to build a federated GraphQL API that aggregates multiple gRPC microservices. It uses a code generation approach to maintain consistency between protocol buffers and GraphQL schemas.

### Architecture

- **gRPC Services**: Independent microservices with Connect API
- **GraphQL Gateway**: Federation layer that composes all services
- **Code Generation**: Automated tools to keep everything in sync

## Quick Start

1. **Build all services**:
   ```bash
   make build
   ```

2. **Run the services**:
   ```bash
   # Each in a separate terminal
   make run-users
   make run-products
   make run-graphql-gateway
   ```

3. **Access the GraphQL Playground**: 
   Open http://localhost:8080 in your browser

## Development Guide

For detailed development instructions, see [CLAUDE.md](CLAUDE.md), which contains:
- Command reference
- Development workflow
- Adding new services
- Code generation lifecycle
- Project structure

## Code Generation
This project uses protoc plugins to generate code from Protocol Buffer definitions.

### Proto to GraphQL Pipeline

1. Define services and entities in Protocol Buffers
2. Generate Go code with buf and Connect
3. Generate GraphQL schemas with custom protoc plugin
4. Implement GraphQL resolvers in the gateway

### protoc-gen-graphql
The `protoc-gen-graphql` tool generates GraphQL schema files from Protocol Buffer service definitions. It maps:

- Proto services → GraphQL operations
- Proto messages → GraphQL types
- Federation metadata → GraphQL federation directives

#### Custom Templates
You can use a custom template file with the `protoc-gen-graphql` generator by configuring the `template_path` option in your `buf.gen.yaml` file:

```yaml
- local: protoc-gen-graphql
  out: ../gen/graphql
  opt:
    - paths=source_relative
    - template_path=/path/to/your/custom/template.tmpl
```

If the specified template file is not found, the generator will fall back to using the embedded default template.

## Key Features

- **Auto-reloading**: Services use `gow` for hot-reloading during development
- **Federation metadata**: Extended protocol buffer definitions for federation
- **Custom code generator**: Creates GraphQL schemas from Proto definitions
- **Connect protocol**: Modern, lightweight gRPC alternative

## Project Structure

- `/proto`: Protocol buffer definitions 
- `/gen`: Generated code
- `/services`: Individual service implementations
- `/tools`: Code generation tools

For more details, see the [CLAUDE.md](CLAUDE.md) development guide.
