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
	// Validate input
	if req.Msg.UserId == "" {
		return nil, connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("user_id is required"))
	}

	// Simulate user retrieval with some predefined users
	var user *userv1.User
	switch req.Msg.UserId {
	case "alice":
		user = &userv1.User{
			UserId: "alice",
			Name:   "Alice Johnson",
		}
	case "bob":
		user = &userv1.User{
			UserId: "bob",
			Name:   "Bob Smith",
		}
	default:
		user = &userv1.User{
			UserId: req.Msg.UserId,
			Name:   fmt.Sprintf("User %s", req.Msg.UserId),
		}
	}

	return connect.NewResponse(&userv1.GetUserResponse{
		User: user,
	}), nil
}

func main() {
	server := &userServer{}
	mux := http.NewServeMux()
	path, handler := userv1connect.NewUserServiceHandler(server)
	mux.Handle(path, handler)

	// Add health check endpoint
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Changed default port to 8082
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
