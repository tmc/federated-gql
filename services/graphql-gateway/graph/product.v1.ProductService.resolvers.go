package graph

import (
	"context"

	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph/model"
)

// Product is the resolver for the product field.
func (r *queryResolver) Product(ctx context.Context, productId string) (*model.Product, error) {
	// Dummy data
	return &model.Product{
		ProductID: productId,
		Name:      func() *string { s := "Dummy Product"; return &s }(),
		Price:     func() *float64 { f := 99.99; return &f }(),
	}, nil
}
