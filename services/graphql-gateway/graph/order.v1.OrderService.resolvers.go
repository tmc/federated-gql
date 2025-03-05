package graph

import (
	"context"

	"connectrpc.com/connect"
	orderv1 "github.com/fraser-isbester/federated-gql/gen/go/order/v1"
	"github.com/fraser-isbester/federated-gql/services/graphql-gateway/graph/model"
)

// Order is the resolver for the order field.
func (r *queryResolver) Order(ctx context.Context, id string) (*model.Order, error) {
	// Call the order service
	resp, err := r.orderClient.GetOrder(ctx, connect.NewRequest(&orderv1.GetOrderRequest{
		OrderId: id,
	}))
	if err != nil {
		return nil, err
	}

	// Map the proto response to our GraphQL model
	var status model.OrderStatus
	switch resp.Msg.Status {
	case orderv1.OrderStatus_ORDER_STATUS_PENDING:
		status = model.OrderStatusPending
	case orderv1.OrderStatus_ORDER_STATUS_PROCESSING:
		status = model.OrderStatusProcessing
	case orderv1.OrderStatus_ORDER_STATUS_SHIPPED:
		status = model.OrderStatusShipped
	case orderv1.OrderStatus_ORDER_STATUS_DELIVERED:
		status = model.OrderStatusDelivered
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		status = model.OrderStatusCancelled
	default:
		status = model.OrderStatusPending
	}

	return &model.Order{
		ID:          resp.Msg.OrderId,
		CustomerId:  resp.Msg.CustomerId,
		TotalAmount: resp.Msg.TotalAmount,
		Status:      status,
		CreatedAt:   resp.Msg.CreatedAt,
	}, nil
}