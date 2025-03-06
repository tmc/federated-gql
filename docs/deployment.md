# Deployment Guide

This document provides guidance on deploying the federated GraphQL architecture to various environments.

## Deployment Overview

The project consists of multiple services that should be deployed separately:

1. **GraphQL Gateway**: The federation layer that combines all services
2. **Domain Services**: Independent microservices (Users, Products, etc.)
3. **Supporting Infrastructure**: Databases, caches, etc.

## Preparing for Production

### 1. Configuration

Each service should be configurable through environment variables:

```go
// Example configuration in main.go
port := os.Getenv("PORT")
if port == "" {
    port = "8080" // Default
}

serviceHost := os.Getenv("USER_SERVICE_HOST")
if serviceHost == "" {
    serviceHost = "localhost:8082" // Default
}
```

Create a `.env.example` file for each service to document required variables.

### 2. Security

#### TLS Configuration

For production, enable TLS:

```go
// HTTP/2 with TLS
err := http.ListenAndServeTLS(
    addr,
    "/path/to/cert.pem",
    "/path/to/key.pem",
    handler,
)
```

#### Authentication

Add authentication middleware to the gateway:

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if !validateToken(token) {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

// Apply middleware
router.Use(authMiddleware)
```

### 3. Observability

#### Logging

Implement structured logging with a library like zap:

```go
logger, _ := zap.NewProduction()
defer logger.Sync()

logger.Info("Starting service",
    zap.String("service", "user-service"),
    zap.String("port", port),
)
```

#### Metrics

Add Prometheus metrics:

```go
// Initialize metrics
requestCounter := prometheus.NewCounterVec(
    prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total number of HTTP requests",
    },
    []string{"method", "endpoint", "status"},
)
prometheus.MustRegister(requestCounter)

// Expose metrics endpoint
http.Handle("/metrics", promhttp.Handler())
```

#### Tracing

Implement distributed tracing with OpenTelemetry:

```go
// Initialize tracer
tp := initTracer()
defer func() {
    if err := tp.Shutdown(context.Background()); err != nil {
        log.Printf("Error shutting down tracer provider: %v", err)
    }
}()

// Create a span for each request
func tracingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        ctx, span := otel.Tracer("http").Start(r.Context(), r.URL.Path)
        defer span.End()
        next.ServeHTTP(w, r.WithContext(ctx))
    })
}
```

### 4. Health Checks

Add health check endpoints to each service:

```go
// Liveness probe
mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("OK"))
})

// Readiness probe
mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
    // Check dependencies
    if !isDatabaseConnected() {
        w.WriteHeader(http.StatusServiceUnavailable)
        return
    }
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("Ready"))
})
```

## Containerization

### Docker Configuration

Create a Dockerfile for each service:

```dockerfile
# services/users/Dockerfile
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY . .
RUN go build -o bin/users ./services/users

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bin/users /app/users
EXPOSE 8082
CMD ["/app/users"]
```

### Docker Compose

For local development and testing:

```yaml
# docker-compose.yml
version: '3'
services:
  users:
    build:
      context: .
      dockerfile: services/users/Dockerfile
    ports:
      - "8082:8082"
    environment:
      - PORT=8082
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8082/healthz"]

  products:
    build:
      context: .
      dockerfile: services/product/Dockerfile
    ports:
      - "8081:8081"
    environment:
      - PORT=8081
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8081/healthz"]

  graphql-gateway:
    build:
      context: .
      dockerfile: services/graphql-gateway/Dockerfile
    ports:
      - "8080:8080"
    environment:
      - PORT=8080
      - USER_SERVICE_HOST=users:8082
      - PRODUCT_SERVICE_HOST=products:8081
    depends_on:
      - users
      - products
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/healthz"]
```

## Kubernetes Deployment

### Manifests

Create Kubernetes deployment manifests for each service:

```yaml
# k8s/users-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: users
spec:
  replicas: 2
  selector:
    matchLabels:
      app: users
  template:
    metadata:
      labels:
        app: users
    spec:
      containers:
      - name: users
        image: your-registry/federated-gql/users:latest
        ports:
        - containerPort: 8082
        env:
        - name: PORT
          value: "8082"
        livenessProbe:
          httpGet:
            path: /healthz
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8082
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          limits:
            cpu: "500m"
            memory: "512Mi"
          requests:
            cpu: "100m"
            memory: "128Mi"
```

### Service Definitions

Create service definitions for each component:

```yaml
# k8s/users-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: users
spec:
  selector:
    app: users
  ports:
  - port: 8082
    targetPort: 8082
  type: ClusterIP
```

### Gateway Ingress

Create an ingress for the GraphQL gateway:

```yaml
# k8s/gateway-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: graphql-gateway
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: api.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: graphql-gateway
            port:
              number: 8080
  tls:
  - hosts:
    - api.example.com
    secretName: api-tls-secret
```

## Continuous Deployment

### GitHub Actions Pipeline

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [ main ]
    tags: [ 'v*' ]

jobs:
  build-and-push:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Docker Buildx
      uses: docker/setup-buildx-action@v2
    
    - name: Login to Container Registry
      uses: docker/login-action@v2
      with:
        registry: your-registry.io
        username: ${{ secrets.REGISTRY_USERNAME }}
        password: ${{ secrets.REGISTRY_PASSWORD }}
    
    - name: Build and push Users service
      uses: docker/build-push-action@v4
      with:
        context: .
        file: services/users/Dockerfile
        push: true
        tags: your-registry.io/federated-gql/users:latest
    
    - name: Build and push Products service
      uses: docker/build-push-action@v4
      with:
        context: .
        file: services/product/Dockerfile
        push: true
        tags: your-registry.io/federated-gql/products:latest
    
    - name: Build and push GraphQL Gateway
      uses: docker/build-push-action@v4
      with:
        context: .
        file: services/graphql-gateway/Dockerfile
        push: true
        tags: your-registry.io/federated-gql/graphql-gateway:latest
    
    - name: Deploy to Kubernetes
      uses: steebchen/kubectl@v2
      with:
        config: ${{ secrets.KUBE_CONFIG_DATA }}
        command: apply -f k8s/
```

## Scaling Considerations

### Horizontal Scaling

Services can scale independently based on load:

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: users-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: users
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

### Connection Pooling

Implement connection pooling for service-to-service communication:

```go
// Create a client with connection pooling
httpClient := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 20,
        MaxConnsPerHost:     100,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 10 * time.Second,
}

// Use the client for Connect RPC
client := userv1connect.NewUserServiceClient(httpClient, "http://users-service:8082")
```

## Database Considerations

### Connection Management

Proper database connection management:

```go
// Create a connection pool
db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
if err != nil {
    log.Fatalf("Failed to connect to database: %v", err)
}

// Configure the pool
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(5 * time.Minute)

// Verify connection
if err := db.Ping(); err != nil {
    log.Fatalf("Failed to ping database: %v", err)
}
```

### Migrations

Use a migration tool like golang-migrate:

```go
// Apply migrations
m, err := migrate.New(
    "file://migrations",
    os.Getenv("DATABASE_URL"),
)
if err != nil {
    log.Fatalf("Migration failed to initialize: %v", err)
}

if err := m.Up(); err != nil && err != migrate.ErrNoChange {
    log.Fatalf("Migration failed: %v", err)
}
```

## Cloud Provider Specifics

### AWS

- Use EKS for Kubernetes
- Use ECR for container registry
- Use RDS for databases
- Use CloudWatch for monitoring

### Google Cloud

- Use GKE for Kubernetes
- Use Artifact Registry for containers
- Use Cloud SQL for databases
- Use Cloud Monitoring for observability

### Azure

- Use AKS for Kubernetes
- Use Container Registry
- Use Azure SQL
- Use Azure Monitor

## Next Steps

- [Architecture Overview](./architecture.md)
- [Testing Best Practices](./testing.md)
- [Performance Optimization](./performance.md)