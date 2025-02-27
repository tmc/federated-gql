package main

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	productv1 "github.com/fraser-isbester/federated-gql/gen/go/product/v1"
)

func TestGetProduct(t *testing.T) {
	server := &productServer{}
	ctx := context.Background()

	testCases := []struct {
		name           string
		productID      string
		expectedName   string
		expectedPrice  float64
		expectError    bool
		errorContains  string
	}{
		{
			name:          "Get predefined product laptop",
			productID:     "laptop",
			expectedName:  "High-Performance Laptop",
			expectedPrice: 1299.99,
			expectError:   false,
		},
		{
			name:          "Get predefined product smartphone",
			productID:     "smartphone",
			expectedName:  "Advanced Smartphone",
			expectedPrice: 799.99,
			expectError:   false,
		},
		{
			name:          "Get dynamically generated product",
			productID:     "headphones",
			expectedName:  "Product headphones",
			expectedPrice: 99.99,
			expectError:   false,
		},
		{
			name:          "Empty product ID returns error",
			productID:     "",
			expectError:   true,
			errorContains: "product_id is required",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := connect.NewRequest(&productv1.GetProductRequest{
				ProductId: tc.productID,
			})

			resp, err := server.GetProduct(ctx, req)

			if tc.expectError {
				if err == nil {
					t.Fatalf("expected error containing '%s', got no error", tc.errorContains)
				}
				if connectErr, ok := err.(*connect.Error); ok {
					if connectErr.Message() != tc.errorContains {
						t.Fatalf("expected error containing '%s', got '%s'", tc.errorContains, connectErr.Message())
					}
				} else {
					t.Fatalf("expected connect.Error, got %T", err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if resp.Msg.Product.Name != tc.expectedName {
				t.Errorf("expected product name '%s', got '%s'", tc.expectedName, resp.Msg.Product.Name)
			}

			if resp.Msg.Product.ProductId != tc.productID {
				t.Errorf("expected product ID '%s', got '%s'", tc.productID, resp.Msg.Product.ProductId)
			}

			if resp.Msg.Product.Price != tc.expectedPrice {
				t.Errorf("expected product price %.2f, got %.2f", tc.expectedPrice, resp.Msg.Product.Price)
			}
		})
	}
}