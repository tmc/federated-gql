package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	orderv1 "github.com/fraser-isbester/federated-gql/gen/go/order/v1"
	"github.com/fraser-isbester/federated-gql/gen/go/order/v1/orderv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// orderServer implements the OrderService interface
type orderServer struct{}

// GetOrder implements the OrderService GetOrder RPC method
func (s *orderServer) GetOrder(ctx context.Context, req *connect.Request[orderv1.GetOrderRequest]) (*connect.Response[orderv1.GetOrderResponse], error) {
	if req.Msg.OrderId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("order_id is required"))
	}

	// For demonstration purposes, we'll generate a mock order
	// In a real application, this would query a database
	return connect.NewResponse(&orderv1.GetOrderResponse{
		OrderId:     req.Msg.OrderId,
		CustomerId:  fmt.Sprintf("cust_%s", req.Msg.OrderId),
		TotalAmount: 149.99,
		Status:      orderv1.OrderStatus_ORDER_STATUS_PROCESSING,
		CreatedAt:   time.Now().Format(time.RFC3339),
	}), nil
}

func main() {
	server := &orderServer{}
	mux := http.NewServeMux()
	path, handler := orderv1connect.NewOrderServiceHandler(server)
	mux.Handle(path, handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Using 8082 since users is 8080 and products is likely 8081
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting order service on %s", addr)

	err := http.ListenAndServe(
		addr,
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Fatal(err)
	}
}