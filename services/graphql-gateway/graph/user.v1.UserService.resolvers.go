package graph

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	userv1 "github.com/fraser-isbester/federated-gql/gen/go/user/v1"
	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph/model"
)

// User is the resolver for the user field.
func (r *queryResolver) User(ctx context.Context, userId string) (*model.User, error) {

	// Construct the Connect Request
	input := &userv1.GetUserRequest{UserId: userId}

	// Make the RPC call using the Connect client
	resp, err := r.userClient.GetUser(ctx, connect.NewRequest(input))
	if err != nil {
		fmt.Printf("Error fetching user: %v\n", err)
		return nil, err
	}

	fmt.Println("user.v1.UserService.resolvers: ", resp.Msg.User.Name)

	// Map the protobuf response to GraphQL model
	if resp.Msg.User != nil {
		return &model.User{
			UserID: resp.Msg.User.UserId, // Direct assignment
			Name:   strPtr(resp.Msg.User.Name),
		}, nil
	}

	// Return nil if user not found
	return nil, nil
}
