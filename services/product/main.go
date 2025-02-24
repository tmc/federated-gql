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
	if req.Msg.ProductId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("product_id is required"))
	}

	return connect.NewResponse(&productv1.GetProductResponse{
		ProductId: req.Msg.ProductId,
		Name:      fmt.Sprintf("Product %s", req.Msg.ProductId),
		Price:     99.99,
	}), nil
}

func main() {
	server := &productServer{}
	mux := http.NewServeMux()
	path, handler := productv1connect.NewProductServiceHandler(server)
	mux.Handle(path, handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
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
