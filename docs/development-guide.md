# Development Guide

This guide serves as a quick reference for developers working on the Federated GraphQL project.

## Quick Reference

### Commands
- Build/generate: `make generate` (runs both proto and GraphQL generation)
- Generate specific parts:
  - Proto only: `make generate-proto` (runs `buf generate` in proto directory)
  - GraphQL only: `make generate-gql` (cleans generated files and runs gqlgen)
  - Clean GraphQL: `make clean-gql` (removes generated GraphQL files)
- Build all services: `make build` (builds all services to the bin directory)
- Test: `make test` (runs tests for all services)
- Test specific services:
  - Users service: `make test-users`
  - Product service: `make test-product`
  - GraphQL gateway: `make test-graphql`
- Test single package: `go test ./path/to/package`
- Run services:
  - Users service: `make run-users` (uses gow for auto-reloading)
  - Products service: `make run-products` (uses gow for auto-reloading)
  - GraphQL gateway: `make run-graphql-gateway` (uses gow for auto-reloading)

### Development Workflow
**IMPORTANT**: After every change, run the following commands before committing:
1. `make generate` - ensure all generated code is up to date
2. `make build` - verify all services build successfully
3. `make test` - ensure all tests pass

## Code Style & Project Structure

### Code Style
- **Error handling**: Check errors with `if err != nil`, no wrapping, use `connect.NewError` for Connect errors
- **Imports**: Standard library first, then third-party packages, use block style with blank lines between groups
- **Naming**: CamelCase (Go standard), descriptive function names, exported functions capitalized
- **Types**: Use pointers for nullable values, helper functions for pointer conversions
- **Formatting**: Follow standard Go formatting (gofmt)
- **Documentation**: Comment all exported functions and types with meaningful descriptions

### Project Structure
- `/proto`: Protocol buffer definitions 
  - `/metadata`: Federation metadata options
  - `/user`, `/product`, etc.: Domain-specific service definitions
- `/gen`: Generated code from protobuf
  - `/go`: Generated Go code
  - `/graphql`: Generated GraphQL schemas
- `/services`: Individual service implementations
  - `/graphql-gateway`: Federation gateway service
  - `/users`, `/product`, etc.: Domain-specific services
- `/tools`: Code generation tools
  - `/protoc-gen-graphql`: Custom protoc plugin for GraphQL schema generation
- `/docs`: Comprehensive documentation

### Development Tips
- **Use auto-reloading**: Services use `gow` for hot-reloading during development
- **Consistent federation metadata**: Use the metadata extensions consistently to ensure proper federation
- **Test each layer**: Write tests for proto validation, service implementation, and GraphQL resolvers

## Detailed Documentation

For more detailed documentation, refer to these guides:

- [Getting Started](./getting-started.md): Setting up and running the project
- [Architecture](./architecture.md): Overview of system architecture
- [Adding Services](./adding-services.md): Step-by-step guide to adding new services
- [Code Generation](./code-generation.md): Details on the code generation lifecycle
- [Federation](./federation.md): Federation concepts and implementation
- [Testing](./testing.md): Best practices for testing
- [Deployment](./deployment.md): Deploying to production environments
- [Performance](./performance.md): Performance optimization techniques