package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
	productv1connect "github.com/fraser-isbester/federated-gql/gen/go/product/v1/productv1connect"
	userv1connect "github.com/fraser-isbester/federated-gql/gen/go/user/v1/userv1connect"
	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create Connect RPC clients
	productClient := productv1connect.NewProductServiceClient(
		http.DefaultClient,
		"http://localhost:8081",
	)

	userClient := userv1connect.NewUserServiceClient(
		http.DefaultClient,
		"http://localhost:8082",
	)

	// Create resolver with RPC clients
	resolver := graph.NewResolver(productClient, userClient)
	fmt.Println(resolver)

	// Create a placeholder for the executable schema
	// This will be generated after gqlgen generation
	srv := handler.NewDefaultServer(nil)

	// Add playground handler
	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
