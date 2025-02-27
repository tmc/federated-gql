package graph

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	productv1 "github.com/fraser-isbester/federated-gql/gen/go/product/v1"
	"github.com/fraser-isbester/federated-gql/gen/go/product/v1/productv1connect"
	userv1 "github.com/fraser-isbester/federated-gql/gen/go/user/v1"
	"github.com/fraser-isbester/federated-gql/gen/go/user/v1/userv1connect"
)

// Mock Product Service Client
type mockProductServiceClient struct {
	productv1connect.ProductServiceClient
	mockProduct *productv1.Product
	mockError   error
}

func (m *mockProductServiceClient) GetProduct(
	ctx context.Context,
	req *connect.Request[productv1.GetProductRequest],
) (*connect.Response[productv1.GetProductResponse], error) {
	if m.mockError != nil {
		return nil, m.mockError
	}

	return connect.NewResponse(&productv1.GetProductResponse{
		Product: m.mockProduct,
	}), nil
}

// Mock User Service Client
type mockUserServiceClient struct {
	userv1connect.UserServiceClient
	mockUser  *userv1.User
	mockError error
}

func (m *mockUserServiceClient) GetUser(
	ctx context.Context,
	req *connect.Request[userv1.GetUserRequest],
) (*connect.Response[userv1.GetUserResponse], error) {
	if m.mockError != nil {
		return nil, m.mockError
	}

	return connect.NewResponse(&userv1.GetUserResponse{
		User: m.mockUser,
	}), nil
}

func TestProductResolver(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		productID     string
		mockProduct   *productv1.Product
		mockError     error
		expectedName  string
		expectedPrice float64
		expectError   bool
	}{
		{
			name:      "Successful product retrieval",
			productID: "laptop",
			mockProduct: &productv1.Product{
				ProductId: "laptop",
				Name:      "High-Performance Laptop",
				Price:     1299.99,
			},
			mockError:     nil,
			expectedName:  "High-Performance Laptop",
			expectedPrice: 1299.99,
			expectError:   false,
		},
		{
			name:          "Error from product service",
			productID:     "error",
			mockProduct:   nil,
			mockError:     errors.New("service error"),
			expectedName:  "",
			expectedPrice: 0,
			expectError:   true,
		},
		{
			name:      "Product with minimal data",
			productID: "minimal",
			mockProduct: &productv1.Product{
				ProductId: "minimal",
				Name:      "",
				Price:     0,
			},
			mockError:     nil,
			expectedName:  "",
			expectedPrice: 0,
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock clients
			mockProductClient := &mockProductServiceClient{
				mockProduct: tt.mockProduct,
				mockError:   tt.mockError,
			}
			mockUserClient := &mockUserServiceClient{}

			// Create resolver
			resolver := NewResolver(mockProductClient, mockUserClient)
			queryResolver := resolver.Query()

			// Execute the resolver
			product, err := queryResolver.Product(ctx, tt.productID)

			// Check error
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check product
			if product == nil {
				t.Fatalf("Expected product but got nil")
			}

			if product.ProductID != tt.productID {
				t.Errorf("Expected productID %s, got %s", tt.productID, product.ProductID)
			}

			if product.Name != nil && *product.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, *product.Name)
			}

			if product.Price != nil && *product.Price != tt.expectedPrice {
				t.Errorf("Expected price %.2f, got %.2f", tt.expectedPrice, *product.Price)
			}
		})
	}
}

func TestUserResolver(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		userID       string
		mockUser     *userv1.User
		mockError    error
		expectedName string
		expectError  bool
	}{
		{
			name:   "Successful user retrieval",
			userID: "alice",
			mockUser: &userv1.User{
				UserId: "alice",
				Name:   "Alice Johnson",
			},
			mockError:    nil,
			expectedName: "Alice Johnson",
			expectError:  false,
		},
		{
			name:         "Error from user service",
			userID:       "error",
			mockUser:     nil,
			mockError:    errors.New("service error"),
			expectedName: "",
			expectError:  true,
		},
		{
			name:   "User with minimal data",
			userID: "minimal",
			mockUser: &userv1.User{
				UserId: "minimal",
				Name:   "",
			},
			mockError:    nil,
			expectedName: "",
			expectError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock clients
			mockProductClient := &mockProductServiceClient{}
			mockUserClient := &mockUserServiceClient{
				mockUser:  tt.mockUser,
				mockError: tt.mockError,
			}

			// Create resolver
			resolver := NewResolver(mockProductClient, mockUserClient)
			queryResolver := resolver.Query()

			// Execute the resolver
			user, err := queryResolver.User(ctx, tt.userID)

			// Check error
			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			// Check user
			if user == nil {
				t.Fatalf("Expected user but got nil")
			}

			if user.UserID != tt.userID {
				t.Errorf("Expected userID %s, got %s", tt.userID, user.UserID)
			}

			if user.Name != nil && *user.Name != tt.expectedName {
				t.Errorf("Expected name %s, got %s", tt.expectedName, *user.Name)
			}
		})
	}
}
