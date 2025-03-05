package graph

//go:generate go run github.com/99designs/gqlgen generate

import (
	"github.com/fraser-isbester/federated-gql/gen/go/order/v1/orderv1connect"
	productv1connect "github.com/fraser-isbester/federated-gql/gen/go/product/v1/productv1connect"
	userv1connect "github.com/fraser-isbester/federated-gql/gen/go/user/v1/userv1connect"
	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

// Resolver is the root resolver that stores dependencies
type Resolver struct {
	productClient productv1connect.ProductServiceClient
	userClient    userv1connect.UserServiceClient
	orderClient   orderv1connect.OrderServiceClient
}

// NewResolver creates a new root resolver with the given clients
func NewResolver(pc productv1connect.ProductServiceClient, uc userv1connect.UserServiceClient, oc orderv1connect.OrderServiceClient) *Resolver {
	return &Resolver{
		productClient: pc,
		userClient:    uc,
		orderClient:   oc,
	}
}