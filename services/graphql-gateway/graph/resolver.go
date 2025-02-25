package graph

import (
	productv1connect "github.com/fraser-isbester/federated-gql/gen/go/product/v1/productv1connect"
	userv1connect "github.com/fraser-isbester/federated-gql/gen/go/user/v1/userv1connect"
)

// Resolver is the root resolver that stores dependencies
type Resolver struct {
	productClient productv1connect.ProductServiceClient
	userClient    userv1connect.UserServiceClient
}

// NewResolver creates a new root resolver with the given clients
func NewResolver(pc productv1connect.ProductServiceClient, uc userv1connect.UserServiceClient) *Resolver {
	return &Resolver{
		productClient: pc,
		userClient:    uc,
	}
}

// Query returns the root query resolver
func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

// Centralized queryResolver to avoid redeclaration
type queryResolver struct{ *Resolver }
