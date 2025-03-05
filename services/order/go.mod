module github.com/fraser-isbester/federated-gql/services/order

go 1.24.0

replace github.com/fraser-isbester/federated-gql/gen/go => ../../gen/go

require (
	connectrpc.com/connect v1.18.1
	github.com/fraser-isbester/federated-gql/gen/go v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.23.0
)

require (
	golang.org/x/text v0.14.0 // indirect
	google.golang.org/protobuf v1.36.5 // indirect
)