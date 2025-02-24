package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	userv1 "github.com/fraser-isbester/federated-gql/gen/go/user/v1"
	"github.com/fraser-isbester/federated-gql/gen/go/user/v1/userv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

// userServer implements the UserService interface
type userServer struct{}

// GetUser implements the UserService GetUser RPC method
func (s *userServer) GetUser(ctx context.Context, req *connect.Request[userv1.GetUserRequest]) (*connect.Response[userv1.GetUserResponse], error) {
	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id is required"))
	}

	return connect.NewResponse(&userv1.GetUserResponse{
		UserId: req.Msg.UserId,
		Name:   fmt.Sprintf("User %s", req.Msg.UserId),
	}), nil
}

func main() {
	server := &userServer{}
	mux := http.NewServeMux()
	path, handler := userv1connect.NewUserServiceHandler(server)
	mux.Handle(path, handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Starting user service on %s", addr)

	err := http.ListenAndServe(
		addr,
		h2c.NewHandler(mux, &http2.Server{}),
	)
	if err != nil {
		log.Fatal(err)
	}
}
