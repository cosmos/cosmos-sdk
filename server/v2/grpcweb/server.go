package grpcweb

import (
	"net/http"

	"cosmossdk.io/log"
	"github.com/gorilla/mux"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
)

type Server struct {
	logger  log.Logger
	GRPCSrv *grpc.Server
	unsafe  bool
	cfg     Config
}

// GRPCWebConfig defines configuration for the gRPC-web server.
type Config struct {
	// Enable defines if the gRPC-web should be enabled.
	Enable bool `mapstructure:"enable"`
}

func New(logger log.Logger, grpcSrv *grpc.Server, cfg Config, unsafe bool) *Server {
	return &Server{
		logger:  logger,
		GRPCSrv: grpcSrv,
		unsafe:  unsafe,
		cfg:     cfg,
	}
}

// Register implements registers a grpc-web server
func (s *Server) Register(r mux.Router) error {
	// configure grpc-web server
	if s.cfg.Enable {
		var options []grpcweb.Option
		if s.unsafe {
			options = append(options,
				grpcweb.WithOriginFunc(func(origin string) bool {
					return true
				}),
			)
		}

		wrappedGrpc := grpcweb.WrapServer(s.GRPCSrv, options...)
		r.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			if wrappedGrpc.IsGrpcWebRequest(req) {
				wrappedGrpc.ServeHTTP(w, req)
				return
			}
		}))
	}
	return nil
}
