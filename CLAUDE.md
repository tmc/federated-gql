# CLAUDE.md - Development Guidelines

## Commands
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
  - Users service: `make run-users`
  - Products service: `make run-products`
  - GraphQL gateway: `make run-graphql-gateway`

## Required Commands After Changes
**IMPORTANT**: After every change, run the following commands before committing:
1. `make generate` - ensure all generated code is up to date
2. `make build` - verify all services build successfully
3. `make test` - ensure all tests pass

## Code Style
- **Error handling**: Check errors with `if err != nil`, no wrapping, use `connect.NewError` for Connect errors
- **Imports**: Standard library first, then third-party packages, use block style with blank lines between groups
- **Naming**: CamelCase (Go standard), descriptive function names, exported functions capitalized
- **Types**: Use pointers for nullable values, helper functions for pointer conversions
- **Formatting**: Follow standard Go formatting (gofmt)
- **Documentation**: Comment all exported functions and types with meaningful descriptions

## Project Structure
- `/proto`: Protocol buffer definitions
- `/gen`: Generated code from protobuf
- `/services`: Individual service implementations
- `/tools`: Code generation tools