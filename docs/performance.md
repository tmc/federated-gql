# Performance Optimization

This document provides guidance on optimizing the performance of your federated GraphQL architecture.

## Performance Considerations

### GraphQL Query Performance

GraphQL's flexibility can lead to performance challenges. Here are key optimization areas:

#### 1. Query Complexity Analysis

Implement query complexity analysis to prevent expensive queries:

```go
func ComplexityLimit(h http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Parse the GraphQL query
        var params struct {
            Query string `json:"query"`
        }
        if err := json.NewDecoder(r.Body).Decode(&params); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }
        r.Body.Close()
        
        // Calculate complexity
        complexity, err := calculateQueryComplexity(params.Query)
        if err != nil {
            http.Error(w, "Invalid query", http.StatusBadRequest)
            return
        }
        
        // Reject overly complex queries
        if complexity > 100 {
            http.Error(w, "Query too complex", http.StatusTooManyRequests)
            return
        }
        
        // Create a new body with the original query
        var buf bytes.Buffer
        if err := json.NewEncoder(&buf).Encode(params); err != nil {
            http.Error(w, "Internal error", http.StatusInternalServerError)
            return
        }
        r.Body = io.NopCloser(&buf)
        
        h.ServeHTTP(w, r)
    })
}
```

#### 2. Query Depth Limiting

Limit the depth of GraphQL queries:

```go
func calculateQueryDepth(query string) (int, error) {
    // Parse the query
    parsedQuery, err := parser.ParseQuery(&ast.Source{
        Body: []byte(query),
    })
    if err != nil {
        return 0, err
    }
    
    // Calculate max depth
    maxDepth := 0
    for _, op := range parsedQuery.Operations {
        for _, sel := range op.SelectionSet {
            depth := calculateSelectionDepth(sel, 1)
            if depth > maxDepth {
                maxDepth = depth
            }
        }
    }
    
    return maxDepth, nil
}

func calculateSelectionDepth(selection ast.Selection, currentDepth int) int {
    switch sel := selection.(type) {
    case *ast.Field:
        if sel.SelectionSet == nil {
            return currentDepth
        }
        
        maxDepth := currentDepth
        for _, subSel := range sel.SelectionSet {
            depth := calculateSelectionDepth(subSel, currentDepth+1)
            if depth > maxDepth {
                maxDepth = depth
            }
        }
        return maxDepth
        
    // Handle other selection types...
    }
    
    return currentDepth
}
```

#### 3. Field Selection Limiting

Restrict the number of fields that can be requested:

```go
func countFields(query string) (int, error) {
    // Parse the query
    parsedQuery, err := parser.ParseQuery(&ast.Source{
        Body: []byte(query),
    })
    if err != nil {
        return 0, err
    }
    
    // Count fields
    count := 0
    for _, op := range parsedQuery.Operations {
        for _, sel := range op.SelectionSet {
            count += countSelectionFields(sel)
        }
    }
    
    return count, nil
}

func countSelectionFields(selection ast.Selection) int {
    switch sel := selection.(type) {
    case *ast.Field:
        count := 1
        if sel.SelectionSet != nil {
            for _, subSel := range sel.SelectionSet {
                count += countSelectionFields(subSel)
            }
        }
        return count
        
    // Handle other selection types...
    }
    
    return 0
}
```

### Federation Performance

#### 1. Optimize Entity Resolution

Minimize entity resolution across services:

```go
// BAD: Resolving entities individually
for _, id := range productIDs {
    product, err := r.productClient.GetProduct(ctx, connect.NewRequest(&productv1.GetProductRequest{
        ProductId: id,
    }))
    // Process product...
}

// GOOD: Batch entity resolution
products, err := r.productClient.GetProducts(ctx, connect.NewRequest(&productv1.GetProductsRequest{
    ProductIds: productIDs,
}))
```

#### 2. Implement Dataloader Pattern

Use DataLoader for batching and caching:

```go
type Loaders struct {
    UserLoader     *dataloader.Loader
    ProductLoader  *dataloader.Loader
}

func NewLoaders(userClient userv1connect.UserServiceClient, productClient productv1connect.ProductServiceClient) *Loaders {
    return &Loaders{
        UserLoader: dataloader.NewBatchedLoader(func(ctx context.Context, keys dataloader.Keys) []*dataloader.Result {
            userIDs := make([]string, len(keys))
            for i, key := range keys {
                userIDs[i] = key.String()
            }
            
            // Batch request to user service
            resp, err := userClient.GetUsers(ctx, connect.NewRequest(&userv1.GetUsersRequest{
                UserIds: userIDs,
            }))
            
            if err != nil {
                return createErrorResults(keys, err)
            }
            
            // Map results by ID for fast lookup
            userMap := make(map[string]*userv1.User)
            for _, user := range resp.Msg.Users {
                userMap[user.UserId] = user
            }
            
            // Create results in the same order as keys
            results := make([]*dataloader.Result, len(keys))
            for i, key := range keys {
                user, ok := userMap[key.String()]
                if !ok {
                    results[i] = &dataloader.Result{Error: fmt.Errorf("user not found")}
                } else {
                    results[i] = &dataloader.Result{Data: user}
                }
            }
            
            return results
        }),
        
        // Similarly implement ProductLoader
    }
}

// Use in middleware to add loaders to context
func DataloaderMiddleware(loaders *Loaders) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Add loaders to request context
            ctx := context.WithValue(r.Context(), loadersKey, loaders)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Use in resolvers
func (r *queryResolver) User(ctx context.Context, userID string) (*model.User, error) {
    // Get loaders from context
    loaders := ctx.Value(loadersKey).(*Loaders)
    
    // Use loader to get user
    userThunk := loaders.UserLoader.Load(ctx, dataloader.StringKey(userID))
    user, err := userThunk()
    if err != nil {
        return nil, err
    }
    
    // Convert to model
    return &model.User{
        UserID: user.(*userv1.User).UserId,
        Name:   strPtr(user.(*userv1.User).Name),
    }, nil
}
```

#### 3. Optimize Resolver Chain

Analyze and optimize resolver execution chains:

```go
// Example tracing middleware to identify slow resolvers
func TracingMiddleware() graphql.FieldMiddleware {
    return func(ctx context.Context, next graphql.Resolver) (interface{}, error) {
        path, _ := graphql.GetFieldContext(ctx)
        fieldName := path.Field.Name
        typeName := path.Field.ObjectDefinition.Name
        
        start := time.Now()
        res, err := next(ctx)
        duration := time.Since(start)
        
        if duration > 100*time.Millisecond {
            log.Printf("Slow resolver: %s.%s took %s", typeName, fieldName, duration)
        }
        
        return res, err
    }
}
```

### Network Optimization

#### 1. Connection Pooling

Optimize HTTP client settings:

```go
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 20,
        MaxConnsPerHost:     100,
        IdleConnTimeout:     90 * time.Second,
        DisableCompression:  false,
        ForceAttemptHTTP2:   true,
    },
    Timeout: 10 * time.Second,
}

// Use for all service clients
userClient := userv1connect.NewUserServiceClient(httpClient, "http://users:8082")
productClient := productv1connect.NewProductServiceClient(httpClient, "http://products:8081")
```

#### 2. Request Compression

Enable compression for requests and responses:

```go
// Enable compression in Connect client
client := userv1connect.NewUserServiceClient(
    httpClient, 
    "http://users:8082",
    connect.WithAcceptCompression(
        connect.GzipEncoding{}, 
        connect.IdentityEncoding{},
    ),
    connect.WithSendCompression(connect.GzipEncoding{}),
)

// Enable compression in server
userHandler := userv1connect.NewUserServiceHandler(
    &userServer{},
    connect.WithCompression(
        connect.GzipEncoding{},
        connect.IdentityEncoding{},
    ),
)
```

### Cache Optimization

#### 1. Response Caching

Implement response caching at the gateway level:

```go
type cacheKey struct {
    query     string
    variables string
}

func (k cacheKey) String() string {
    return fmt.Sprintf("%s:%s", k.query, k.variables)
}

var responseCache = cache.New(5*time.Minute, 10*time.Minute)

func CachingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "POST" {
            next.ServeHTTP(w, r)
            return
        }
        
        // Read request body
        var body bytes.Buffer
        tee := io.TeeReader(r.Body, &body)
        
        var params struct {
            Query     string                 `json:"query"`
            Variables map[string]interface{} `json:"variables"`
        }
        
        if err := json.NewDecoder(tee).Decode(&params); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }
        
        // Skip mutations
        if strings.HasPrefix(strings.TrimSpace(params.Query), "mutation") {
            r.Body = io.NopCloser(&body)
            next.ServeHTTP(w, r)
            return
        }
        
        // Create cache key
        varsJson, _ := json.Marshal(params.Variables)
        key := cacheKey{
            query:     params.Query,
            variables: string(varsJson),
        }
        
        // Check cache
        if cachedResp, found := responseCache.Get(key.String()); found {
            w.Header().Set("Content-Type", "application/json")
            w.Header().Set("X-Cache", "HIT")
            w.Write(cachedResp.([]byte))
            return
        }
        
        // Create response recorder
        rec := httptest.NewRecorder()
        r.Body = io.NopCloser(&body)
        
        // Execute request
        next.ServeHTTP(rec, r)
        
        // Cache the response if successful
        if rec.Code == http.StatusOK {
            responseCache.Set(key.String(), rec.Body.Bytes(), cache.DefaultExpiration)
        }
        
        // Copy response to original writer
        for k, v := range rec.Header() {
            w.Header()[k] = v
        }
        w.Header().Set("X-Cache", "MISS")
        w.WriteHeader(rec.Code)
        w.Write(rec.Body.Bytes())
    })
}
```

#### 2. Cache Directives

Use cache directives in GraphQL schemas:

```graphql
type Query {
  user(userID: String!): User @cacheControl(maxAge: 60)
  products: [Product!]! @cacheControl(maxAge: 300)
}

type User @cacheControl(maxAge: 60) {
  userID: String!
  name: String
  # Dynamic data with shorter cache time
  status: String @cacheControl(maxAge: 10)
}
```

### Database Optimization

#### 1. Connection Pooling

Optimize database connection pools:

```go
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

// Configure the pool based on service needs
db.SetMaxOpenConns(25) // Depends on database capacity
db.SetMaxIdleConns(25) // Keep connections ready
db.SetConnMaxLifetime(5 * time.Minute) // Recycle connections
```

#### 2. Query Optimization

Optimize database queries:

```go
// BAD: N+1 query problem
for _, userID := range userIDs {
    var user User
    db.QueryRow("SELECT * FROM users WHERE user_id = $1", userID).Scan(&user.ID, &user.Name)
    // Process user...
}

// GOOD: Single query with IN clause
rows, err := db.Query("SELECT * FROM users WHERE user_id = ANY($1)", pq.Array(userIDs))
```

### Memory Management

#### 1. Response Streaming

For large responses, use streaming:

```go
func streamingHandler(w http.ResponseWriter, r *http.Request) {
    // Set up a streaming response
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    w.Header().Set("Transfer-Encoding", "chunked")
    
    // Start the response
    fmt.Fprintf(w, "{\"data\":{\"items\":[")
    flusher.Flush()
    
    // Stream items
    for i := 0; i < 1000; i++ {
        if i > 0 {
            fmt.Fprintf(w, ",")
        }
        
        item := fmt.Sprintf("{\"id\":%d,\"value\":\"item-%d\"}", i, i)
        fmt.Fprintf(w, item)
        
        // Flush periodically
        if i%10 == 0 {
            flusher.Flush()
        }
    }
    
    // End the response
    fmt.Fprintf(w, "]}}")
    flusher.Flush()
}
```

#### 2. Memory Profiling

Implement memory profiling:

```go
import _ "net/http/pprof"

func main() {
    // Enable profiling endpoints
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // Rest of the service...
}
```

### Load Testing

Use k6 for load testing:

```js
// k6-script.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '1m', target: 50 },  // Ramp up to 50 users
    { duration: '3m', target: 50 },  // Stay at 50 users
    { duration: '1m', target: 100 }, // Ramp up to 100 users
    { duration: '3m', target: 100 }, // Stay at 100 users
    { duration: '1m', target: 0 },   // Ramp down to 0 users
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95% of requests should be below 500ms
  },
};

export default function() {
  const query = `
    query {
      user(userID: "alice") {
        userID
        name
        products {
          productID
          name
        }
      }
    }
  `;
  
  const res = http.post('http://localhost:8080/query', 
    JSON.stringify({ query: query }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });
  
  sleep(1);
}
```

## Performance Metrics to Monitor

1. **Response Time**: Average and percentile (p95, p99) response times
2. **Request Rate**: Requests per second
3. **Error Rate**: Percentage of failed requests
4. **CPU Usage**: Overall and per-service CPU utilization
5. **Memory Usage**: Heap and non-heap memory consumption
6. **GC Metrics**: Garbage collection frequency and duration
7. **Network I/O**: Bytes sent/received
8. **Database Metrics**: Query execution time, connection pool usage

## Tools for Performance Analysis

1. **pprof**: Go's built-in profiling tool
2. **Prometheus**: Metrics collection and alerting
3. **Grafana**: Metrics visualization
4. **Jaeger/Zipkin**: Distributed tracing
5. **k6**: Load testing
6. **New Relic/Datadog**: Application performance monitoring

## Next Steps

- [Deployment Guide](./deployment.md)
- [Testing Best Practices](./testing.md)
- [Architecture Overview](./architecture.md)