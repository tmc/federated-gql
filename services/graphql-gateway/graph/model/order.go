package model

// OrderStatus represents the current status of an order
type OrderStatus string

// Order status constants
const (
	OrderStatusPending    OrderStatus = "PENDING"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusShipped    OrderStatus = "SHIPPED"
	OrderStatusDelivered  OrderStatus = "DELIVERED"
	OrderStatusCancelled  OrderStatus = "CANCELLED"
)

// Order represents a customer order
type Order struct {
	ID          string      `json:"id"`
	CustomerId  string      `json:"customerId"`
	TotalAmount float64     `json:"totalAmount"`
	Status      OrderStatus `json:"status"`
	CreatedAt   string      `json:"createdAt"`
}