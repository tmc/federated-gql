package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	productv1 "github.com/fraser-isbester/federated-gql/gen/go/product/v1"
	"github.com/fraser-isbester/federated-gql/gen/go/product/v1/productv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type productServer struct{}

func (s *productServer) GetProduct(ctx context.Context, req *connect.Request[productv1.GetProductRequest]) (*connect.Response[productv1.GetProductResponse], error) {
	// Validate input
	if req.Msg.ProductId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("product_id is required"))
	}

	// Simulate product retrieval with some predefined products
	var product *productv1.Product
	switch req.Msg.ProductId {
	case "laptop":
		product = &productv1.Product{
			ProductId: "laptop",
			Name:      "High-Performance Laptop",
			Price:     1299.99,
		}
	case "smartphone":
		product = &productv1.Product{
			ProductId: "smartphone",
			Name:      "Advanced Smartphone",
			Price:     799.99,
		}
	default:
		product = &productv1.Product{
			ProductId: req.Msg.ProductId,
			Name:      fmt.Sprintf("Product %s", req.Msg.ProductId),
			Price:     99.99,
		}
	}

	fmt.Printf("GetProduct: returning product: %v\n", product.Name)

	return connect.NewResponse(&productv1.GetProductResponse{
		Product: product,
	}), nil
}

func main() {
	server := &productServer{}
	mux := http.NewServeMux()
	path, handler := productv1connect.NewProductServiceHandler(server)
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Changed default port to 8081
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting product service on %s", addr)

	err := http.ListenAndServe(
		addr,
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
