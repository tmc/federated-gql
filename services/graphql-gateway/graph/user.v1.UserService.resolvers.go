package graph

import (
	"context"

	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph/model"
)

func (r *Resolver) User(ctx context.Context, userId string) (*model.User, error) {
	return &model.User{
		UserID: userId,
		Name:   func() *string { s := "Dummy User"; return &s }(),
	}, nil
}
