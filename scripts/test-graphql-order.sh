#!/bin/bash

# Start the order service in the background
echo "Starting order service..."
make run-order &
ORDER_PID=$!

# Give it a moment to start
sleep 2

# Start GraphQL gateway in the background
echo "Starting GraphQL gateway..."
make run-graphql-gateway &
GRAPHQL_PID=$!

# Give it a moment to start
sleep 2

# Test with GraphQL query
echo "Testing GraphQL query for order..."
curl -s -X POST \
  -H "Content-Type: application/json" \
  -d '{"query": "{ order(id: \"123\") { id customerId totalAmount status createdAt } }"}' \
  http://localhost:8080/query

# Kill services
kill $ORDER_PID
kill $GRAPHQL_PID

echo -e "\n\nGraphQL order query test completed."