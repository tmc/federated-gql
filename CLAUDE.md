# CLAUDE.md - Development Guidelines

## Commands
- Build/generate: `make generate` (runs `buf generate` for protocol buffers)
- Test: `make test` (runs `go test ./...`)
- Test single package: `go test ./path/to/package`
- Run services:
  - Users service: `make run-users`
  - Products service: `make run-products`
  - GraphQL gateway: `make run-graphql-gateway`

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

Always run tests before committing changes.