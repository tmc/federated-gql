package graph

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	productv1 "github.com/fraser-isbester/federated-gql/gen/go/product/v1"
	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph/model"
)

// Product is the resolver for the product field.
func (r *queryResolver) Product(ctx context.Context, productID string) (*model.Product, error) {
	// Construct the Connect Request
	input := &productv1.GetProductRequest{ProductId: productID}

	// Make the RPC call using the Connect client
	resp, err := r.productClient.GetProduct(ctx, connect.NewRequest(input))
	if err != nil {
		fmt.Printf("Error fetching product: %v\n", err)
		return nil, err
	}

	fmt.Println("product.v1.ProductService.resolvers: ", resp.Msg.Product.Name)

	// Map the protobuf response to GraphQL model
	if resp.Msg.Product != nil {
		return &model.Product{
			ProductID: resp.Msg.Product.ProductId, // Direct assignment
			Name:      strPtr(resp.Msg.Product.Name),
			Price:     floatPtr(resp.Msg.Product.Price),
		}, nil
	}

	// Return nil if product not found
	return nil, nil
}
