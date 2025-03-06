package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/99designs/gqlgen/graphql/playground"
	productv1connect "github.com/fraser-isbester/federated-gql/gen/go/product/v1/productv1connect"
	userv1connect "github.com/fraser-isbester/federated-gql/gen/go/user/v1/userv1connect"
	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/websocket"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create Connect RPC clients (not used yet but initialized for future steps)
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

	// Add middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(60 * time.Second))

	// Custom middleware to log GraphQL operations
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/graphql" && r.Method == "POST" {
				log.Printf("GraphQL operation received from %s", r.RemoteAddr)
			}
			next.ServeHTTP(w, r)
		})
	})

	// Use Apollo Sandbox as the default interface
	router.Get("/", http.HandlerFunc(RenderApolloSandbox))
	
	// Keep the GraphQL playground as an alternative
	router.Handle("/playground", playground.Handler("GraphQL playground", "/graphql"))
	
	// The GraphQL endpoint
	router.Handle("/graphql", srv)

	log.Printf("Connect to http://localhost:%s/ for Apollo Sandbox", port)
	log.Printf("Connect to http://localhost:%s/playground for GraphQL Playground", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}