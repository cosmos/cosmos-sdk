package grpcgateway

import (
	"net/http"
	"strings"

	gateway "github.com/cosmos/gogogateway"
	"github.com/cosmos/gogoproto/jsonpb"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/log"
)

const (
	// GRPCBlockHeightHeader is the gRPC header for block height.
	GRPCBlockHeightHeader = "x-cosmos-block-height"
)

type Server struct {
	logger            log.Logger
	GRPCSrv           *grpc.Server
	GRPCGatewayRouter *runtime.ServeMux
	config            Config
}

// New creates a new gRPC-gateway server.
func New(logger log.Logger, grpcSrv *grpc.Server, cfg Config, ir jsonpb.AnyResolver) *Server {
	// The default JSON marshaller used by the gRPC-Gateway is unable to marshal non-nullable non-scalar fields.
	// Using the gogo/gateway package with the gRPC-Gateway WithMarshaler option fixes the scalar field marshaling issue.
	marshalerOption := &gateway.JSONPb{
		EmitDefaults: true,
		Indent:       "",
		OrigName:     true,
		AnyResolver:  ir,
	}
	return &Server{
		logger: logger,
		GRPCGatewayRouter: runtime.NewServeMux(
			// Custom marshaler option is required for gogo proto
			runtime.WithMarshalerOption(runtime.MIMEWildcard, marshalerOption),

			// This is necessary to get error details properly
			// marshaled in unary requests.
			runtime.WithProtoErrorHandler(runtime.DefaultHTTPProtoErrorHandler),

			// Custom header matcher for mapping request headers to
			// GRPC metadata
			runtime.WithIncomingHeaderMatcher(CustomGRPCHeaderMatcher),
		),
		config: cfg,
	}
}

// CustomGRPCHeaderMatcher for mapping request headers to
// GRPC metadata.
// HTTP headers that start with 'Grpc-Metadata-' are automatically mapped to
// gRPC metadata after removing prefix 'Grpc-Metadata-'. We can use this
// CustomGRPCHeaderMatcher if headers don't start with `Grpc-Metadata-`
func CustomGRPCHeaderMatcher(key string) (string, bool) {
	switch strings.ToLower(key) {
	case GRPCBlockHeightHeader:
		return GRPCBlockHeightHeader, true

	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

// Register implements registers a grpc-gateway server
func (s *Server) Register(r mux.Router) error {
	// configure grpc-gatway server
	if s.config.Enable {
		r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			// Fall back to grpc gateway server.
			s.GRPCGatewayRouter.ServeHTTP(w, req)
		}))
	}

	return nil
}
