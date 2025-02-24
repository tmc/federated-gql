// Code generated by protoc-gen-connect-go. DO NOT EDIT.
//
// Source: products/v1/products.proto

package productv1connect

import (
	connect "connectrpc.com/connect"
	context "context"
	errors "errors"
	v1 "github.com/fraser-isbester/federated-gql/gen/go/product/v1"
	http "net/http"
	strings "strings"
)

// This is a compile-time assertion to ensure that this generated file and the connect package are
// compatible. If you get a compiler error that this constant is not defined, this code was
// generated with a version of connect newer than the one compiled into your binary. You can fix the
// problem by either regenerating this code with an older version of connect or updating the connect
// version compiled into your binary.
const _ = connect.IsAtLeastVersion1_13_0

const (
	// ProductServiceName is the fully-qualified name of the ProductService service.
	ProductServiceName = "product.v1.ProductService"
)

// These constants are the fully-qualified names of the RPCs defined in this package. They're
// exposed at runtime as Spec.Procedure and as the final two segments of the HTTP route.
//
// Note that these are different from the fully-qualified method names used by
// google.golang.org/protobuf/reflect/protoreflect. To convert from these constants to
// reflection-formatted method names, remove the leading slash and convert the remaining slash to a
// period.
const (
	// ProductServiceGetProductProcedure is the fully-qualified name of the ProductService's GetProduct
	// RPC.
	ProductServiceGetProductProcedure = "/product.v1.ProductService/GetProduct"
)

// ProductServiceClient is a client for the product.v1.ProductService service.
type ProductServiceClient interface {
	// GetProduct returns a product by its ID.
	GetProduct(context.Context, *connect.Request[v1.GetProductRequest]) (*connect.Response[v1.GetProductResponse], error)
}

// NewProductServiceClient constructs a client for the product.v1.ProductService service. By
// default, it uses the Connect protocol with the binary Protobuf Codec, asks for gzipped responses,
// and sends uncompressed requests. To use the gRPC or gRPC-Web protocols, supply the
// connect.WithGRPC() or connect.WithGRPCWeb() options.
//
// The URL supplied here should be the base URL for the Connect or gRPC server (for example,
// http://api.acme.com or https://acme.com/grpc).
func NewProductServiceClient(httpClient connect.HTTPClient, baseURL string, opts ...connect.ClientOption) ProductServiceClient {
	baseURL = strings.TrimRight(baseURL, "/")
	productServiceMethods := v1.File_products_v1_products_proto.Services().ByName("ProductService").Methods()
	return &productServiceClient{
		getProduct: connect.NewClient[v1.GetProductRequest, v1.GetProductResponse](
			httpClient,
			baseURL+ProductServiceGetProductProcedure,
			connect.WithSchema(productServiceMethods.ByName("GetProduct")),
			connect.WithClientOptions(opts...),
		),
	}
}

// productServiceClient implements ProductServiceClient.
type productServiceClient struct {
	getProduct *connect.Client[v1.GetProductRequest, v1.GetProductResponse]
}

// GetProduct calls product.v1.ProductService.GetProduct.
func (c *productServiceClient) GetProduct(ctx context.Context, req *connect.Request[v1.GetProductRequest]) (*connect.Response[v1.GetProductResponse], error) {
	return c.getProduct.CallUnary(ctx, req)
}

// ProductServiceHandler is an implementation of the product.v1.ProductService service.
type ProductServiceHandler interface {
	// GetProduct returns a product by its ID.
	GetProduct(context.Context, *connect.Request[v1.GetProductRequest]) (*connect.Response[v1.GetProductResponse], error)
}

// NewProductServiceHandler builds an HTTP handler from the service implementation. It returns the
// path on which to mount the handler and the handler itself.
//
// By default, handlers support the Connect, gRPC, and gRPC-Web protocols with the binary Protobuf
// and JSON codecs. They also support gzip compression.
func NewProductServiceHandler(svc ProductServiceHandler, opts ...connect.HandlerOption) (string, http.Handler) {
	productServiceMethods := v1.File_products_v1_products_proto.Services().ByName("ProductService").Methods()
	productServiceGetProductHandler := connect.NewUnaryHandler(
		ProductServiceGetProductProcedure,
		svc.GetProduct,
		connect.WithSchema(productServiceMethods.ByName("GetProduct")),
		connect.WithHandlerOptions(opts...),
	)
	return "/product.v1.ProductService/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case ProductServiceGetProductProcedure:
			productServiceGetProductHandler.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}

// UnimplementedProductServiceHandler returns CodeUnimplemented from all methods.
type UnimplementedProductServiceHandler struct{}

func (UnimplementedProductServiceHandler) GetProduct(context.Context, *connect.Request[v1.GetProductRequest]) (*connect.Response[v1.GetProductResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, errors.New("product.v1.ProductService.GetProduct is not implemented"))
}
