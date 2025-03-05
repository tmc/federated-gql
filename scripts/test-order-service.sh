#!/bin/bash

# Start the order service in the background
echo "Starting order service..."
make run-order &
ORDER_PID=$!

# Give it a moment to start
sleep 2

# Test with direct HTTP request to the order service
echo "Testing order service directly..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"order_id": "123"}' \
  http://localhost:8082/order.v1.OrderService/GetOrder

# Kill the order service
kill $ORDER_PID

echo -e "\n\nOrder service test completed."