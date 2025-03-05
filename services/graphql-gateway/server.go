package main

import (
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	"github.com/fraser-isbester/federated-gql/gen/go/order/v1/orderv1connect"
	productv1connect "github.com/fraser-isbester/federated-gql/gen/go/product/v1/productv1connect"
	userv1connect "github.com/fraser-isbester/federated-gql/gen/go/user/v1/userv1connect"
	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph"
	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Initialize service clients
	productClient := productv1connect.NewProductServiceClient(
		http.DefaultClient,
		"http://localhost:8081",
	)

	userClient := userv1connect.NewUserServiceClient(
		http.DefaultClient,
		"http://localhost:8082",
	)

	// Initialize the order service client
	orderClient := orderv1connect.NewOrderServiceClient(
		http.DefaultClient,
		"http://localhost:8083",
	)

	// Create resolver with RPC clients
	resolver := graph.NewResolver(productClient, userClient, orderClient)

	// Create executable schema
	srv := handler.New(graph.NewExecutableSchema(graph.Config{
		Resolvers: resolver,
	}))

	// Add supported transports
	srv.AddTransport(transport.POST{})
	srv.AddTransport(transport.Options{})
	srv.AddTransport(transport.Websocket{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	})

	// Setup routing with Chi
	router := chi.NewRouter()
	router.Handle("/", playground.Handler("GraphQL playground", "/query"))
	router.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}